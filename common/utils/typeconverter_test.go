package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_segmenters "github.com/gojek/xp/common/segmenters"
)

func TestStringSliceToListSegmenterValue(t *testing.T) {
	segmenter := &[]string{"test"}
	segmenterValue := StringSliceToListSegmenterValue(segmenter)
	expectedValue := &_segmenters.ListSegmenterValue{
		Values: []*_segmenters.SegmenterValue{
			{
				Value: &_segmenters.SegmenterValue_String_{String_: "test"},
			},
		},
	}
	assert.Equal(t, expectedValue, segmenterValue)
}

func TestBoolSliceToListSegmenterValue(t *testing.T) {
	segmenter := &[]bool{true}
	segmenterValue := BoolSliceToListSegmenterValue(segmenter)
	expectedValue := &_segmenters.ListSegmenterValue{
		Values: []*_segmenters.SegmenterValue{
			{
				Value: &_segmenters.SegmenterValue_Bool{Bool: true},
			},
		},
	}
	assert.Equal(t, expectedValue, segmenterValue)
}

func TestInt64ListToListSegmenterValue(t *testing.T) {
	segmenter := &[]int64{1}
	segmenterValue := Int64ListToListSegmenterValue(segmenter)
	expectedValue := &_segmenters.ListSegmenterValue{
		Values: []*_segmenters.SegmenterValue{
			{
				Value: &_segmenters.SegmenterValue_Integer{Integer: 1},
			},
		},
	}
	assert.Equal(t, expectedValue, segmenterValue)
}

func TestFloatListToListSegmenterValue(t *testing.T) {
	segmenter := &[]float64{1.0}
	segmenterValue := FloatListToListSegmenterValue(segmenter)
	expectedValue := &_segmenters.ListSegmenterValue{
		Values: []*_segmenters.SegmenterValue{
			{
				Value: &_segmenters.SegmenterValue_Real{Real: 1.0},
			},
		},
	}
	assert.Equal(t, expectedValue, segmenterValue)
}

func TestSegmenterValueToInterface(t *testing.T) {
	tests := []struct {
		Name           string
		SegmenterValue *_segmenters.SegmenterValue
		Expected       interface{}
	}{
		{
			Name:           "success | string",
			SegmenterValue: &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "test"}},
			Expected:       "test",
		},
		{
			Name:           "success | int",
			SegmenterValue: &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
			Expected:       int64(1),
		},
		{
			Name:           "success | float",
			SegmenterValue: &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}},
			Expected:       1.0,
		},
		{
			Name:           "success | bool",
			SegmenterValue: &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
			Expected:       true,
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.Name, func(t *testing.T) {
			val := SegmenterValueToInterface(data.SegmenterValue)
			assert.Equal(t, data.Expected, val)
		})
	}
}

