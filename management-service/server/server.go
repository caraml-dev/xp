package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gojek/mlp/api/pkg/authz/enforcer"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/heptiolabs/healthcheck"
	"github.com/rs/cors"

	"github.com/caraml-dev/xp/common/web"
	"github.com/caraml-dev/xp/management-service/api"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/config"
	"github.com/caraml-dev/xp/management-service/controller"
	"github.com/caraml-dev/xp/management-service/database"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/middleware"
)

type Server struct {
	*http.Server
	// cleanup captures all the actions to be executed on server shut down
	cleanup []func()
}

// NewServer creates and configures an APIServer serving all application routes.
func NewServer(configFiles []string) (*Server, error) {
	// Collect all the clean up actions
	cleanup := []func(){}

	// Load config
	cfg, err := config.Load(configFiles...)
	if err != nil {
		return nil, errors.Newf(errors.GetType(err), fmt.Sprintf("Failed loading config files: %v", err))
	}

	// Init DB and run migrations
	db, err := database.Open(cfg.DbConfig)
	if err != nil {
		panic(err)
	}
	err = database.Migrate(cfg.DbConfig)
	if err != nil {
		panic(err)
	}
	db.LogMode(false)
	cleanup = append(cleanup, func() { db.Close() })

	// Init NewRelic
	if cfg.NewRelicConfig.Enabled {
		if err := newrelic.InitNewRelic(cfg.NewRelicConfig); err != nil {
			return nil, errors.Newf(errors.GetType(err), fmt.Sprintf("Failed initializing NewRelic: %v", err))
		}
		cleanup = append(cleanup, func() { newrelic.Shutdown(5 * time.Second) })
	}

	// Init Sentry client
	if cfg.SentryConfig.Enabled {
		cfg.SentryConfig.Labels["environment"] = cfg.DeploymentConfig.EnvironmentType
		if err := sentry.InitSentry(cfg.SentryConfig); err != nil {
			return nil, errors.Newf(errors.GetType(err), fmt.Sprintf("Failed initializing Sentry Client: %v", err))
		}
		cleanup = append(cleanup, func() { sentry.Close() })
	}

	// Init Authorizer
	var authorizer *middleware.Authorizer
	if cfg.AuthorizationConfig.Enabled {
		// Use product mlp as the policies are shared across the mlp products.
		authzEnforcer, err := enforcer.NewEnforcerBuilder().Product("mlp").
			URL(cfg.AuthorizationConfig.URL).Build()
		if err != nil {
			return nil, errors.Newf(errors.GetType(err), fmt.Sprintf("Failed initializing Authorizer: %v", err))
		}
		authorizer, err = middleware.NewAuthorizer(authzEnforcer)
		if err != nil {
			return nil, errors.Newf(errors.GetType(err), fmt.Sprintf("Failed initializing Authorizer: %v", err))
		}
	}

	// Init AppContext
	appCtx, err := appcontext.NewAppContext(db, authorizer, cfg)
	if err != nil {
		return nil, errors.Newf(errors.GetType(err), fmt.Sprintf("Failed initializing AppContext: %v", err))
	}

	// Create Chi router and add middlewares
	router := chi.NewRouter()
	router.Use(appCtx.OpenAPIValidator.Middleware())
	router.Use(cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		// Ref: https://swagger.io/docs/open-source-tools/swagger-ui/usage/cors/
		AllowedHeaders: []string{"Authorization", "Content-Type", "api_key"},
	}).Handler)
	// Add Authorization middleware
	if appCtx.Authorizer != nil {
		router.Use(appCtx.Authorizer.Middleware)
	}
	// Add NewRelic middleware
	if cfg.NewRelicConfig.Enabled {
		router.Use(middleware.NewRelicMiddleware())
	}

	// Register handlers
	apiHandler := api.HandlerFromMux(
		controller.NewWrapper(
			controller.NewProjectSettingsController(appCtx),
			controller.NewExperimentController(appCtx, cfg.DeploymentConfig.EnvironmentType),
			controller.NewExperimentHistoryController(appCtx),
			controller.NewSegmentController(appCtx, cfg.DeploymentConfig.EnvironmentType),
			controller.NewSegmentHistoryController(appCtx),
			controller.NewSegmenterController(appCtx),
			controller.NewTreatmentController(appCtx, cfg.DeploymentConfig.EnvironmentType),
			controller.NewTreatmentHistoryController(appCtx),
			controller.NewValidationController(appCtx),
		),
		router,
	)
	// Add Authorization middleware
	if cfg.SentryConfig.Enabled {
		apiHandler = sentry.Recoverer(apiHandler)
	}
	healthHandler := healthcheck.NewHandler()
	healthHandler.AddReadinessCheck("database", healthcheck.DatabasePingCheck(db.DB(), 1*time.Second))

	mux := http.NewServeMux()
	mux.Handle("/v1/", http.StripPrefix("/v1", apiHandler))
	mux.Handle("/v1/internal/", http.StripPrefix("/v1/internal", healthHandler))
	// Serve Swagger Specs
	mux.Handle("/experiments.yaml", web.FileHandler(path.Join(cfg.OpenAPISpecsPath, "experiments.yaml"), false))
	mux.Handle("/schema.yaml", web.FileHandler(path.Join(cfg.OpenAPISpecsPath, "schema.yaml"), false))
	// Serve UI
	if cfg.XpUIConfig.AppDirectory != "" {
		log.Printf(
			"Serving XP UI from %s at %s",
			cfg.XpUIConfig.AppDirectory,
			cfg.XpUIConfig.Homepage)
		web.ServeReactApp(mux, cfg.XpUIConfig.Homepage, cfg.XpUIConfig.AppDirectory)
	}

	// Init CORS and middleware, wrap mux and configure server
	srv := http.Server{
		Addr:    cfg.ListenAddress(),
		Handler: mux,
	}

	return &Server{&srv, cleanup}, nil
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (srv *Server) Start() {
	log.Println("starting server...")
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Printf("Listening on %s\n", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	log.Println("Shutting down server... Reason:", sig)

	// Execute clean up actions
	for _, cleanupFunc := range srv.cleanup {
		cleanupFunc()
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	log.Println("Server gracefully stopped")
}
