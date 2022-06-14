package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/management-service/models"
)

func TestSegmentToApiSchema(t *testing.T) {
	segmenterTypes := map[string]schema.SegmenterType{
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"string_segmenter":  schema.SegmenterTypeString,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}
	stringSegmenter := []string{"seg-1", "seg-2"}
	integerSegmenter := []string{"1", "2"}
	floatSegmenter := []string{"1.0", "2.0"}
	boolSegmenter := []string{"false"}

	segmentModel := models.Segment{
		ID: models.ID(1),
		Model: models.Model{
			CreatedAt: time.Date(2021, 1, 1, 3, 4, 5, 0, time.UTC),
			UpdatedAt: time.Date(2021, 12, 1, 3, 4, 5, 0, time.UTC),
		},
		ProjectID: models.ID(10),
		Name:      "segment-1",
		Segment: models.ExperimentSegment{
			"integer_segmenter": integerSegmenter,
			"float_segmenter":   floatSegmenter,
			"string_segmenter":  stringSegmenter,
			"bool_segmenter":    boolSegmenter,
		},
		UpdatedBy: "user-1",
	}

	id := segmentModel.ID.ToApiSchema()
	projectId := segmentModel.ProjectID.ToApiSchema()
	segmentConfig := segmentModel.Segment.ToApiSchema(segmenterTypes)

	// Test ToApiSchema without fields
	assert.Equal(t, schema.Segment{
		CreatedAt: &segmentModel.CreatedAt,
		Id:        &id,
		Name:      &segmentModel.Name,
		ProjectId: &projectId,
		Segment:   &segmentConfig,
		UpdatedAt: &segmentModel.UpdatedAt,
		UpdatedBy: &segmentModel.UpdatedBy,
	}, segmentModel.ToApiSchema(segmenterTypes))

	// Test with fields
	assert.Equal(t, schema.Segment{
		Id:   &id,
		Name: &segmentModel.Name,
	}, segmentModel.ToApiSchema(segmenterTypes, []models.SegmentField{"id", "name"}...))
}
