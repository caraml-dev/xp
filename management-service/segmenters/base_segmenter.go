package segmenters

import (
	"fmt"

	"github.com/golang-collections/collections/set"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

type BaseSegmenter struct {
	config *_segmenters.SegmenterConfiguration
}

func NewBaseSegmenter(s *_segmenters.SegmenterConfiguration) *BaseSegmenter {
	return &BaseSegmenter{
		config: s,
	}
}

func (s *BaseSegmenter) GetName() string {
	return s.config.Name
}

func (s *BaseSegmenter) GetType() _segmenters.SegmenterValueType {
	return s.config.Type
}

func (s *BaseSegmenter) GetConfiguration() (*_segmenters.SegmenterConfiguration, error) {
	return s.config, nil
}

func (s *BaseSegmenter) GetExperimentVariables() *_segmenters.ListExperimentVariables {
	return s.config.TreatmentRequestFields
}

// IsValidType checks that the input values for the segmenter are of the configured type
func (s *BaseSegmenter) IsValidType(inputValues []*_segmenters.SegmenterValue) bool {
	valueType := s.GetType()
	for _, val := range inputValues {
		switch valueType {
		case _segmenters.SegmenterValueType_STRING:
			if _, ok := val.GetValue().(*_segmenters.SegmenterValue_String_); !ok {
				return false
			}
		case _segmenters.SegmenterValueType_BOOL:
			if _, ok := val.GetValue().(*_segmenters.SegmenterValue_Bool); !ok {
				return false
			}
		case _segmenters.SegmenterValueType_INTEGER:
			if _, ok := val.GetValue().(*_segmenters.SegmenterValue_Integer); !ok {
				return false
			}
		case _segmenters.SegmenterValueType_REAL:
			if _, ok := val.GetValue().(*_segmenters.SegmenterValue_Real); !ok {
				return false
			}
		}
	}
	return true
}

// ValidateSegmenterAndConstraints validates the current segmenter's values, taking into account the
// configuration of the whole segment.
func (s *BaseSegmenter) ValidateSegmenterAndConstraints(segment map[string]*_segmenters.ListSegmenterValue) error {
	// Get the values for the current segmenter, in the input segment
	listInputValues, ok := segment[s.config.Name]
	if !ok || listInputValues == nil || len(listInputValues.GetValues()) == 0 {
		// Required segmenter should be set. Else, it is an optional segmenter.
		if s.config.Required {
			return fmt.Errorf("Segmenter %s is not set", s.config.Name)
		} else {
			return nil
		}
	}
	inputValues := listInputValues.GetValues()

	// Check if single-valued
	if !s.config.MultiValued && len(inputValues) > 1 {
		return fmt.Errorf("Segmenter %s is configured as single-valued but has multiple input values", s.config.Name)
	}

	// Check value types
	if !s.IsValidType(inputValues) {
		return fmt.Errorf("Segmenter %s has one or more values that do not match the configured type", s.config.Name)
	}

	// Check that the input segment contains valid values for the current segmenter
	if len(s.config.Options) > 0 {
		allValues := []*_segmenters.SegmenterValue{}
		for _, val := range s.config.Options {
			allValues = append(allValues, val)
		}
		if !isValidValues(allValues, inputValues) {
			return fmt.Errorf("Segmenter %s uses one or more invalid values", s.config.Name)
		}
	}

	// Check constraints
	return s.checkConstraints(inputValues, segment)
}

func (s *BaseSegmenter) checkConstraints(
	inputValues []*_segmenters.SegmenterValue,
	segment map[string]*_segmenters.ListSegmenterValue,
) error {
	// For each constraint, check if all pre-requisites are met. For the first constraint
	// whose pre-requisites are satisfied, check that the segmenter's values are in the allowed
	// list.
	for _, constraint := range s.config.Constraints {
		// Check each pre-requisite
		constraintMatch := true
		for _, preReq := range constraint.PreRequisites {
			if preReq.SegmenterValues == nil {
				// This shouldn't happen, but handle it anyway
				return fmt.Errorf("Pre-requisite segmenter %s for %s not configured correctly",
					preReq.SegmenterName, s.config.Name)
			}
			expectedValues := preReq.SegmenterValues.GetValues()
			listSegmenterValues := segment[preReq.SegmenterName]
			if listSegmenterValues == nil {
				// The segmenter is not in use. Thus, the constraint will not be met.
				constraintMatch = false
				break
			} else {
				segmenterValues := listSegmenterValues.GetValues()
				if !isValidValues(expectedValues, segmenterValues) {
					constraintMatch = false
					break
				}
			}
		}

		if constraintMatch {
			// Check that the current segmenter's values match the allowed values
			listAllowedValues := constraint.AllowedValues
			if listAllowedValues == nil {
				// This shouldn't happen, but handle it anyway
				return fmt.Errorf("Constraint for segmenter %s not configured correctly", s.config.Name)
			}
			allowedValues := listAllowedValues.GetValues()
			if !isValidValues(allowedValues, inputValues) {
				return fmt.Errorf("Values for segmenter %s do not satisfy the constraint", s.config.Name)
			}
			break
		}
	}
	return nil
}

// isValidValues checks that the input values are from the list of all segmenter values configured
func isValidValues(allValues []*_segmenters.SegmenterValue, inputValues []*_segmenters.SegmenterValue) bool {
	// Convert data to generic interface list using the string representation for
	// equality comparison
	allList := []interface{}{}
	inputList := []interface{}{}
	for _, val := range allValues {
		if val != nil {
			allList = append(allList, val.String())

		}
	}
	for _, val := range inputValues {
		if val != nil {
			inputList = append(inputList, val.String())
		}
	}

	// Create sets
	allSet := set.New(allList...)
	inputSet := set.New(inputList...)

	// Check that the inputSet is a subset of allSet
	return inputSet.SubsetOf(allSet)
}
