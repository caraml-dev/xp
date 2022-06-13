package services

import (
	"encoding/json"
	"fmt"

	"github.com/golang-collections/collections/set"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/gojek/turing-experiments/common/api/schema"
	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
	"github.com/gojek/turing-experiments/management-service/models"
	"github.com/gojek/turing-experiments/management-service/segmenters"
)

type SegmenterService interface {
	GetFormattedSegmenters(expSegment models.ExperimentSegmentRaw) (map[string]*[]interface{}, error)
	GetSegmenterConfigurations(segmenterNames []string) ([]*_segmenters.SegmenterConfiguration, error)
	ValidateExperimentSegment(userSegmenters []string, expSegment models.ExperimentSegmentRaw) error
	ValidateSegmentOrthogonality(
		userSegmenters []string,
		expSegment models.ExperimentSegmentRaw,
		allExps []models.Experiment,
	) error
	ValidatePrereqSegmenters(segmenters []string) error
	ValidateRequiredSegmenters(segmenters []string) error
	ValidateExperimentVariables(projectSegmenters models.ProjectSegmenters) error
	ListSegmenterNames() []string
	GetSegmenterTypes() map[string]schema.SegmenterType
}

type segmenterService struct {
	segmenters map[string]segmenters.Segmenter
}

func NewSegmenterService(cfg map[string]interface{}) (SegmenterService, error) {
	experimentSegmenters := make(map[string]segmenters.Segmenter)

	for name := range segmenters.Segmenters {
		if _, ok := cfg[name]; ok {
			configJSON, err := json.Marshal(cfg[name])
			if err != nil {
				return nil, err
			}

			m, err := segmenters.Get(name, configJSON)
			if err != nil {
				return nil, err
			}
			experimentSegmenters[name] = m
			continue
		}
		m, err := segmenters.Get(name, nil)
		if err != nil {
			return nil, err
		}
		experimentSegmenters[name] = m
	}

	return &segmenterService{segmenters: experimentSegmenters}, nil
}

func (svc *segmenterService) ListSegmenterNames() []string {
	segmenterNameList := make([]string, len(svc.segmenters))

	i := 0
	for segmenterName := range svc.segmenters {
		segmenterNameList[i] = segmenterName
		i++
	}

	return segmenterNameList
}

func (svc *segmenterService) GetSegmenterTypes() map[string]schema.SegmenterType {
	segmenterTypes := map[string]schema.SegmenterType{}

	for key, val := range svc.segmenters {
		switch val.GetType() {
		case _segmenters.SegmenterValueType_STRING:
			segmenterTypes[key] = schema.SegmenterTypeString
		case _segmenters.SegmenterValueType_INTEGER:
			segmenterTypes[key] = schema.SegmenterTypeInteger
		case _segmenters.SegmenterValueType_REAL:
			segmenterTypes[key] = schema.SegmenterTypeReal
		case _segmenters.SegmenterValueType_BOOL:
			segmenterTypes[key] = schema.SegmenterTypeBool
		}
	}

	return segmenterTypes
}

func (svc *segmenterService) GetSegmenterConfigurations(segmenterNames []string) ([]*_segmenters.SegmenterConfiguration, error) {
	// Convert to a generic interface map with formatted values
	segmenterConfigList := []*_segmenters.SegmenterConfiguration{}

	for _, segmenterName := range segmenterNames {
		segmenter, err := svc.getSegmenter(segmenterName)
		if err != nil {
			return segmenterConfigList, err
		}

		config, err := segmenter.GetConfiguration()
		if err != nil {
			return segmenterConfigList, err
		}
		segmenterConfigList = append(segmenterConfigList, config)
	}

	return segmenterConfigList, nil
}

func (svc *segmenterService) GetFormattedSegmenters(expSegment models.ExperimentSegmentRaw) (map[string]*[]interface{}, error) {
	inputSegmenters, err := segmenters.ToProtoValues(expSegment, svc.GetSegmenterTypes())
	if err != nil {
		return nil, err
	}

	// Convert to a generic interface map with formatted values
	formattedMap := map[string]*[]interface{}{}

	for segmenterName, values := range inputSegmenters {
		segmenter, err := svc.getSegmenter(segmenterName)
		if err != nil {
			return formattedMap, err
		}

		if values != nil {
			// Format the segmenter values and add to the map
			segmenterType := segmenter.GetType()
			formattedValues := []interface{}{}
			for _, val := range values.GetValues() {
				switch segmenterType {
				case _segmenters.SegmenterValueType_STRING:
					// Quote the string
					formattedValues = append(formattedValues, fmt.Sprintf("%q", val.GetString_()))
				case _segmenters.SegmenterValueType_BOOL:
					formattedValues = append(formattedValues, val.GetBool())
				case _segmenters.SegmenterValueType_INTEGER:
					formattedValues = append(formattedValues, val.GetInteger())
				case _segmenters.SegmenterValueType_REAL:
					formattedValues = append(formattedValues, val.GetReal())
				}
			}
			formattedMap[segmenterName] = &formattedValues
		}
	}
	return formattedMap, nil
}

