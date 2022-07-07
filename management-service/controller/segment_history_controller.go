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

type SegmentHistoryController struct {
	*appcontext.AppContext
}

func NewSegmentHistoryController(ctx *appcontext.AppContext) *SegmentHistoryController {
	return &SegmentHistoryController{ctx}
}

func (s SegmentHistoryController) ListSegmentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	segmentId int64,
	params api.ListSegmentHistoryParams,
) {
	err := s.checkProjectAndSegment(projectId, segmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// List historical versions
	versions, paging, err := s.Services.SegmentHistoryService.ListSegmentHistory(segmentId, s.toListSegmentHistoryParams(params))
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	segmenterTypes, err := s.Services.SegmenterService.GetSegmenterTypes(projectId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	versionsResp := []schema.SegmentHistory{}
	for _, v := range versions {
		versionsResp = append(versionsResp, v.ToApiSchema(segmenterTypes))
	}
	Ok(w, versionsResp, ToPagingSchema(paging))
}

func (s SegmentHistoryController) GetSegmentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	segmentId int64,
	version int64,
) {
	err := s.checkProjectAndSegment(projectId, segmentId)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Get history record
	segment, err := s.Services.SegmentHistoryService.GetSegmentHistory(segmentId, version)
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

func (s SegmentHistoryController) toListSegmentHistoryParams(params api.ListSegmentHistoryParams) services.ListSegmentHistoryParams {
	return services.ListSegmentHistoryParams{
		PaginationOptions: pagination.PaginationOptions{
			Page:     params.Page,
			PageSize: params.PageSize,
		},
	}
}

func (s SegmentHistoryController) checkProjectAndSegment(projectId int64, segmentId int64) error {
	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		return err
	}
	// Check if the projectId has been set up
	_, err := s.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		return errors.Newf(errors.NotFound, "Settings for project_id %d cannot be retrieved: %v", projectId, err)
	}
	// Check that the segment exists
	_, err = s.Services.SegmentService.GetDBRecord(models.ID(projectId), models.ID(segmentId))
	if err != nil {
		return errors.Newf(errors.NotFound, "Segment with id %d cannot be retrieved: %v", segmentId, err)
	}
	return nil
}
