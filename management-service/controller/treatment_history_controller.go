package controller

import (
	"net/http"

	"github.com/gojek/turing-experiments/common/api/schema"
	"github.com/gojek/turing-experiments/management-service/api"
	"github.com/gojek/turing-experiments/management-service/appcontext"
	"github.com/gojek/turing-experiments/management-service/errors"
	"github.com/gojek/turing-experiments/management-service/models"
	"github.com/gojek/turing-experiments/management-service/pagination"
	"github.com/gojek/turing-experiments/management-service/services"
)

type TreatmentHistoryController struct {
	*appcontext.AppContext
}

func NewTreatmentHistoryController(ctx *appcontext.AppContext) *TreatmentHistoryController {
	return &TreatmentHistoryController{ctx}
}

func (t TreatmentHistoryController) ListTreatmentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	treatmentId int64,
	params api.ListTreatmentHistoryParams,
) {
	err := t.checkProjectAndTreatment(projectId, treatmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// List historical versions
	versions, paging, err := t.Services.TreatmentHistoryService.ListTreatmentHistory(treatmentId, t.toListTreatmentHistoryParams(params))
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	versionsResp := []schema.TreatmentHistory{}
	for _, v := range versions {
		versionsResp = append(versionsResp, v.ToApiSchema())
	}
	Ok(w, versionsResp, ToPagingSchema(paging))
}

func (t TreatmentHistoryController) GetTreatmentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	treatmentId int64,
	version int64,
) {
	err := t.checkProjectAndTreatment(projectId, treatmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Get history record
	treatment, err := t.Services.TreatmentHistoryService.GetTreatmentHistory(treatmentId, version)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, treatment.ToApiSchema())
}

func (t TreatmentHistoryController) toListTreatmentHistoryParams(params api.ListTreatmentHistoryParams) services.ListTreatmentHistoryParams {
	return services.ListTreatmentHistoryParams{
		PaginationOptions: pagination.PaginationOptions{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
	}
}

func (t TreatmentHistoryController) checkProjectAndTreatment(projectId int64, treatmentId int64) error {
	// Check if the projectId is valid
	if _, err := t.Services.MLPService.GetProject(projectId); err != nil {
		return err
	}
	// Check if the projectId has been set up
	_, err := t.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		return errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err)
	}
	// Check that the treatment exists
	_, err = t.Services.TreatmentService.GetDBRecord(models.ID(projectId), models.ID(treatmentId))
	if err != nil {
		return errors.Newf(errors.NotFound, "Treatment with id %d cannot be retrieved: %v", treatmentId, err)
	}
	return nil
}
