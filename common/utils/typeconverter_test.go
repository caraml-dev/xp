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
		SegmenterValue interface{}
		SegmenterType  *_segmenters.SegmenterValueType
		Expected       *_segmenters.SegmenterValue
	}{
		{
			Name:           "success | infer string",
			SegmenterValue: "test",
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "test"}},
		},
		{
			Name:           "success | string",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeString,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "test"}},
		},
		{
			Name:           "failure | string, wrong input type",
			SegmenterValue: 1,
			SegmenterType:  &segmenterTypeString,
			Expected:       nil,
		},
		{
			Name:           "success | infer int",
			SegmenterValue: int64(1),
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		},
		{
			Name:           "success | int, float input json test",
			SegmenterValue: 1.0,
			SegmenterType:  &segmenterTypeInt,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		},
		{
			Name:           "success | int",
			SegmenterValue: int64(1),
			SegmenterType:  &segmenterTypeInt,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		},
		{
			Name:           "failure | int, wrong input type",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeInt,
			Expected:       nil,
		},
		{
			Name:           "success | infer float",
			SegmenterValue: 1.0,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}},
		},
		{
			Name:           "success | float",
			SegmenterValue: 1.0,
			SegmenterType:  &segmenterTypeReal,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}},
		},
		{
			Name:           "failure | float, wrong input type",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeReal,
			Expected:       nil,
		},
		{
			Name:           "success | infer bool",
			SegmenterValue: true,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
		},
		{
			Name:           "success | bool",
			SegmenterValue: true,
			SegmenterType:  &segmenterTypeBool,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
		},
		{
			Name:           "failure | bool, wrong input type",
			SegmenterValue: "test",
			SegmenterType:  &segmenterTypeBool,
			Expected:       nil,
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.Name, func(t *testing.T) {
			val := InterfaceToSegmenterValue(data.SegmenterValue, data.SegmenterType)
			assert.Equal(t, data.Expected, val)
		})
	}
}
