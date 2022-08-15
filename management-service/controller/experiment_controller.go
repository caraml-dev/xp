package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-collections/collections/set"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/api"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
	"github.com/caraml-dev/xp/management-service/services"
)

const localEmail = "test@email.com"

const DefaultExperimentTier = models.ExperimentTierDefault

type ExperimentController struct {
	*appcontext.AppContext
	environmentType string
}

func NewExperimentController(ctx *appcontext.AppContext, environmentType string) *ExperimentController {
	return &ExperimentController{ctx, environmentType}
}

func (e ExperimentController) GetExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	// Check if the projectId is valid
	if _, err := e.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	_, err := e.Services.ProjectSettingsService.GetProjectSettings(projectId)
	if err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	exp, err := e.Services.ExperimentService.GetExperiment(projectId, experimentId)
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

func (e ExperimentController) ListExperiments(w http.ResponseWriter, r *http.Request, projectId int64, params api.ListExperimentsParams) {
	// Check if the projectId is valid
	if _, err := e.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	if _, err := e.Services.ProjectSettingsService.GetProjectSettings(projectId); err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	listExperimentParams, err := e.toListExperimentParams(params, projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// List experiments
	exps, paging, err := e.Services.ExperimentService.ListExperiments(projectId, *listExperimentParams)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	segmenterTypes, err := e.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	var expsResp []schema.Experiment
	for _, exp := range exps {
		expsResp = append(expsResp, exp.ToApiSchema(segmenterTypes))
	}

	Ok(w, expsResp, ToPagingSchema(paging))
}

func (e ExperimentController) CreateExperiment(w http.ResponseWriter, r *http.Request, projectId int64) {
	expData := api.CreateExperimentRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&expData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	if expData.UpdatedBy == nil || *expData.UpdatedBy == "" {
		userEmail := r.Header.Get("User-Email")
		if userEmail == "" && e.environmentType == "local" {
			userEmail = localEmail
		}
		if userEmail == "" {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "field (updated_by) cannot be unset"))
			return
		}
		expData.UpdatedBy = &userEmail
	}

	// Check if the projectId is valid
	if _, err := e.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	settings, err := e.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err))
		return
	}
	createExperimentBody, err := e.toCreateExperimentBody(expData)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	exp, err := e.Services.ExperimentService.CreateExperiment(*settings, *createExperimentBody)
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

func (e ExperimentController) UpdateExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	expData := api.UpdateExperimentRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&expData)

	if expData.UpdatedBy == nil || *expData.UpdatedBy == "" {
		userEmail := r.Header.Get("User-Email")
		if userEmail == "" && e.environmentType == "local" {
			userEmail = localEmail
		}
		if userEmail == "" {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "field (updated_by) cannot be unset"))
			return
		}
		expData.UpdatedBy = &userEmail
	}

	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	// Check if the projectId is valid
	if _, err := e.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	settings, err := e.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err))
		return
	}

	updateExperimentBody, err := e.toUpdateExperimentBody(expData)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	exp, err := e.Services.ExperimentService.UpdateExperiment(*settings, experimentId, *updateExperimentBody)
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

func (e ExperimentController) EnableExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	// Check if the projectId is valid
	if _, err := e.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	settings, err := e.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err))
		return
	}

	err = e.Services.ExperimentService.EnableExperiment(*settings, experimentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	Ok(w, nil)
}

func (e ExperimentController) DisableExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	// Check if the projectId is valid
	if _, err := e.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	if _, err := e.Services.ProjectSettingsService.GetProjectSettings(projectId); err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	err := e.Services.ExperimentService.DisableExperiment(projectId, experimentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	Ok(w, nil)
}

func (e ExperimentController) toCreateExperimentBody(body api.CreateExperimentRequestBody) (*services.CreateExperimentRequestBody, error) {
	var treatments []models.ExperimentTreatment
	for _, treatment := range body.Treatments {
		treatments = append(treatments, models.ExperimentTreatment(treatment))
	}

	reqBody := &services.CreateExperimentRequestBody{
		Description: body.Description,
		EndTime:     body.EndTime,
		Interval:    body.Interval,
		Name:        body.Name,
		Segment:     models.ExperimentSegmentRaw(body.Segment),
		StartTime:   body.StartTime,
		Status:      models.ExperimentStatus(body.Status),
		Treatments:  treatments,
		Tier:        DefaultExperimentTier, // Set default
		Type:        models.ExperimentType(body.Type),
		UpdatedBy:   body.UpdatedBy,
	}

	// Replace tier if set in the request body
	if body.Tier != nil {
		reqBody.Tier = models.ExperimentTier(*body.Tier)
	}

	return reqBody, nil
}

func (e ExperimentController) toUpdateExperimentBody(body api.UpdateExperimentRequestBody) (*services.UpdateExperimentRequestBody, error) {
	var treatments []models.ExperimentTreatment
	for _, treatment := range body.Treatments {
		treatments = append(treatments, models.ExperimentTreatment(treatment))
	}

	reqBody := &services.UpdateExperimentRequestBody{
		Description: body.Description,
		EndTime:     body.EndTime,
		Interval:    body.Interval,
		Segment:     models.ExperimentSegmentRaw(body.Segment),
		StartTime:   body.StartTime,
		Status:      models.ExperimentStatus(body.Status),
		Treatments:  treatments,
		Tier:        DefaultExperimentTier, // Set default
		Type:        models.ExperimentType(body.Type),
		UpdatedBy:   body.UpdatedBy,
	}

	// Replace tier if set in the request body
	if body.Tier != nil {
		reqBody.Tier = models.ExperimentTier(*body.Tier)
	}

	return reqBody, nil
}

func (e ExperimentController) toListExperimentParams(params api.ListExperimentsParams, projectId int64) (*services.ListExperimentsParams, error) {
	var status *models.ExperimentStatus
	if params.Status != nil {
		val := models.ExperimentStatus(*params.Status)
		status = &val
	}
	var expTier *models.ExperimentTier
	if params.Tier != nil {
		val := models.ExperimentTier(*params.Tier)
		expTier = &val
	}
	var expType *models.ExperimentType
	if params.Type != nil {
		val := models.ExperimentType(*params.Type)
		expType = &val
	}

	// Retrieve existing segmenters and remove invalid ones from request params
	registeredSegmenters := set.New()
	segmenters, err := e.Services.SegmenterService.ListSegmenters(projectId, services.ListSegmentersParams{})
	if err != nil {
		return nil, err
	}
	for _, segmenter := range segmenters {
		registeredSegmenters.Insert(segmenter.Name)
	}
	validSegmentParam := models.ExperimentSegment{}
	if params.Segment != nil {
		for k, v := range *params.Segment {
			if registeredSegmenters.Has(k) {
				segmenterValue := v.([]string)
				validSegmentParam[k] = segmenterValue
			} else {
				return nil, fmt.Errorf("provided segmenter (%s) is not a registered segmenter", k)
			}
		}
	}

	return &services.ListExperimentsParams{
		PaginationOptions: pagination.PaginationOptions{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
		Status:           status,
		EndTime:          params.EndTime,
		Tier:             expTier,
		Type:             expType,
		Name:             params.Name,
		UpdatedBy:        params.UpdatedBy,
		Search:           params.Search,
		StartTime:        params.StartTime,
		Segment:          validSegmentParam,
		IncludeWeakMatch: params.IncludeWeakMatch != nil && *params.IncludeWeakMatch,
	}, nil
}
