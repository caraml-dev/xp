package models

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojek/xp/common/api/schema"
	_segmenters "github.com/gojek/xp/common/segmenters"
	_utils "github.com/gojek/xp/common/utils"
)

var testSegment = ExperimentSegment{
	"string_segmenter":  []string{"seg-1", "seg-2"},
	"integer_segmenter": []string{"1", "2"},
	"float_segmenter":   []string{"3.0", "4.0"},
	"bool_segmenter":    []string{"false"},
}

func TestSegmentValue(t *testing.T) {
	value, err := testSegment.Value()
	// Convert to string for comparison
	byteValue, ok := value.([]byte)
	assert.True(t, ok)
	// Validate
	assert.NoError(t, err)
	assert.JSONEq(t, `
		{
			"string_segmenter": ["seg-1", "seg-2"],
			"integer_segmenter": ["1", "2"],
			"float_segmenter": ["3.0", "4.0"],
			"bool_segmenter": ["false"]
		}
	`, string(byteValue))
}

func TestSegmentScan(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		errString string
		expected  ExperimentSegment
	}{
		{
			name: "success",
			value: []byte(`
				{
					"string_segmenter": ["seg-1", "seg-2"],
					"integer_segmenter": ["1", "2"],
					"float_segmenter": ["3.0", "4.0"],
					"bool_segmenter": ["false"]
				}
			`),
			expected: testSegment,
		},
		{
			name:      "failure | invalid value",
			value:     100,
			errString: "type assertion to []byte failed",
		},
	}

	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			var segment ExperimentSegment
			err := segment.Scan(data.value)
			if data.errString == "" {
				// Success
				require.NoError(t, err)
				assert.Equal(t, data.expected, segment)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}

func TestSegmentToApiSchema(t *testing.T) {
	segmenterTypes := map[string]schema.SegmenterType{
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"string_segmenter":  schema.SegmenterTypeString,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}

	experimentSegment := schema.ExperimentSegment{
		"string_segmenter":  []string{"seg-1", "seg-2"},
		"integer_segmenter": []int64{1, 2},
		"float_segmenter":   []float64{3.0, 4.0},
		"bool_segmenter":    []bool{false},
	}
	assert.Equal(t, experimentSegment, testSegment.ToApiSchema(segmenterTypes))
}

func TestSegmentToProtoSchema(t *testing.T) {
	protoSchema := map[string]*_segmenters.ListSegmenterValue{
		"string_segmenter":  _utils.StringSliceToListSegmenterValue(&[]string{"seg-1", "seg-2"}),
		"integer_segmenter": _utils.Int64ListToListSegmenterValue(&[]int64{int64(1), int64(2)}),
		"float_segmenter":   _utils.FloatListToListSegmenterValue(&[]float64{3.0, 4.0}),
		"bool_segmenter":    _utils.BoolSliceToListSegmenterValue(&[]bool{false}),
	}

	segmenterTypes := map[string]schema.SegmenterType{
		"string_segmenter":  schema.SegmenterTypeString,
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}

	// Compare the JSON representations
	expectedJSON, err := json.Marshal(protoSchema)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(testSegment.ToProtoSchema(segmenterTypes))
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func TestSegmentToStorageSchema(t *testing.T) {
	segmentersType := map[string]schema.SegmenterType{
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"string_segmenter":  schema.SegmenterTypeString,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}
	experimentIntSegment := ExperimentSegment{
		"integer_segmenter": []string{"1"},
	}
	experimentStringSegment := ExperimentSegment{
		"string_segmenter": []string{"1"},
	}
	experimentBoolSegment := ExperimentSegment{
		"bool_segmenter": []string{"true"},
	}
	errInteger := "received wrong type of segmenter value; integer_segmenter expects type integer"
	errFloat := "received wrong type of segmenter value; float_segmenter expects type real"
	errString := "received wrong type of segmenter value; string_segmenter expects type string"
	errBool := "received wrong type of segmenter value; bool_segmenter expects type bool"

	tests := []struct {
		name           string
		segment        ExperimentSegmentRaw
		segmentersType map[string]schema.SegmenterType
		expected       ExperimentSegment
		err            *string
	}{
		{
			name:           "invalid type | expected integer, got string",
			segment:        ExperimentSegmentRaw{"integer_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			err:            &errInteger,
		},
		{
			name:           "invalid type | expected float, got string",
			segment:        ExperimentSegmentRaw{"float_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			err:            &errFloat,
		},
		{
			name:           "invalid type | expected string, got integer",
			segment:        ExperimentSegmentRaw{"string_segmenter": []interface{}{float64(1)}},
			segmentersType: segmentersType,
			err:            &errString,
		},
		{
			name:           "invalid type | expected bool, got integer",
			segment:        ExperimentSegmentRaw{"bool_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			err:            &errBool,
		},
		{
			name:           "success | integer",
			segment:        ExperimentSegmentRaw{"integer_segmenter": []interface{}{float64(1)}},
			segmentersType: segmentersType,
			expected:       experimentIntSegment,
		},
		{
			name:           "success | string",
			segment:        ExperimentSegmentRaw{"string_segmenter": []interface{}{"1"}},
			segmentersType: segmentersType,
			expected:       experimentStringSegment,
		},
		{
			name:           "success | bool",
			segment:        ExperimentSegmentRaw{"bool_segmenter": []interface{}{true}},
			segmentersType: segmentersType,
			expected:       experimentBoolSegment,
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			got, err := data.segment.ToStorageSchema(segmentersType)

			if data.err != nil {
				assert.EqualError(t, fmt.Errorf("%s", *data.err), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, got)
			}
		})
	}
}

func TestSegmentToRawSchema(t *testing.T) {
	segmentersType := map[string]schema.SegmenterType{
		"integer_segmenter": schema.SegmenterTypeInteger,
		"float_segmenter":   schema.SegmenterTypeReal,
		"string_segmenter":  schema.SegmenterTypeString,
		"bool_segmenter":    schema.SegmenterTypeBool,
	}
	experimentIntSegment := ExperimentSegment{
		"integer_segmenter": []string{"1"},
	}
	experimentFloatSegment := ExperimentSegment{
		"float_segmenter": []string{"1"},
	}
	experimentStringSegment := ExperimentSegment{
		"string_segmenter": []string{"1"},
	}
	experimentBoolSegment := ExperimentSegment{
		"bool_segmenter": []string{"true"},
	}

	tests := []struct {
		name           string
		segment        ExperimentSegment
		segmentersType map[string]schema.SegmenterType
		expected       ExperimentSegmentRaw
	}{
		{
			name:           "success | integer",
			segment:        experimentIntSegment,
			segmentersType: segmentersType,
			expected:       ExperimentSegmentRaw{"integer_segmenter": []interface{}{float64(1)}},
		},
		{
			name:           "success | float",
			segment:        experimentFloatSegment,
			segmentersType: segmentersType,
			expected:       ExperimentSegmentRaw{"float_segmenter": []interface{}{float64(1)}},
		},
		{
			name:           "success | string",
			segment:        experimentStringSegment,
			segmentersType: segmentersType,
			expected:       ExperimentSegmentRaw{"string_segmenter": []interface{}{"1"}},
		},
		{
			name:           "success | bool",
			segment:        experimentBoolSegment,
			segmentersType: segmentersType,
			expected:       ExperimentSegmentRaw{"bool_segmenter": []interface{}{true}},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			got := data.segment.ToRawSchema(segmentersType)
			assert.Equal(t, data.expected, got)
		})
	}
}
