package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/xp/common/api/schema"
)

func TestExperimentHistoryToApiSchema(t *testing.T) {
	segmenterTypes := map[string]schema.SegmenterType{
		"string_segmenter":  schema.SegmenterTypeString,
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}

	var testExperimentInterval int32 = 10
	var testExperimentTraffic int32 = 80
	var testDescription = "exp history desc"
	var testTreatmentTraffic20 int32 = 20
	config := map[string]interface{}{
		"config-1": "value",
		"config-2": 2,
	}
	e := ExperimentHistory{
		Model: Model{
			CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
			UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		},
		ID:           ID(100),
		ExperimentID: ID(40),
		Version:      int64(8),
		Name:         "exp-hist-1",
		Description:  &testDescription,
		Type:         ExperimentTypeSwitchback,
		Interval:     &testExperimentInterval,
		Treatments: ExperimentTreatments([]ExperimentTreatment{
			{
				Configuration: config,
				Name:          "control",
				Traffic:       &testExperimentTraffic,
			},
			{
				Configuration: config,
				Name:          "treatment",
				Traffic:       &testTreatmentTraffic20,
			},
		}),
		Segment: ExperimentSegment{
			"string_segmenter":  []string{"value"},
			"integer_segmenter": []string{"1"},
			"float_segmenter":   []string{"1.0"},
			"bool_segmenter":    []string{"true"},
		},
		Status:    ExperimentStatusInactive,
		Tier:      ExperimentTierOverride,
		EndTime:   time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC),
		StartTime: time.Date(2022, 2, 2, 1, 1, 1, 0, time.UTC),
		UpdatedBy: "test-updated-by",
	}
	assert.Equal(t, schema.ExperimentHistory{
		Id:           int64(100),
		CreatedAt:    time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		UpdatedAt:    time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		ExperimentId: int64(40),
		Version:      int64(8),
		Name:         "exp-hist-1",
		Description:  &testDescription,
		Type:         schema.ExperimentTypeSwitchback,
		Interval:     &testExperimentInterval,
		Treatments: []schema.ExperimentTreatment{
			{
				Configuration: map[string]interface{}{
					"config-1": "value",
					"config-2": 2,
				},
				Name:    "control",
				Traffic: &testExperimentTraffic,
			},
			{
				Configuration: map[string]interface{}{
					"config-1": "value",
					"config-2": 2,
				},
				Name:    "treatment",
				Traffic: &testTreatmentTraffic20,
			},
		},
		Segment: schema.ExperimentSegment{
			"string_segmenter":  []string{"value"},
			"integer_segmenter": []int64{1},
			"float_segmenter":   []float64{1.0},
			"bool_segmenter":    []bool{true},
		},
		Status:    schema.ExperimentStatusInactive,
		Tier:      schema.ExperimentTierOverride,
		EndTime:   time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC),
		StartTime: time.Date(2022, 2, 2, 1, 1, 1, 0, time.UTC),
		UpdatedBy: "test-updated-by",
	}, e.ToApiSchema(segmenterTypes))
}
