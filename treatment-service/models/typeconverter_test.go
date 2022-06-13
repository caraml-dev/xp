package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gojek/turing-experiments/common/api/schema"
	"github.com/gojek/turing-experiments/common/pubsub"
	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
	"github.com/gojek/turing-experiments/treatment-service/models"
)

func TestOpenAPIProjectSettingsSpecToProtobuf(t *testing.T) {
	protoSegmenters := &pubsub.Segmenters{
		Names: []string{"string_segmenter", "integer_segmenter"},
		Variables: map[string]*pubsub.ExperimentVariables{
			"string_segmenter":  {Value: []string{"string_segmenter"}},
			"integer_segmenter": {Value: []string{"integer_segmenter"}},
		},
	}

	tests := []struct {
		Name     string
		Settings schema.ProjectSettings
		Expected *pubsub.ProjectSettings
	}{
		{
			Name: "basic",
			Settings: schema.ProjectSettings{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
				ProjectId: 1,
				Username:  "client-1",
				Passkey:   "passkey-1",
				Segmenters: schema.ProjectSegmenters{
					Names: []string{"string_segmenter", "integer_segmenter"},
					Variables: schema.ProjectSegmenters_Variables{
						AdditionalProperties: map[string][]string{
							"string_segmenter":  {"string_segmenter"},
							"integer_segmenter": {"integer_segmenter"},
						},
					},
				},
				RandomizationKey:     "rand-1",
				EnableS2idClustering: true,
			},
			Expected: &pubsub.ProjectSettings{
				ProjectId:            1,
				CreatedAt:            timestamppb.New(time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC)),
				RandomizationKey:     "rand-1",
				Segmenters:           protoSegmenters,
				UpdatedAt:            timestamppb.New(time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC)),
				Username:             "client-1",
				Passkey:              "passkey-1",
				EnableS2IdClustering: true,
			},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.Name, func(t *testing.T) {
			assert.Equal(t, data.Expected, models.OpenAPIProjectSettingsSpecToProtobuf(data.Settings))
		})
	}
}

func TestOpenAPIExperimentSpecToProtobuf(t *testing.T) {
	traffic100 := int32(100)
	interval := int32(60)
	segmentersType := map[string]schema.SegmenterType{
		"string_segmenter": "STRING",
	}
	pubsubCfg, _ := structpb.NewStruct(map[string]interface{}{
		"key": "value",
	})
	tests := []struct {
		Name       string
		Experiment schema.Experiment
		Expected   *pubsub.Experiment
		Error      string
	}{
		{
			Name: "active default a/b experiment",
			Experiment: schema.Experiment{
				ProjectId: 1,
				Id:        2,
				Name:      "experiment-1",
				Segment: schema.ExperimentSegment{
					"string_segmenter": []interface{}{"ID"},
				},
				Status: schema.ExperimentStatusActive,
				Treatments: []schema.ExperimentTreatment{
					{
						Name: "default",
						Configuration: map[string]interface{}{
							"key": "value",
						},
						Traffic: &traffic100,
					},
				},
				Tier:      schema.ExperimentTierDefault,
				Type:      schema.ExperimentTypeAB,
				StartTime: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				EndTime:   time.Date(2022, 1, 1, 2, 3, 4, 0, time.UTC),
				CreatedAt: time.Date(2020, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2020, 2, 1, 2, 3, 4, 0, time.UTC),
			},
			Expected: &pubsub.Experiment{
				ProjectId: 1,
				Id:        2,
				Interval:  0,
				Name:      "experiment-1",
				Segments: map[string]*_segmenters.ListSegmenterValue{
					"string_segmenter": {
						Values: []*_segmenters.SegmenterValue{
							{Value: &_segmenters.SegmenterValue_String_{String_: "ID"}},
						},
					},
				},
				Status: pubsub.Experiment_Active,
				Treatments: []*pubsub.ExperimentTreatment{
					{
						Name:    "default",
						Config:  pubsubCfg,
						Traffic: 100,
					},
				},
				Tier:      pubsub.Experiment_Default,
				Type:      pubsub.Experiment_A_B,
				StartTime: timestamppb.New(time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC)),
				EndTime:   timestamppb.New(time.Date(2022, 1, 1, 2, 3, 4, 0, time.UTC)),
				UpdatedAt: timestamppb.New(time.Date(2020, 2, 1, 2, 3, 4, 0, time.UTC)),
			},
		},
		{
			Name: "inactive override switchback experiment",
			Experiment: schema.Experiment{
				ProjectId: 3,
				Id:        4,
				Interval:  &interval,
				Name:      "experiment-2",
				Status:    schema.ExperimentStatusInactive,
				Tier:      schema.ExperimentTierOverride,
				Type:      schema.ExperimentTypeSwitchback,
				StartTime: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				EndTime:   time.Date(2022, 1, 1, 2, 3, 4, 0, time.UTC),
				CreatedAt: time.Date(2020, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2020, 2, 1, 2, 3, 4, 0, time.UTC),
			},
			Expected: &pubsub.Experiment{
				ProjectId:  3,
				Id:         4,
				Interval:   60,
				Name:       "experiment-2",
				Segments:   map[string]*_segmenters.ListSegmenterValue{},
				Status:     pubsub.Experiment_Inactive,
				Treatments: []*pubsub.ExperimentTreatment{},
				Tier:       pubsub.Experiment_Override,
				Type:       pubsub.Experiment_Switchback,
				StartTime:  timestamppb.New(time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC)),
				EndTime:    timestamppb.New(time.Date(2022, 1, 1, 2, 3, 4, 0, time.UTC)),
				UpdatedAt:  timestamppb.New(time.Date(2020, 2, 1, 2, 3, 4, 0, time.UTC)),
			},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.Name, func(t *testing.T) {
			exp, err := models.OpenAPIExperimentSpecToProtobuf(data.Experiment, segmentersType)
			assert.Equal(t, data.Expected, exp)
			if data.Error != "" {
				assert.EqualError(t, err, data.Error)
			}
		})
	}
}
