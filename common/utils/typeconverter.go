package utils

import (
	"reflect"

	_segmenters "github.com/gojek/xp/common/segmenters"
)

func StringSliceToListSegmenterValue(values *[]string) *_segmenters.ListSegmenterValue {
	if values == nil {
		return nil
	}
	segmenterValues := []*_segmenters.SegmenterValue{}
	for _, val := range *values {
		segmenterValues = append(segmenterValues, &_segmenters.SegmenterValue{
			Value: &_segmenters.SegmenterValue_String_{String_: val},
		})
	}
	return &_segmenters.ListSegmenterValue{Values: segmenterValues}
}

func BoolSliceToListSegmenterValue(values *[]bool) *_segmenters.ListSegmenterValue {
	if values == nil {
		return nil
	}
	segmenterValues := []*_segmenters.SegmenterValue{}
	for _, val := range *values {
		segmenterValues = append(segmenterValues, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: val}})
	}
	return &_segmenters.ListSegmenterValue{Values: segmenterValues}
}

func Int64ListToListSegmenterValue(values *[]int64) *_segmenters.ListSegmenterValue {
	if values == nil {
		return nil
	}
	segmenterValues := []*_segmenters.SegmenterValue{}
	for _, val := range *values {
		segmenterValues = append(segmenterValues, &_segmenters.SegmenterValue{
			Value: &_segmenters.SegmenterValue_Integer{Integer: val},
		})
	}
	return &_segmenters.ListSegmenterValue{Values: segmenterValues}
}

func FloatListToListSegmenterValue(values *[]float64) *_segmenters.ListSegmenterValue {
	if values == nil {
		return nil
	}
	segmenterValues := []*_segmenters.SegmenterValue{}
	for _, val := range *values {
		segmenterValues = append(segmenterValues, &_segmenters.SegmenterValue{
			Value: &_segmenters.SegmenterValue_Real{Real: val},
		})
	}
	return &_segmenters.ListSegmenterValue{Values: segmenterValues}
}

func SegmenterValueToInterface(value *_segmenters.SegmenterValue) interface{} {
	switch value.Value.(type) {
	case *_segmenters.SegmenterValue_String_:
		return value.GetString_()
	case *_segmenters.SegmenterValue_Integer:
		return value.GetInteger()
	case *_segmenters.SegmenterValue_Real:
		return value.GetReal()
	case *_segmenters.SegmenterValue_Bool:
		return value.GetBool()
	default:
		return nil
	}
}

func InterfaceToSegmenterValue(value interface{}) *_segmenters.SegmenterValue {
	val := reflect.ValueOf(value)
	var segmenterValue *_segmenters.SegmenterValue
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: val.Int()}}
	case reflect.Float32, reflect.Float64:
		segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: val.Float()}}
	case reflect.String:
		segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: val.String()}}
	case reflect.Bool:
		segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: val.Bool()}}
	default:
		return nil
	}

	return segmenterValue
}
