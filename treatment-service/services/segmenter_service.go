package services

import (
	"encoding/json"
	"fmt"

	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
	"github.com/gojek/turing-experiments/treatment-service/segmenters"
)

type SegmenterService interface {
	GetTransformation(segmenter string, requestValues map[string]interface{}, experimentVariables []string) ([]*_segmenters.SegmenterValue, error)
}

type segmenterService struct {
	runners map[string]segmenters.Runner
}

func NewSegmenterService(cfg map[string]interface{}) (SegmenterService, error) {
	segmentersRunner := make(map[string]segmenters.Runner)

	for name := range segmenters.Runners {
		if _, ok := cfg[name]; ok {
			configJSON, err := json.Marshal(cfg[name])
			if err != nil {
				return nil, err
			}

			m, err := segmenters.Get(name, configJSON)
			if err != nil {
				return nil, err
			}
			segmentersRunner[name] = m
			continue
		}
		m, err := segmenters.Get(name, nil)
		if err != nil {
			return nil, err
		}
		segmentersRunner[name] = m
	}

	return &segmenterService{runners: segmentersRunner}, nil
}

func (svc *segmenterService) GetTransformation(
	segmenter string,
	requestValues map[string]interface{},
	experimentVariables []string,
) ([]*_segmenters.SegmenterValue, error) {
	err := validateAllPresent(segmenter, requestValues, experimentVariables)
	if err != nil {
		return nil, err
	}

	transformation, err := svc.runners[segmenter].Transform(segmenter, requestValues, experimentVariables)
	if err != nil {
		return nil, err
	}

	return transformation, nil
}

func validateAllPresent(segmenter string, providedVariables map[string]interface{}, requiredVariables []string) error {
	for _, requiredVariable := range requiredVariables {
		if _, ok := providedVariables[requiredVariable]; !ok {
			return fmt.Errorf("experiment variable (%s) was not provided for segmenter (%s)", requiredVariable, segmenter)
		}
	}

	return nil
}
