package controller

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/heptiolabs/healthcheck"

	"github.com/gojek/xp/treatment-service/appcontext"
	"github.com/gojek/xp/treatment-service/config"
)

type InternalController struct {
	http.Handler
	*appcontext.AppContext
	Config *config.Config
}

func NewInternalController(ctx *appcontext.AppContext, cfg *config.Config) *InternalController {
	healthCheckHandler := healthcheck.NewHandler()
	healthCheckHandler.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))

	mux := http.NewServeMux()
	mux.Handle("/health/", http.StripPrefix("/health", healthCheckHandler))
	mux.Handle("/debug/dump", NewCacheDumpHandler(ctx, cfg))
	// For profiling. net/http/pprof will register itself to http.DefaultServeMux.
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
	return &InternalController{Handler: mux, AppContext: ctx, Config: cfg}
}

type debugHandler struct {
	*appcontext.AppContext
	Config *config.Config
}

func NewCacheDumpHandler(ctx *appcontext.AppContext, cfg *config.Config) http.Handler {
	return &debugHandler{AppContext: ctx, Config: cfg}
}

func (h *debugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filepath, err := h.ExperimentService.DumpExperiments(h.Config.DebugConfig.OutputPath)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err, nil)
		return
	}
	Ok(w, map[string]interface{}{"filepath": filepath}, nil)
}
