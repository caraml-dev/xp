package segmenters

import (
	"fmt"

	_segmenters "github.com/gojek/xp/common/segmenters"
	_utils "github.com/gojek/xp/common/utils"
)

const TypeCastingErrorTmpl = "invalid type of variable (%s) was provided for %s segmenter; expected %s"

type SegmenterConfig struct {
	Name string
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
	transformedVals := []*_segmenters.SegmenterValue{}
	convertedVal := _utils.InterfaceToSegmenterValue(requestValues[segmenter])
	transformedVals = append(transformedVals, convertedVal)
	if convertedVal == nil {
		return nil, fmt.Errorf("segmenter type for %s is not supported", segmenter)
	}
	return transformedVals, nil
}
