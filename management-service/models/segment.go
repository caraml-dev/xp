package models

import (
	"github.com/gojek/turing-experiments/common/api/schema"
)

type SegmentField string

// Defines values for SegmentField.
const (
	SegmentFieldId SegmentField = "id"

	SegmentFieldName SegmentField = "name"
)

type Segment struct {
	Model

	// ID is the id of the Segment
	ID ID `json:"id" gorm:"primary_key"`

	// ProjectID is the id of the project that this client belongs to,
	// as retrieved from the MLP API.
	ProjectID ID `json:"project_id"`

	// Name is the segment's name
	Name string `json:"name"`

	Segment ExperimentSegment `json:"segment"`

	// UpdatedBy holds the details of the last person/job that updated the experiment
	UpdatedBy string `json:"updated_by"`
}

// ToApiSchema converts the configured segment DB model to a format compatible with the
// OpenAPI specifications.
func (s *Segment) ToApiSchema(segmentersType map[string]schema.SegmenterType, fields ...SegmentField) schema.Segment {
	segment := schema.Segment{}

	// Only return requested fields
	if fields != nil {
		for _, field := range fields {
			switch field {
			case "name":
				segment.Name = &s.Name
			case "id":
				id := s.ID.ToApiSchema()
				segment.Id = &id
			}
		}
		return segment
	}

	id := s.ID.ToApiSchema()
	projectId := s.ProjectID.ToApiSchema()
	segmentConfig := s.Segment.ToApiSchema(segmentersType)

	return schema.Segment{
		Name:      &s.Name,
		CreatedAt: &s.CreatedAt,
		UpdatedAt: &s.UpdatedAt,
		UpdatedBy: &s.UpdatedBy,
		Id:        &id,
		Segment:   &segmentConfig,
		ProjectId: &projectId,
	}
}
