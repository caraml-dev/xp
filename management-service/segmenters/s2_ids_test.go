package segmenters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	_segmenters "github.com/gojek/xp/common/segmenters"
)

func TestS2IdsValidateSegmenterAndConstraints(t *testing.T) {
	// Init s2ids segmenter with allowed test levels 14 and 15
	jsonConfig := `{"mins2celllevel": 10, "maxs2celllevel": 14}`

	s2idsSegmenter, _ := NewS2IDSegmenter(json.RawMessage(jsonConfig))
	tests := map[string]struct {
		segmenter Segmenter
		values    map[string]*_segmenters.ListSegmenterValue
		errString string
	}{
		"success | empty map": {
			values: map[string]*_segmenters.ListSegmenterValue{},
		},
		"success | empty list": {
			values: map[string]*_segmenters.ListSegmenterValue{
				"s2_ids": {
					Values: []*_segmenters.SegmenterValue{},
				},
			},
		},
		"failure | invalid value type": {
			values: map[string]*_segmenters.ListSegmenterValue{
				"s2_ids": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_String_{String_: "2"}},
					},
				},
			},
			errString: "Segmenter s2_ids has one or more values that do not match the configured type",
		},
		"failure | invalid value": {
			values: map[string]*_segmenters.ListSegmenterValue{
				"s2_ids": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 0}},
					},
				},
			},
			errString: "One or more s2_ids values is invalid",
		},
		"failure | invalid s2id level": {
			values: map[string]*_segmenters.ListSegmenterValue{
				"s2_ids": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592201244967436288}}, // Level 8 (invalid)
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592195425286750208}}, // Level 14 (valid)
					},
				},
			},
			errString: "One or more s2_ids values is at level 8, only the following levels are allowed: [10 11 12 13 14]",
		},
		"success | valid values": {
			values: map[string]*_segmenters.ListSegmenterValue{
				"s2_ids": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592184752293019648}}, // Level 10
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592184408695635968}}, // Level 12
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 3592195425286750208}}, // Level 14
					},
				},
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			err := s2idsSegmenter.ValidateSegmenterAndConstraints(data.values)
			if data.errString == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}

func TestNewS2IDSegmenter(t *testing.T) {
	tests := []struct {
		name       string
		configData json.RawMessage
		want       *s2ids
		wantErr    string
	}{
		{
			name:       "ok",
			configData: json.RawMessage(`{"mins2celllevel": 10, "maxs2celllevel": 12}`),
			want: &s2ids{
				Segmenter: NewBaseSegmenter(&_segmenters.SegmenterConfiguration{
					Name:        "s2_ids",
					Type:        _segmenters.SegmenterValueType_INTEGER,
					Options:     map[string]*_segmenters.SegmenterValue{},
					MultiValued: true,
					TreatmentRequestFields: &_segmenters.ListExperimentVariables{
						Values: []*_segmenters.ExperimentVariables{
							{
								Value: []string{"s2id"},
							},
							{
								Value: []string{"latitude", "longitude"},
							},
						},
					},
					Required:    false,
					Description: fmt.Sprintf("S2 Cell IDs between levels %d and %d are supported.", 10, 12),
				}),
				AllowedLevels: []int{10, 11, 12},
			},
		},
		{
			name:       "invalid range, min larger than max",
			configData: json.RawMessage(`{"mins2celllevel": 13, "maxs2celllevel": 12}`),
			wantErr:    "failed to create segmenter (s2_ids): Min S2 cell level cannot be greater than max",
		},
		{
			name:       "invalid range, outside limit",
			configData: json.RawMessage(`{"mins2celllevel": -1, "maxs2celllevel": 12}`),
			wantErr:    "failed to create segmenter (s2_ids): S2 cell levels should be in the range 0 - 30",
		},
	}
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			got, err := NewS2IDSegmenter(data.configData)
			if data.wantErr == "" {
				assert.Equalf(t, data.want, got, "NewS2IDSegmenter(%v)", data.configData)
			} else {
				assert.EqualError(t, err, data.wantErr)
			}
		})
	}
}
