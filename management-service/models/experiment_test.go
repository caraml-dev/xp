package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	_utils "github.com/caraml-dev/xp/common/utils"
)

var testExperimentInterval int32 = 100
var testExperimentTraffic int32 = 20
var testExperimentDescription = "desc"
var testExperiment = Experiment{
	Model: Model{
		CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		UpdatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
	},
	ID:          ID(5),
	ProjectID:   ID(1),
	UpdatedBy:   "admin",
	EndTime:     time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC),
	StartTime:   time.Date(2022, 2, 2, 1, 1, 1, 1, time.UTC),
	Interval:    &testExperimentInterval,
	Name:        "test-exp",
	Description: &testExperimentDescription,
	Segment: ExperimentSegment{
		"string_segmenter": []string{"seg-1"},
	},
	Status: ExperimentStatusActive,
	Treatments: ExperimentTreatments([]ExperimentTreatment{
		{
			Configuration: map[string]interface{}{
				"config-1": "value",
				"config-2": 2,
			},
			Name:    "control",
			Traffic: &testExperimentTraffic,
		},
	}),
	Type: ExperimentTypeSwitchback,
	Tier: ExperimentTierDefault,
}

func TestExperimentToApiSchema(t *testing.T) {
	segmenterTypes := map[string]schema.SegmenterType{
		"string_segmenter": schema.SegmenterTypeString,
	}

	assert.Equal(t, schema.Experiment{
		Id:          int64(5),
		ProjectId:   int64(1),
		CreatedAt:   time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		UpdatedAt:   time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
		UpdatedBy:   "admin",
		EndTime:     time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC),
		StartTime:   time.Date(2022, 2, 2, 1, 1, 1, 1, time.UTC),
		Name:        "test-exp",
		Description: &testExperimentDescription,
		Interval:    &testExperimentInterval,
		Status:      schema.ExperimentStatusActive,
		Type:        schema.ExperimentTypeSwitchback,
		Tier:        schema.ExperimentTierDefault,
		Treatments: []schema.ExperimentTreatment{
			{
				Configuration: map[string]interface{}{
					"config-1": "value",
					"config-2": 2,
				},
				Name:    "control",
				Traffic: &testExperimentTraffic,
			},
		},
		Segment: schema.ExperimentSegment{
			"string_segmenter": []string{"seg-1"},
		},
	}, testExperiment.ToApiSchema(segmenterTypes))
}

func TestExperimentToProtoSchema(t *testing.T) {
	treatmentConfig, err := testExperiment.Treatments.ToProtoSchema()
	require.NoError(t, err)

	segmentersType := map[string]schema.SegmenterType{
		"string_segmenter": schema.SegmenterTypeString,
	}

	stringSegment := []string{"seg-1"}
	protoRecord, err := testExperiment.ToProtoSchema(segmentersType)
	require.NoError(t, err)
	assert.Equal(t, &_pubsub.Experiment{
		Id:         int64(5),
		ProjectId:  int64(1),
		UpdatedAt:  timestamppb.New(time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC)),
		EndTime:    timestamppb.New(time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC)),
		StartTime:  timestamppb.New(time.Date(2022, 2, 2, 1, 1, 1, 1, time.UTC)),
		Name:       "test-exp",
		Interval:   testExperimentInterval,
		Status:     _pubsub.Experiment_Active,
		Type:       _pubsub.Experiment_Switchback,
		Treatments: treatmentConfig,
		Segments: map[string]*_segmenters.ListSegmenterValue{
			"string_segmenter": _utils.StringSliceToListSegmenterValue(&stringSegment),
		},
		Tier: _pubsub.Experiment_Default,
	}, protoRecord)
}
