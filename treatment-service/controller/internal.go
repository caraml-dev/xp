package controller

import (
	"net/http"

	"github.com/heptiolabs/healthcheck"

	"github.com/gojek/turing-experiments/treatment-service/appcontext"
	"github.com/gojek/turing-experiments/treatment-service/config"
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
	mux.Handle("/debug", NewDebugHandler(ctx, cfg))
	return &InternalController{Handler: mux, AppContext: ctx, Config: cfg}
}

type debugHandler struct {
	*appcontext.AppContext
	Config *config.Config
}

func NewDebugHandler(ctx *appcontext.AppContext, cfg *config.Config) http.Handler {
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
