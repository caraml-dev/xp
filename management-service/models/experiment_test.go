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
	Type:    ExperimentTypeSwitchback,
	Tier:    ExperimentTierDefault,
	Version: 2,
}

func TestExperimentToApiSchema(t *testing.T) {
	segmenterTypes := map[string]schema.SegmenterType{
		"string_segmenter": schema.SegmenterTypeString,
	}

	id := int64(5)
	projectId := int64(1)
	createdAt := time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC)
	updatedAt := time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC)
	updatedBy := "admin"
	endTime := time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC)
	startTime := time.Date(2022, 2, 2, 1, 1, 1, 1, time.UTC)
	name := "test-exp"
	status := schema.ExperimentStatusActive
	statusFriendly := schema.ExperimentStatusFriendlyCompleted
	experimentType := schema.ExperimentTypeSwitchback
	tier := schema.ExperimentTierDefault
	version := int64(2)

	assert.Equal(t, schema.Experiment{
		Id:             &id,
		ProjectId:      &projectId,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
		UpdatedBy:      &updatedBy,
		EndTime:        &endTime,
		StartTime:      &startTime,
		Name:           &name,
		Description:    &testExperimentDescription,
		Interval:       &testExperimentInterval,
		Status:         &status,
		StatusFriendly: &statusFriendly,
		Type:           &experimentType,
		Tier:           &tier,
		Treatments: &[]schema.ExperimentTreatment{
			{
				Configuration: map[string]interface{}{
					"config-1": "value",
					"config-2": 2,
				},
				Name:    "control",
				Traffic: &testExperimentTraffic,
			},
		},
		Segment: &schema.ExperimentSegment{
			"string_segmenter": []string{"seg-1"},
		},
		Version: &version,
	}, testExperiment.ToApiSchema(segmenterTypes))
}

func TestExperimentToApiSchemaWithFields(t *testing.T) {
	segmenterTypes := map[string]schema.SegmenterType{
		"string_segmenter": schema.SegmenterTypeString,
	}

	fields := []ExperimentField{
		ExperimentFieldId,
		ExperimentFieldName,
		ExperimentFieldType,
		ExperimentFieldStatusFriendly,
		ExperimentFieldTier,
		ExperimentFieldStartTime,
		ExperimentFieldEndTime,
		ExperimentFieldUpdatedAt,
		ExperimentFieldTreatments,
	}
	id := int64(5)
	updatedAt := time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC)
	endTime := time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC)
	startTime := time.Date(2022, 2, 2, 1, 1, 1, 1, time.UTC)
	name := "test-exp"
	statusFriendly := schema.ExperimentStatusFriendlyCompleted
	experimentType := schema.ExperimentTypeSwitchback
	tier := schema.ExperimentTierDefault

	assert.Equal(t, schema.Experiment{
		Id:             &id,
		UpdatedAt:      &updatedAt,
		EndTime:        &endTime,
		StartTime:      &startTime,
		Name:           &name,
		StatusFriendly: &statusFriendly,
		Type:           &experimentType,
		Tier:           &tier,
		Treatments: &[]schema.ExperimentTreatment{
			{
				Configuration: map[string]interface{}{
					"config-1": "value",
					"config-2": 2,
				},
				Name:    "control",
				Traffic: &testExperimentTraffic,
			},
		},
	}, testExperiment.ToApiSchema(segmenterTypes, fields...))
}

func TestExperimentToApiSchemaStatusFriendly(t *testing.T) {
	tests := map[string]struct {
		startTime time.Time
		endTime   time.Time
		status    ExperimentStatus
		expected  schema.ExperimentStatusFriendly
	}{
		"deactivated": {
			startTime: time.Date(2000, 1, 1, 2, 3, 4, 0, time.UTC),
			endTime:   time.Date(3000, 1, 1, 2, 3, 4, 0, time.UTC),
			status:    ExperimentStatusInactive,
			expected:  schema.ExperimentStatusFriendlyDeactivated,
		},
		"running": {
			startTime: time.Date(2000, 1, 1, 2, 3, 4, 0, time.UTC),
			endTime:   time.Date(3000, 1, 1, 2, 3, 4, 0, time.UTC),
			status:    ExperimentStatusActive,
			expected:  schema.ExperimentStatusFriendlyRunning,
		},
		"scheduled": {
			startTime: time.Date(3000, 1, 1, 2, 3, 4, 0, time.UTC),
			endTime:   time.Date(3000, 1, 1, 2, 3, 4, 0, time.UTC),
			status:    ExperimentStatusActive,
			expected:  schema.ExperimentStatusFriendlyScheduled,
		},
		"completed": {
			startTime: time.Date(2000, 1, 1, 2, 3, 4, 0, time.UTC),
			endTime:   time.Date(2000, 1, 1, 2, 3, 4, 0, time.UTC),
			status:    ExperimentStatusActive,
			expected:  schema.ExperimentStatusFriendlyCompleted,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getExperimentStatusFriendly(tt.startTime, tt.endTime, tt.status))
		})
	}
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
		Tier:    _pubsub.Experiment_Default,
		Version: 2,
	}, protoRecord)
}
