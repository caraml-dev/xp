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
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	_ "go.uber.org/automaxprocs"

	"github.com/gojek/xp/common/web"
	"github.com/gojek/xp/treatment-service/api"
	"github.com/gojek/xp/treatment-service/appcontext"
	"github.com/gojek/xp/treatment-service/config"
	"github.com/gojek/xp/treatment-service/controller"
	"github.com/gojek/xp/treatment-service/middleware"
)

type Server struct {
	*http.Server
	appContext *appcontext.AppContext
	// cleanup captures all the actions to be executed on server shut down
	cleanup []func()
}

// NewServer creates and configures an APIServer serving all application routes.
func NewServer(configFiles []string) (*Server, error) {
	// Collect all the clean up actions
	cleanup := []func(){}

	cfg, err := config.Load(configFiles...)
	if err != nil {
		log.Panicf("Failed initializing config: %v", err)
	}

	// Init NewRelic
	if cfg.NewRelicConfig.Enabled {
		if err := newrelic.InitNewRelic(cfg.NewRelicConfig); err != nil {
			log.Println(fmt.Errorf("failed to initialize newrelic: %s", err))
		}
		cleanup = append(cleanup, func() { newrelic.Shutdown(5 * time.Second) })
	}

	// Init Sentry client
	if cfg.SentryConfig.Enabled {
		cfg.SentryConfig.Labels["environment"] = cfg.DeploymentConfig.EnvironmentType
		if err := sentry.InitSentry(cfg.SentryConfig); err != nil {
			log.Println(fmt.Errorf("failed initializing sentry client: %s", err))
		}
		cleanup = append(cleanup, func() { sentry.Close() })
	}

	// Init AppContext
	appCtx, err := appcontext.NewAppContext(cfg)
	if err != nil {
		log.Panicf("Failed initializing application appcontext: %v", err)
	}

	// Create Chi router and add middlewares
	router := chi.NewRouter()
	if cfg.SwaggerConfig.Enabled {
		router.Use(cors.New(cors.Options{
			AllowCredentials: true,
			AllowedOrigins:   cfg.SwaggerConfig.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
			// Ref: https://swagger.io/docs/open-source-tools/swagger-ui/usage/cors/
			AllowedHeaders: []string{"Authorization", "Content-Type", "api_key"},
		}).Handler)
	}
	// Add NewRelic middleware
	if cfg.NewRelicConfig.Enabled {
		router.Use(middleware.NewRelicMiddleware())
	}

	// Configure controllers
	treatmentController := controller.NewTreatmentController(*appCtx, *cfg)
	apiHandler := api.HandlerFromMux(treatmentController, router)
	if cfg.SentryConfig.Enabled {
		apiHandler = sentry.Recoverer(apiHandler)
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/internal/", http.StripPrefix("/v1/internal", controller.NewInternalController(appCtx, cfg)))
	mux.Handle("/v1/metrics", http.StripPrefix("/v1", promhttp.Handler()))
	mux.Handle("/v1/", http.StripPrefix("/v1", apiHandler))
	// Serve Swagger Specs
	if cfg.SwaggerConfig.Enabled {
		mux.Handle("/treatment.yaml", web.FileHandler(path.Join(cfg.SwaggerConfig.OpenAPISpecsPath, "treatment.yaml"), false))
		mux.Handle("/schema.yaml", web.FileHandler(path.Join(cfg.SwaggerConfig.OpenAPISpecsPath, "schema.yaml"), false))
	}

	srv := http.Server{
		Addr:    cfg.ListenAddress(),
		Handler: mux,
	}

	return &Server{
		Server:     &srv,
		appContext: appCtx,
		cleanup:    cleanup,
	}, nil
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (srv *Server) Start() {
	log.Println("Starting background services...")
	errChannel := make(chan error, 1)
	cancelBackgroundSvc := srv.startBackgroundService(errChannel)
	log.Println("Starting server...")
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			cancelBackgroundSvc()
			panic(err)
		}
	}()
	log.Printf("Listening on %s\n", srv.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case sig := <-stop:
		log.Println("Received signal:", sig)
	case err := <-errChannel:
		log.Println("Background services encounter an error", err.Error())
	}
	log.Println("Shutting down server...")
	cancelBackgroundSvc()

	// Execute clean up actions
	for _, cleanupFunc := range srv.cleanup {
		cleanupFunc()
	}
	err := srv.deleteSubscriptions()
	if err != nil {
		log.Printf("Failed to delete subscriptions when shutting down: %s", err)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	log.Println("Server gracefully stopped")
}

func (srv *Server) deleteSubscriptions() error {
	err := srv.appContext.ExperimentSubscriber.DeleteSubscriptions(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (srv *Server) startBackgroundService(errChannel chan error) context.CancelFunc {
	backgroundSvcCtx, cancel := context.WithCancel(context.Background())
	go func() {
		err := srv.appContext.ExperimentSubscriber.SubscribeToManagementService(backgroundSvcCtx)
		if err != nil {
			errChannel <- err
		}
	}()

	return cancel
}
