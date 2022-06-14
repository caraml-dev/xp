package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/xp/common/api/schema"
)

func TestSegmentHistoryToApiSchema(t *testing.T) {
	segmenterTypes := map[string]schema.SegmenterType{
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"string_segmenter":  schema.SegmenterTypeString,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}

	name := "segment-hist"
	updatedBy := "test-updated-by"
	stringSegmenter := []string{"seg-1", "seg-2"}
	integerSegmenter := []string{"1", "2"}
	floatSegmenter := []string{"1.0", "2.0"}
	boolSegmenter := []string{"true"}
	experimentSegment := ExperimentSegment{
		"integer_segmenter": integerSegmenter,
		"float_segmenter":   floatSegmenter,
		"string_segmenter":  stringSegmenter,
		"bool_segmenter":    boolSegmenter,
	}

	segment := SegmentHistory{
		Model: Model{
			CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
			UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		},
		ID:        ID(100),
		SegmentID: ID(40),
		Version:   int64(8),
		Name:      name,
		Segment:   experimentSegment,
		UpdatedBy: updatedBy,
	}

	assert.Equal(t, schema.SegmentHistory{
		Id:        int64(100),
		CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		SegmentId: int64(40),
		Version:   int64(8),
		Name:      name,
		Segment:   experimentSegment.ToApiSchema(segmenterTypes),
		UpdatedBy: updatedBy,
	}, segment.ToApiSchema(segmenterTypes))
}
