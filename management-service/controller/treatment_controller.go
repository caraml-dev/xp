package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/management-service/api"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
	"github.com/gojek/xp/management-service/services"
)

type TreatmentController struct {
	*appcontext.AppContext
	environmentType string
}

func NewTreatmentController(ctx *appcontext.AppContext, environmentType string) *TreatmentController {
	return &TreatmentController{ctx, environmentType}
}

func (t TreatmentController) GetTreatment(w http.ResponseWriter, r *http.Request, projectId int64, treatmentId int64) {
	// Check if the projectId is valid
	if _, err := t.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	_, err := t.Services.ProjectSettingsService.GetProjectSettings(projectId)
	if err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	treatment, err := t.Services.TreatmentService.GetTreatment(projectId, treatmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, treatment.ToApiSchema())
}

func (t TreatmentController) ListTreatments(w http.ResponseWriter, r *http.Request, projectId int64, params api.ListTreatmentsParams) {
	// Check if the projectId is valid
	if _, err := t.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	if _, err := t.Services.ProjectSettingsService.GetProjectSettings(projectId); err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	// List treatments
	treatments, paging, err := t.Services.TreatmentService.ListTreatments(projectId, t.toListTreatmentParams(params))
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	treatmentsResp := []schema.Treatment{}
	var fields []models.TreatmentField
	if params.Fields != nil {
		for _, field := range *params.Fields {
			fields = append(fields, models.TreatmentField(field))
		}
	}
	for _, t := range treatments {
		treatmentsResp = append(treatmentsResp, t.ToApiSchema(fields...))
	}

	Ok(w, treatmentsResp, ToPagingSchema(paging))
}

func (t TreatmentController) CreateTreatment(w http.ResponseWriter, r *http.Request, projectId int64) {
	treatmentData := api.CreateTreatmentRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&treatmentData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	if treatmentData.UpdatedBy == nil || *treatmentData.UpdatedBy == "" {
		userEmail := r.Header.Get("User-Email")
		if userEmail == "" && t.environmentType == "local" {
			userEmail = localEmail
		}
		if userEmail == "" {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "field (updated_by) cannot be empty"))
			return
		}
		treatmentData.UpdatedBy = &userEmail
	}

	// Check if the projectId is valid
	if _, err := t.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	settings, err := t.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err))
		return
	}
	treatment, err := t.Services.TreatmentService.CreateTreatment(*settings, t.toCreateTreatmentBody(treatmentData))
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	Ok(w, treatment.ToApiSchema())
}

func (t TreatmentController) UpdateTreatment(w http.ResponseWriter, r *http.Request, projectId int64, treatmentId int64) {
	treatmentData := api.UpdateTreatmentRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&treatmentData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	if treatmentData.UpdatedBy == nil || *treatmentData.UpdatedBy == "" {
		userEmail := r.Header.Get("User-Email")
		if userEmail == "" && t.environmentType == "local" {
			userEmail = localEmail
		}
		if userEmail == "" {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "field (updated_by) cannot be empty"))
			return
		}
		treatmentData.UpdatedBy = &userEmail
	}

	// Check if the projectId is valid
	if _, err := t.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	settings, err := t.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err))
		return
	}
	treatment, err := t.Services.TreatmentService.UpdateTreatment(*settings, treatmentId, t.toUpdateTreatmentBody(treatmentData))
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	Ok(w, treatment.ToApiSchema())
}

func (t TreatmentController) DeleteTreatment(w http.ResponseWriter, r *http.Request, projectId int64, treatmentId int64) {
	// Check if the projectId is valid
	if _, err := t.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	_, err := t.Services.ProjectSettingsService.GetProjectSettings(projectId)
	if err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	err = t.Services.TreatmentService.DeleteTreatment(projectId, treatmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	resp := map[string]int64{"id": treatmentId}

	Ok(w, resp)
}

func (t TreatmentController) toCreateTreatmentBody(body api.CreateTreatmentRequestBody) services.CreateTreatmentRequestBody {
	return services.CreateTreatmentRequestBody{
		Name:      body.Name,
		Config:    body.Configuration,
		UpdatedBy: body.UpdatedBy,
	}
}

func (t TreatmentController) toUpdateTreatmentBody(body api.UpdateTreatmentRequestBody) services.UpdateTreatmentRequestBody {
	return services.UpdateTreatmentRequestBody{
		Config:    body.Configuration,
		UpdatedBy: body.UpdatedBy,
	}
}

func (t TreatmentController) toListTreatmentParams(params api.ListTreatmentsParams) services.ListTreatmentsParams {
	finalParams := services.ListTreatmentsParams{
		PaginationOptions: pagination.PaginationOptions{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
		UpdatedBy: params.UpdatedBy,
		Search:    params.Search,
	}
	if params.Fields != nil {
		var fields []models.TreatmentField
		for _, field := range *params.Fields {
			val := models.TreatmentField(field)
			fields = append(fields, val)
		}
		finalParams.Fields = &fields
	}

	return finalParams
}
