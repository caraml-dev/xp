package controller

import (
	"net/http"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/management-service/appcontext"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/segmenters"
)

type SegmenterController struct {
	*appcontext.AppContext
}

func NewSegmenterController(ctx *appcontext.AppContext) *SegmenterController {
	return &SegmenterController{ctx}
}

func (s SegmenterController) ListSegmenters(w http.ResponseWriter, r *http.Request) {
	// Retrieve all segmenters' name
	segmenterNames := s.Services.SegmenterService.ListSegmenterNames()

	// Return the segmenter configurations
	segmenterConfigurations, err := s.Services.SegmenterService.GetSegmenterConfigurations(segmenterNames)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	var resp []*schema.Segmenter
	for _, segmenterConfiguration := range segmenterConfigurations {
		config, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(segmenterConfiguration)
		if err != nil {
			WriteErrorResponse(w, err)
			return
		}
		resp = append(resp, config)
	}

	Ok(w, &resp)
}

func (s SegmenterController) GetSegmenters(w http.ResponseWriter, r *http.Request, projectId int64) {
	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Retrieve required segmenters' name
	dbRecord, err := s.Services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.NotFound, err.Error()))
		return
	}
	settings := dbRecord.ToApiSchema()

	segmenterConfigurations, err := s.Services.SegmenterService.GetSegmenterConfigurations(settings.Segmenters.Names)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	resp := []*schema.Segmenter{}
	for _, segmenterConfiguration := range segmenterConfigurations {
		config, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(segmenterConfiguration)
		if err != nil {
			WriteErrorResponse(w, err)
			return
		}
		resp = append(resp, config)
	}

	Ok(w, &resp)
}
