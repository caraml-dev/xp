package segmenters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/xp/common/api/schema"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

func TestProtobufSegmenterConfigToOpenAPISegmenterConfig(t *testing.T) {
	description := ""
	tests := map[string]struct {
		segments *_segmenters.SegmenterConfiguration
		expected *schema.Segmenter
	}{
		"string values": {
			segments: &_segmenters.SegmenterConfiguration{
				Name: "country",
				Type: 0,
				Options: map[string]*_segmenters.SegmenterValue{
					"singapore": {Value: &_segmenters.SegmenterValue_String_{String_: "SG"}},
					"indonesia": {Value: &_segmenters.SegmenterValue_String_{String_: "ID"}},
				},
				MultiValued: false,
				TreatmentRequestFields: &_segmenters.ListExperimentVariables{
					Values: []*_segmenters.ExperimentVariables{
						{
							Value: []string{"country"},
						},
					},
				},
				Constraints: []*_segmenters.Constraint{
					{
						PreRequisites: []*_segmenters.PreRequisite{
							{
								SegmenterName: "test",
								SegmenterValues: &_segmenters.ListSegmenterValue{
									Values: []*_segmenters.SegmenterValue{
										{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
										{Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
									},
								},
							},
						},
						AllowedValues: &_segmenters.ListSegmenterValue{
							Values: []*_segmenters.SegmenterValue{
								{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
								{Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
								{Value: &_segmenters.SegmenterValue_Integer{Integer: 3}},
							},
						},
						Options: map[string]*_segmenters.SegmenterValue{
							"singapore_new": {Value: &_segmenters.SegmenterValue_String_{String_: "SG"}},
							"indonesia_new": {Value: &_segmenters.SegmenterValue_String_{String_: "ID"}},
						},
					},
				},
			},
			expected: &schema.Segmenter{
				Constraints: []schema.Constraint{
					{
						PreRequisites: []schema.PreRequisite{
							{
								SegmenterName:   "test",
								SegmenterValues: []schema.SegmenterValues{1, 2},
							},
						},
						AllowedValues: []schema.SegmenterValues{1, 2, 3},
						Options: &schema.SegmenterOptions{
							AdditionalProperties: map[string]interface{}{
								"singapore_new": "SG",
								"indonesia_new": "ID",
							},
						},
					},
				},
				MultiValued: false,
				Name:        "country",
				Description: &description,
				Options: schema.SegmenterOptions{
					AdditionalProperties: map[string]interface{}{
						"singapore": "SG",
						"indonesia": "ID",
					},
				},
				TreatmentRequestFields: [][]string{
					{"country"},
				},
				Type: schema.SegmenterTypeString,
			},
		},
		"bool values": {
			segments: &_segmenters.SegmenterConfiguration{
				Name: "bool_test",
				Type: 1,
				Options: map[string]*_segmenters.SegmenterValue{
					"Yes": {Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
					"No":  {Value: &_segmenters.SegmenterValue_Bool{Bool: false}},
				},
				MultiValued: true,
				TreatmentRequestFields: &_segmenters.ListExperimentVariables{
					Values: []*_segmenters.ExperimentVariables{
						{
							Value: []string{"bool_test"},
						},
					},
				},
				Constraints: nil,
			},
			expected: &schema.Segmenter{
				Constraints: nil,
				MultiValued: true,
				Name:        "bool_test",
				Description: &description,
				Options: schema.SegmenterOptions{
					AdditionalProperties: map[string]interface{}{
						"Yes": true,
						"No":  false,
					},
				},
				TreatmentRequestFields: [][]string{
					{"bool_test"},
				},
				Type: schema.SegmenterTypeBool,
			},
		},
		"integer values": {
			segments: &_segmenters.SegmenterConfiguration{
				Name: "days_of_week",
				Type: 2,
				Options: map[string]*_segmenters.SegmenterValue{
					"Monday":    {Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
					"Tuesday":   {Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
					"Wednesday": {Value: &_segmenters.SegmenterValue_Integer{Integer: 3}},
					"Thursday":  {Value: &_segmenters.SegmenterValue_Integer{Integer: 4}},
					"Friday":    {Value: &_segmenters.SegmenterValue_Integer{Integer: 5}},
				},
				MultiValued: true,
				TreatmentRequestFields: &_segmenters.ListExperimentVariables{
					Values: []*_segmenters.ExperimentVariables{
						{
							Value: []string{"tz"},
						},
					},
				},
				Constraints: nil,
			},
			expected: &schema.Segmenter{
				Constraints: nil,
				MultiValued: true,
				Name:        "days_of_week",
				Description: &description,
				Options: schema.SegmenterOptions{
					AdditionalProperties: map[string]interface{}{
						"Monday":    1,
						"Tuesday":   2,
						"Wednesday": 3,
						"Thursday":  4,
						"Friday":    5,
					},
				},
				TreatmentRequestFields: [][]string{
					{"tz"},
				},
				Type: schema.SegmenterTypeInteger,
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ProtobufSegmenterConfigToOpenAPISegmenterConfig(data.segments)
			require.NoError(t, err)

			// Compare the JSON representations
			expectedJSON, err := json.Marshal(data.expected)
			require.NoError(t, err)
			actualJSON, err := json.Marshal(got)
			require.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestToProtoValues(t *testing.T) {
	segmentersType := map[string]schema.SegmenterType{
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"string_segmenter":  schema.SegmenterTypeString,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}
	experimentSegmentListInteger := map[string]*_segmenters.ListSegmenterValue{
		"integer_segmenter": {Values: []*_segmenters.SegmenterValue{
			{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		}},
	}
	experimentSegmentListString := map[string]*_segmenters.ListSegmenterValue{
		"string_segmenter": {Values: []*_segmenters.SegmenterValue{
			{Value: &_segmenters.SegmenterValue_String_{String_: "1"}},
		}},
	}
	experimentSegmentListBool := map[string]*_segmenters.ListSegmenterValue{
		"bool_segmenter": {Values: []*_segmenters.SegmenterValue{
			{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
		}},
	}
	errInteger := "received wrong type of segmenter value; integer_segmenter expects type integer"
	errFloat := "received wrong type of segmenter value; float_segmenter expects type real"
	errString := "received wrong type of segmenter value; string_segmenter expects type string"
	errBool := "received wrong type of segmenter value; bool_segmenter expects type bool"

	tests := []struct {
		name           string
		segment        map[string]interface{}
		segmentersType map[string]schema.SegmenterType
		expected       map[string]*_segmenters.ListSegmenterValue
		err            *string
	}{
		{
			name:           "invalid type | expected integer, got string",
			segment:        map[string]interface{}{"integer_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			err:            &errInteger,
		},
		{
			name:           "invalid type | expected float, got string",
			segment:        map[string]interface{}{"float_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			err:            &errFloat,
		},
		{
			name:           "invalid type | expected string, got integer",
			segment:        map[string]interface{}{"string_segmenter": []interface{}{float64(1)}},
			segmentersType: segmentersType,
			err:            &errString,
		},
		{
			name:           "invalid type | expected bool, got integer",
			segment:        map[string]interface{}{"bool_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			err:            &errBool,
		},
		{
			name:           "success | integer",
			segment:        map[string]interface{}{"integer_segmenter": []interface{}{float64(1)}},
			segmentersType: segmentersType,
			expected:       experimentSegmentListInteger,
		},
		{
			name:           "success | string",
			segment:        map[string]interface{}{"string_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			expected:       experimentSegmentListString,
		},
		{
			name:           "success | bool",
			segment:        map[string]interface{}{"bool_segmenter": []interface{}{true}},
			segmentersType: segmentersType,
			expected:       experimentSegmentListBool,
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			got, err := ToProtoValues(data.segment, data.segmentersType)

			if data.err != nil {
				assert.EqualError(t, fmt.Errorf("%s", *data.err), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, got)
			}
		})
	}
}