func TestInterfaceToSegmenterValue(t *testing.T) {
	segmenterTypeString := _segmenters.SegmenterValueType_STRING
	segmenterTypeBool := _segmenters.SegmenterValueType_BOOL
	segmenterTypeInt := _segmenters.SegmenterValueType_INTEGER
	segmenterTypeReal := _segmenters.SegmenterValueType_REAL

	tests := []struct {
		Name           string
		SegmenterName  string
		SegmenterValue interface{}
		SegmenterType  *_segmenters.SegmenterValueType
		Expected       *_segmenters.SegmenterValue
		ErrString      string
	}{
		{
			Name:           "success | infer string",
			SegmenterName:  "seg-name",
			SegmenterValue: "test",
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "test"}},
		},
		{
			Name:           "success | string",
			SegmenterName:  "seg-name",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeString,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "test"}},
		},
		{
			Name:           "failure | string, wrong input type",
			SegmenterName:  "seg-name",
			SegmenterValue: 1,
			SegmenterType:  &segmenterTypeString,
			Expected:       nil,
			ErrString:      "segmenter type for seg-name is not supported",
		},
		{
			Name:           "success | infer int",
			SegmenterName:  "seg-name",
			SegmenterValue: int64(1),
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		},
		{
			Name:           "success | int, float input json test",
			SegmenterName:  "seg-name",
			SegmenterValue: 1.0,
			SegmenterType:  &segmenterTypeInt,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		},
		{
			Name:           "success | int",
			SegmenterName:  "seg-name",
			SegmenterValue: int64(1),
			SegmenterType:  &segmenterTypeInt,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		},
		{
			Name:           "failure | int, wrong input type",
			SegmenterName:  "seg-name",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeInt,
			Expected:       nil,
			ErrString:      "invalid type of variable (seg-name) was provided for seg-name segmenter; expected int64",
		},
		{
			Name:           "success | infer float",
			SegmenterName:  "seg-name",
			SegmenterValue: 1.0,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}},
		},
		{
			Name:           "success | float",
			SegmenterName:  "seg-name",
			SegmenterValue: 1.0,
			SegmenterType:  &segmenterTypeReal,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}},
		},
		{
			Name:           "failure | float, wrong input type",
			SegmenterName:  "seg-name",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeReal,
			Expected:       nil,
			ErrString:      "invalid type of variable (seg-name) was provided for seg-name segmenter; expected float64",
		},
		{
			Name:           "success | infer bool",
			SegmenterName:  "seg-name",
			SegmenterValue: true,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
		},
		{
			Name:           "success | bool",
			SegmenterName:  "seg-name",
			SegmenterValue: true,
			SegmenterType:  &segmenterTypeBool,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
		},
		{
			Name:           "failure | bool, wrong input type",
			SegmenterName:  "seg-name",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeBool,
			Expected:       nil,
			ErrString:      "invalid type of variable (seg-name) was provided for seg-name segmenter; expected bool",
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.Name, func(t *testing.T) {
			val, err := InterfaceToSegmenterValue(data.SegmenterValue, data.SegmenterName, data.SegmenterType)

			if data.ErrString == "" {
				assert.NoError(t, err)
				assert.Equal(t, data.Expected, val)
			} else {
				assert.EqualError(t, err, data.ErrString)
			}
		})
	}
}

func TestGetIntSegmenter(t *testing.T) {
	tests := map[string]struct {
		values    interface{}
		key       string
		segmenter string
		expected  int64
		errString string
	}{
		"failure | bool": {
			values:    true,
			key:       "key",
			segmenter: "int-seg",
			errString: "invalid type of variable (key) was provided for int-seg segmenter; expected float64",
		},
		"success | float64": {
			values:    1.0,
			key:       "key",
			segmenter: "int-seg",
			expected:  1.0,
		},
		"success | int64": {
			values:    int64(1),
			key:       "key",
			segmenter: "int-seg",
			expected:  1,
		},
		"success | string": {
			values:    "1",
			key:       "key",
			segmenter: "int-seg",
			expected:  1,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := GetIntSegmenter(data.values, data.key, data.segmenter)

			if data.errString == "" {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, *resp)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}

func TestGetFloatSegmenter(t *testing.T) {
	tests := map[string]struct {
		values    interface{}
		key       string
		segmenter string
		expected  float64
		errString string
	}{
		"failure | bool": {
			values:    true,
			key:       "key",
			segmenter: "float-seg",
			errString: "invalid type of variable (key) was provided for float-seg segmenter; expected float64",
		},
		"success | float64": {
			values:    1.2537040223936706,
			key:       "key",
			segmenter: "float-seg",
			expected:  1.2537040223936706,
		},
		"success | string": {
			values:    "1.2537040223936706",
			key:       "key",
			segmenter: "float-seg",
			expected:  1.2537040223936706,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := GetFloatSegmenter(data.values, data.key, data.segmenter)

			if data.errString == "" {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, *resp)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}

func TestGetBoolSegmenter(t *testing.T) {
	tests := map[string]struct {
		values    interface{}
		key       string
		segmenter string
		expected  bool
		errString string
	}{
		"failure | bool": {
			values:    1.23,
			key:       "key",
			segmenter: "bool-seg",
			errString: "invalid type of variable (key) was provided for bool-seg segmenter; expected bool",
		},
		"success | bool": {
			values:    true,
			key:       "key",
			segmenter: "bool-seg",
			expected:  true,
		},
		"success | string": {
			values:    "true",
			key:       "key",
			segmenter: "bool-seg",
			expected:  true,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := GetBoolSegmenter(data.values, data.key, data.segmenter)

			if data.errString == "" {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, *resp)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}
