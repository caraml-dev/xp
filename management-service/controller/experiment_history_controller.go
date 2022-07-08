package controller

import (
	"net/http"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/management-service/api"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
	"github.com/gojek/xp/management-service/services"
)

type ExperimentHistoryController struct {
	*appcontext.AppContext
}

func NewExperimentHistoryController(ctx *appcontext.AppContext) *ExperimentHistoryController {
	return &ExperimentHistoryController{ctx}
}

func (e ExperimentHistoryController) ListExperimentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	experimentId int64,
	params api.ListExperimentHistoryParams,
) {
	err := e.checkProjectAndExperiment(projectId, experimentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// List historical versions
	versions, paging, err := e.Services.ExperimentHistoryService.ListExperimentHistory(experimentId, e.toListExperimentHistoryParams(params))
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	segmenterTypes, err := e.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	versionsResp := []schema.ExperimentHistory{}
	for _, v := range versions {
		versionsResp = append(versionsResp, v.ToApiSchema(segmenterTypes))
	}
	Ok(w, versionsResp, ToPagingSchema(paging))
}

func (e ExperimentHistoryController) GetExperimentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	experimentId int64,
	version int64,
) {
	err := e.checkProjectAndExperiment(projectId, experimentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Get history record
	exp, err := e.Services.ExperimentHistoryService.GetExperimentHistory(experimentId, version)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	segmenterTypes, err := e.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	Ok(w, exp.ToApiSchema(segmenterTypes))
}

func (e ExperimentHistoryController) toListExperimentHistoryParams(params api.ListExperimentHistoryParams) services.ListExperimentHistoryParams {
	return services.ListExperimentHistoryParams{
		PaginationOptions: pagination.PaginationOptions{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
	}
}

func (e ExperimentHistoryController) checkProjectAndExperiment(projectId int64, experimentId int64) error {
	// Check if the projectId is valid
	if _, err := e.Services.MLPService.GetProject(projectId); err != nil {
		return err
	}
	// Check if the projectId has been set up
	_, err := e.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		return errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err)
	}
	// Check that the experiment exists
	_, err = e.Services.ExperimentService.GetDBRecord(models.ID(projectId), models.ID(experimentId))
	if err != nil {
		return errors.Newf(errors.NotFound, "Experiment with id %d cannot be retrieved: %v", experimentId, err)
	}
	return nil
}
