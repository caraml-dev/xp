package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/segmenters"
)

type SegmenterService interface {
	GetTransformation(
		id models.ProjectId,
		segmenter string,
		requestValues map[string]interface{},
		experimentVariables []string) ([]*_segmenters.SegmenterValue, error)
}

type segmenterService struct {
	runners      map[string]segmenters.Runner
	localStorage *models.LocalStorage
}

func NewSegmenterService(
	localStorage *models.LocalStorage,
	cfg map[string]interface{},
) (SegmenterService, error) {

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

	return &segmenterService{
		runners:      segmentersRunner,
		localStorage: localStorage,
	}, nil
}

func (svc *segmenterService) GetTransformation(
	projectId models.ProjectId,
	segmenter string,
	requestValues map[string]interface{},
	experimentVariables []string,
) ([]*_segmenters.SegmenterValue, error) {
	err := validateAllPresent(segmenter, requestValues, experimentVariables)
	if err != nil {
		// If not all variables are supplied, we will match for optional segmenters
		// in the experiments. So we can return an empty list of segmenter values to match.
		return []*_segmenters.SegmenterValue{}, nil
	}
	// Check if segmenter is a global segmenter, else use project segmenters
	runner, ok := svc.runners[segmenter]
	if !ok {
		projectSegmentersTypeMapping, err := svc.localStorage.GetSegmentersTypeMapping(projectId)
		if err != nil {
			return nil, err
		}
		segmenterType, ok := projectSegmentersTypeMapping[segmenter]
		if !ok {
			return nil, errors.New("Type mapping not found for Segmenter:" + segmenter)
		}
		projectSegmenterValueType := _segmenters.SegmenterValueType(_segmenters.SegmenterValueType_value[strings.ToUpper(string(segmenterType))])
		runner = segmenters.NewBaseRunner(&segmenters.SegmenterConfig{
			Name: segmenter,
			Type: &projectSegmenterValueType,
		})
	}
	transformation, err := runner.Transform(segmenter, requestValues, experimentVariables)
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
