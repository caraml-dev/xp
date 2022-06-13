package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
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
	tests := []struct {
		Name           string
		SegmenterValue interface{}
		Expected       *_segmenters.SegmenterValue
	}{
		{
			Name:           "success | string",
			SegmenterValue: "test",
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "test"}},
		},
		{
			Name:           "success | int",
			SegmenterValue: int64(1),
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
		},
		{
			Name:           "success | float",
			SegmenterValue: 1.0,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}},
		},
		{
			Name:           "success | bool",
			SegmenterValue: true,
			Expected:       &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.Name, func(t *testing.T) {
			val := InterfaceToSegmenterValue(data.SegmenterValue)
			assert.Equal(t, data.Expected, val)
		})
	}
}
