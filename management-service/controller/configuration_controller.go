package controller

import (
	"net/http"

	"github.com/caraml-dev/xp/management-service/appcontext"
)

type ConfigurationController struct {
	*appcontext.AppContext
}

func NewConfigurationController(ctx *appcontext.AppContext) *ConfigurationController {
	return &ConfigurationController{ctx}
}

func (e ConfigurationController) GetTreatmentServiceConfig(w http.ResponseWriter, r *http.Request) {
	treatmentServiceConfig := e.Services.ConfigurationService.GetTreatmentServiceConfig()
	Ok(w, treatmentServiceConfig)
}
