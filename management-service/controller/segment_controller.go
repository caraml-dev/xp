package controller

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/api"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
	"github.com/caraml-dev/xp/management-service/services"
)

type SegmentController struct {
	*appcontext.AppContext
	environmentType string
}

func NewSegmentController(ctx *appcontext.AppContext, environmentType string) *SegmentController {
	return &SegmentController{ctx, environmentType}
}

func (s SegmentController) ListSegments(w http.ResponseWriter, r *http.Request, projectId int64, params api.ListSegmentsParams) {
	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	if _, err := s.Services.ProjectSettingsService.GetProjectSettings(projectId); err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	// List segments
	segments, paging, err := s.Services.SegmentService.ListSegments(projectId, s.toListSegmentParams(params))
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	segmentsResp := []schema.Segment{}
	var fields []models.SegmentField
	if params.Fields != nil {
		for _, field := range *params.Fields {
			fields = append(fields, models.SegmentField(field))
		}
	}
	segmenterTypes, err := s.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	for _, segment := range segments {
		segmentsResp = append(segmentsResp, segment.ToApiSchema(segmenterTypes, fields...))
	}

	Ok(w, segmentsResp, ToPagingSchema(paging))
}

func (s SegmentController) GetSegment(w http.ResponseWriter, r *http.Request, projectId int64, segmentId int64) {
	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	_, err := s.Services.ProjectSettingsService.GetProjectSettings(projectId)
	if err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	segment, err := s.Services.SegmentService.GetSegment(projectId, segmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	segmenterTypes, err := s.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, segment.ToApiSchema(segmenterTypes))
}

func (s SegmentController) CreateSegment(w http.ResponseWriter, r *http.Request, projectId int64) {
	segmentData := api.CreateSegmentRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&segmentData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	if segmentData.UpdatedBy == nil || *segmentData.UpdatedBy == "" {
		userEmail := r.Header.Get("User-Email")
		if userEmail == "" && s.environmentType == "local" {
			userEmail = localEmail
		}
		if userEmail == "" {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "field (updated_by) cannot be empty"))
			return
		}
		segmentData.UpdatedBy = &userEmail
	}

	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	settings, err := s.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err))
		return
	}
	createSegmentBody, err := s.toCreateSegmentBody(segmentData)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	segment, err := s.Services.SegmentService.CreateSegment(*settings, *createSegmentBody)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	segmenterTypes, err := s.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	Ok(w, segment.ToApiSchema(segmenterTypes))
}

func (s SegmentController) UpdateSegment(w http.ResponseWriter, r *http.Request, projectId int64, segmentId int64) {
	segmentData := api.UpdateSegmentRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&segmentData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	if segmentData.UpdatedBy == nil || *segmentData.UpdatedBy == "" {
		userEmail := r.Header.Get("User-Email")
		if userEmail == "" && s.environmentType == "local" {
			userEmail = localEmail
		}
		if userEmail == "" {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "field (updated_by) cannot be empty"))
			return
		}
		segmentData.UpdatedBy = &userEmail
	}

	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	settings, err := s.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err))
		return
	}
	updateSegmentBody, err := s.toUpdateSegmentBody(segmentData)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	segment, err := s.Services.SegmentService.UpdateSegment(*settings, segmentId, *updateSegmentBody)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	segmenterTypes, err := s.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	Ok(w, segment.ToApiSchema(segmenterTypes))
}

func (s SegmentController) DeleteSegment(w http.ResponseWriter, r *http.Request, projectId int64, segmentId int64) {
	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check if the projectId has been set up
	_, err := s.Services.ProjectSettingsService.GetProjectSettings(projectId)
	if err != nil {
		WriteErrorResponse(w, errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId))
		return
	}

	err = s.Services.SegmentService.DeleteSegment(projectId, segmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	resp := map[string]int64{"id": segmentId}

	Ok(w, resp)
}

func (s SegmentController) toCreateSegmentBody(body api.CreateSegmentRequestBody) (*services.CreateSegmentRequestBody, error) {
	return &services.CreateSegmentRequestBody{
		Name:      body.Name,
		Segment:   models.ExperimentSegmentRaw(body.Segment),
		UpdatedBy: body.UpdatedBy,
	}, nil
}

func (s SegmentController) toUpdateSegmentBody(body api.UpdateSegmentRequestBody) (*services.UpdateSegmentRequestBody, error) {
	return &services.UpdateSegmentRequestBody{
		Segment:   models.ExperimentSegmentRaw(body.Segment),
		UpdatedBy: body.UpdatedBy,
	}, nil
}

func (s SegmentController) toListSegmentParams(params api.ListSegmentsParams) services.ListSegmentsParams {
	finalParams := services.ListSegmentsParams{
		PaginationOptions: pagination.PaginationOptions{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
		UpdatedBy: params.UpdatedBy,
		Search:    params.Search,
	}

	if params.Fields != nil {
		var fields []models.SegmentField
		for _, field := range *params.Fields {
			val := models.SegmentField(field)
			fields = append(fields, val)
		}
		finalParams.Fields = &fields
	}

	return finalParams
}
