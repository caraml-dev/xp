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

func (e ConfigurationController) GetTreatmentServicePluginConfig(w http.ResponseWriter, r *http.Request) {
	treatmentServicePluginConfig := e.Services.ConfigurationService.GetTreatmentServicePluginConfig()
	Ok(w, treatmentServicePluginConfig)
}