func (svc *segmenterService) ValidateExperimentSegment(userSegmenters []string, expSegment models.ExperimentSegmentRaw) error {
	inputSegmenters, err := segmenters.ToProtoValues(expSegment, svc.GetSegmenterTypes())
	if err != nil {
		return err
	}
	// For each user segmenter, check the detailed segmenter config
	for _, s := range userSegmenters {
		segmenter, err := svc.getSegmenter(s)
		if err != nil {
			return err
		}
		err = segmenter.ValidateSegmenterAndConstraints(inputSegmenters)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateSegmentOrthogonality checks that the given experiment's segment does not overlap
// with other given experiments. A segment is considered to overlap with another if each
// segmenter has one or more common values. The reverse makes them orthogonal - at least
// one segmenter has no common values.
func (svc *segmenterService) ValidateSegmentOrthogonality(
	userSegmenters []string,
	expSegment models.ExperimentSegmentRaw,
	allExps []models.Experiment,
) error {
	expSegmentFormatted, err := svc.GetFormattedSegmenters(expSegment)
	if err != nil {
		return err
	}

	for _, exp := range allExps {
		otherSegmentFormatted, err := svc.GetFormattedSegmenters(exp.Segment.ToRawSchema(svc.GetSegmenterTypes()))
		if err != nil {
			return err
		}

		// Check that the current experiment segment and the other are orthogonal
		segmentsOverlap := true
		for _, name := range userSegmenters {
			isCurrValEmpty, isOtherValEmpty := false, false
			currValues, ok := expSegmentFormatted[name]
			if !ok || currValues == nil || len(*currValues) == 0 {
				isCurrValEmpty = true
			}
			otherValues, ok := otherSegmentFormatted[name]
			if !ok || otherValues == nil || len(*otherValues) == 0 {
				isOtherValEmpty = true
			}

			// If both values non-empty, check overlap.
			// If only one of the values is empty, we can skip further checks.
			// If both empty, nothing to do.
			if !isCurrValEmpty && !isOtherValEmpty {
				currentSet := set.New(*currValues...)
				otherSet := set.New(*otherValues...)
				if currentSet.Intersection(otherSet).Len() == 0 {
					// At least one segmenter does not overlap, we can terminate the check for
					// this other experiment.
					segmentsOverlap = false
					break
				}
			} else if !isCurrValEmpty || !isOtherValEmpty {
				segmentsOverlap = false
				break
			}
		}

		if segmentsOverlap {
			return fmt.Errorf("Segment Orthogonality check failed against experiment ID %d", exp.ID)
		}
	}

	return nil
}

func (svc *segmenterService) ValidateRequiredSegmenters(segmenterNames []string) error {
	providedSegmenterNames := set.New()
	for _, segmenterName := range segmenterNames {
		providedSegmenterNames.Insert(segmenterName)
	}

	// Validate required segmenters are selected
	for k, v := range svc.segmenters {
		config, err := v.GetConfiguration()
		if err != nil {
			return err
		}
		if config.Required {
			if !providedSegmenterNames.Has(k) {
				return fmt.Errorf("segmenter %s is a required segmenter that must be chosen", k)
			}
		}
	}
	return nil
}

func (svc *segmenterService) ValidatePrereqSegmenters(segmenterNames []string) error {
	providedSegmenterNames := set.New()
	for _, segmenterName := range segmenterNames {
		providedSegmenterNames.Insert(segmenterName)
	}

	// Validate pre-requisite segmenters are selected
	for _, segmenterName := range segmenterNames {
		segmenter, err := svc.getSegmenter(segmenterName)
		if err != nil {
			return err
		}
		config, err := segmenter.GetConfiguration()
		if err != nil {
			return err
		}
		for _, constraint := range config.Constraints {
			prereqs := constraint.GetPreRequisites()
			for _, prereq := range prereqs {
				if providedSegmenterNames.Has(prereq.SegmenterName) {
					continue
				} else {
					return fmt.Errorf("segmenter %s requires %s to also be chosen", segmenterName, prereq.SegmenterName)
				}
			}
		}
	}
	return nil
}

func (svc *segmenterService) ValidateExperimentVariables(projectSegmenters models.ProjectSegmenters) error {

	if len(projectSegmenters.Names) != len(projectSegmenters.Variables) {
		return fmt.Errorf("len of project segmenters does not match mapping of experiment variables")
	}
	for _, segmentersName := range projectSegmenters.Names {
		providedVariables, ok := projectSegmenters.Variables[segmentersName]
		if !ok {
			return fmt.Errorf("project segmenters does not match mapping of experiment variables")
		}
		segmenter, err := svc.getSegmenter(segmentersName)
		if err != nil {
			return err
		}
		config, err := segmenter.GetConfiguration()
		if err != nil {
			return err
		}
		// flag to check if segmenter has matching variables as per segmenters setting
		isValid := false
		treatmentRequestFields := config.TreatmentRequestFields.GetValues()
		less := func(a, b string) bool { return a < b }
		for _, supportedVariables := range treatmentRequestFields {
			if isValid {
				break
			}
			// sorts and compare the slice if they are equal. Returns "" if equal.\
			isValid = cmp.Diff(supportedVariables.Value, providedVariables, cmpopts.SortSlices(less)) == ""
		}
		if !isValid {
			return fmt.Errorf("segmenter (%s) does not have valid experiment variable(s) provided", segmentersName)
		}
	}
	return nil
}

func (svc *segmenterService) getSegmenter(name string) (segmenters.Segmenter, error) {
	segmenter, ok := svc.segmenters[name]
	if !ok {
		return nil, fmt.Errorf("Unknown segmenter %s", name)
	}
	return segmenter, nil
}
