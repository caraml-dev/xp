package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/management-service/api"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
)

type ProjectSettingsController struct {
	*appcontext.AppContext
}

func NewProjectSettingsController(ctx *appcontext.AppContext) *ProjectSettingsController {
	return &ProjectSettingsController{ctx}
}

func (p ProjectSettingsController) GetProjectSettings(w http.ResponseWriter, r *http.Request, projectId int64) {
	// Check if the projectId is valid
	if _, err := p.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}

	settings, err := p.Services.ProjectSettingsService.GetProjectSettings(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	resp := settings.ToApiSchema()

	Ok(w, resp)
}

func (p ProjectSettingsController) GetProjectExperimentVariables(w http.ResponseWriter, r *http.Request, projectId int64) {
	// Check if the projectId is valid
	if _, err := p.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}

	parameters, err := p.Services.ProjectSettingsService.GetExperimentVariables(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, *parameters)
}

func (p ProjectSettingsController) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := p.Services.ProjectSettingsService.ListProjects()
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Filter only the valid MLP Projects
	filteredProjects := []schema.Project{}
	if projects != nil {
		for _, project := range *projects {
			if _, err := p.Services.MLPService.GetProject(project.Id); err == nil {
				filteredProjects = append(filteredProjects, project.ToApiSchema())
			}
		}
	}

	Ok(w, filteredProjects)
}

func (p ProjectSettingsController) CreateProjectSettings(w http.ResponseWriter, r *http.Request, projectId int64) {
	settingsData := api.CreateProjectSettingsRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&settingsData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	// Check if the projectId is valid
	project, err := p.Services.MLPService.GetProject(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	settings, err := p.Services.ProjectSettingsService.CreateProjectSettings(
		projectId,
		services.CreateProjectSettingsRequestBody{
			Segmenters: models.ProjectSegmenters{
				Names:     settingsData.Segmenters.Names,
				Variables: settingsData.Segmenters.Variables.AdditionalProperties,
			},
			TreatmentSchema:      parseTreatmentSchema(settingsData.TreatmentSchema),
			ValidationUrl:        settingsData.ValidationUrl,
			RandomizationKey:     settingsData.RandomizationKey,
			Username:             project.Name,
			EnableS2idClustering: settingsData.EnableS2idClustering,
		},
	)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	resp := settings.ToApiSchema()

	Ok(w, resp)
}

func (p ProjectSettingsController) UpdateProjectSettings(w http.ResponseWriter, r *http.Request, projectId int64) {
	settingsData := api.UpdateProjectSettingsRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&settingsData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	// Check if the projectId is valid
	if _, err := p.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check that the settings exists
	if _, err := p.Services.ProjectSettingsService.GetProjectSettings(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}

	settings, err := p.Services.ProjectSettingsService.UpdateProjectSettings(
		projectId,
		services.UpdateProjectSettingsRequestBody{
			Segmenters: models.ProjectSegmenters{
				Names:     settingsData.Segmenters.Names,
				Variables: settingsData.Segmenters.Variables.AdditionalProperties,
			},
			TreatmentSchema:      parseTreatmentSchema(settingsData.TreatmentSchema),
			ValidationUrl:        settingsData.ValidationUrl,
			RandomizationKey:     settingsData.RandomizationKey,
			EnableS2idClustering: settingsData.EnableS2idClustering,
		},
	)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	resp := settings.ToApiSchema()

	Ok(w, resp)
}

// parseTreatmentSchema parses treatmentSchema from an api struct into a model struct
func parseTreatmentSchema(treatmentSchema *schema.TreatmentSchema) (parsedTreatmentSchema *models.TreatmentSchema) {
	if treatmentSchema == nil {
		return
	}

	parsedTreatmentSchema = &models.TreatmentSchema{Rules: make([]models.Rule, 0)}
	for _, rule := range treatmentSchema.Rules {
		parsedTreatmentSchema.Rules = append(parsedTreatmentSchema.Rules, models.Rule{Name: rule.Name,
			Predicate: rule.Predicate})
	}

	return
}
