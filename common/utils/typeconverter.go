package utils

import (
	"fmt"
	"reflect"

	"github.com/spf13/cast"

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

func InterfaceToSegmenterValue(value interface{}, segmenter string, valueType *_segmenters.SegmenterValueType) (*_segmenters.SegmenterValue, error) {
	// If value type is not defined, use reflection as base implementation
	var segmenterValue *_segmenters.SegmenterValue
	incorrectSegmenterTypeErrTmpl := fmt.Errorf("segmenter type for %s is not supported", segmenter)
	if valueType == nil {
		val := reflect.ValueOf(value)
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
			return nil, incorrectSegmenterTypeErrTmpl
		}
		return segmenterValue, nil
	} else {
		switch *valueType {
		case _segmenters.SegmenterValueType_STRING:
			stringVal, ok := value.(string)
			if !ok {
				return nil, incorrectSegmenterTypeErrTmpl
			}
			segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: stringVal}}
		case _segmenters.SegmenterValueType_INTEGER:
			intVal, err := cast.ToInt64E(value)
			if err != nil {
				return nil, err
			}
			segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: intVal}}
		case _segmenters.SegmenterValueType_REAL:
			floatVal, err := cast.ToFloat64E(value)
			if err != nil {
				return nil, err
			}
			segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: floatVal}}
		case _segmenters.SegmenterValueType_BOOL:
			boolVal, err := cast.ToBoolE(value)
			if err != nil {
				return nil, err
			}
			segmenterValue = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: boolVal}}
		default:
			return nil, incorrectSegmenterTypeErrTmpl
		}
	}
	return segmenterValue, nil
}
