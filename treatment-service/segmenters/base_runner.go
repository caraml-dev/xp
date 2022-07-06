package segmenters

import (
	"fmt"

	_segmenters "github.com/gojek/xp/common/segmenters"
	_utils "github.com/gojek/xp/common/utils"
)

const TypeCastingErrorTmpl = "invalid type of variable (%s) was provided for %s segmenter; expected %s"

type SegmenterConfig struct {
	Name string
	Type *_segmenters.SegmenterValueType
}

type BaseRunner struct {
	config *SegmenterConfig
}

func NewBaseRunner(s *SegmenterConfig) *BaseRunner {
	return &BaseRunner{
		config: s,
	}
}

func (r *BaseRunner) GetName() string {
	return r.config.Name
}

func (r *BaseRunner) Transform(
	segmenter string,
	requestValues map[string]interface{},
	experimentVariables []string,
) ([]*_segmenters.SegmenterValue, error) {
	errTmpl := fmt.Sprintf("received wrong type of segmenter value; %s expects type", segmenter)
	transformedVals := []*_segmenters.SegmenterValue{}

	convertedVal := _utils.InterfaceToSegmenterValue(requestValues[segmenter], r.config.Type)
	if convertedVal == nil {
		return nil, fmt.Errorf("segmenter type for %s is not supported", segmenter)
	}
	// If type is provided, validate the converted val. Do conversion for integer.
	if r.config.Type != nil {
		switch *r.config.Type {
		case _segmenters.SegmenterValueType_STRING:
			if _, ok := convertedVal.GetValue().(*_segmenters.SegmenterValue_String_); !ok {
				return nil, fmt.Errorf("%s %s", errTmpl, _segmenters.SegmenterValueType_STRING.String())
			}
		case _segmenters.SegmenterValueType_BOOL:
			if _, ok := convertedVal.GetValue().(*_segmenters.SegmenterValue_Bool); !ok {
				return nil, fmt.Errorf("%s %s", errTmpl, _segmenters.SegmenterValueType_BOOL.String())
			}
		case _segmenters.SegmenterValueType_INTEGER:
			if _, ok := convertedVal.GetValue().(*_segmenters.SegmenterValue_Integer); !ok {
				// JSON value sent through browser will be float64 by javascript syntax and definition;
				// the check below ensures that val is minimally a number-like variable
				if _, ok := convertedVal.GetValue().(*_segmenters.SegmenterValue_Real); ok {
					convertedVal = &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(convertedVal.GetReal())}}
					break
				}
				return nil, fmt.Errorf("%s %s", errTmpl, _segmenters.SegmenterValueType_INTEGER.String())
			}
		case _segmenters.SegmenterValueType_REAL:
			if _, ok := convertedVal.GetValue().(*_segmenters.SegmenterValue_Real); !ok {
				return nil, fmt.Errorf("%s %s", errTmpl, _segmenters.SegmenterValueType_REAL.String())
			}
		}
	}
	transformedVals = append(transformedVals, convertedVal)
	return transformedVals, nil
}
