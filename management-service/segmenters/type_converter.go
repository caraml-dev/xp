package segmenters

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gojek/xp/common/api/schema"
	_segmenters "github.com/gojek/xp/common/segmenters"
	_utils "github.com/gojek/xp/common/utils"
)

func ToProtoValues(
	segments map[string]interface{},
	segmentersType map[string]schema.SegmenterType,
) (map[string]*_segmenters.ListSegmenterValue, error) {
	protoSegments := make(map[string]*_segmenters.ListSegmenterValue)
	for key, vals := range segments {
		errTmpl := fmt.Sprintf("received wrong type of segmenter value; %s expects type", key)
		values := vals.([]interface{})
		if len(values) > 0 {
			switch segmentersType[key] {
			case schema.SegmenterTypeString:
				strVals := []string{}
				for _, val := range values {
					strVal, ok := val.(string)
					if !ok {
						return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeString)
					}
					strVals = append(strVals, strVal)
				}
				protoSegments[key] = _utils.StringSliceToListSegmenterValue(&strVals)
			case schema.SegmenterTypeInteger:
				intVals := []int64{}
				for _, val := range values {
					floatVal, ok := val.(float64)
					if !ok {
						return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeInteger)
					}
					intVals = append(intVals, int64(floatVal))
				}
				protoSegments[key] = _utils.Int64ListToListSegmenterValue(&intVals)
			case schema.SegmenterTypeReal:
				floatVals := []float64{}
				for _, val := range values {
					floatVal, ok := val.(float64)
					if !ok {
						return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeReal)
					}
					floatVals = append(floatVals, floatVal)
				}
				protoSegments[key] = _utils.FloatListToListSegmenterValue(&floatVals)
			case schema.SegmenterTypeBool:
				boolVals := []bool{}
				for _, val := range values {
					boolVal, ok := val.(bool)
					if !ok {
						return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeBool)
					}
					boolVals = append(boolVals, boolVal)
				}
				protoSegments[key] = _utils.BoolSliceToListSegmenterValue(&boolVals)
			}
		}
	}

	return protoSegments, nil
}

func ProtobufSegmenterConfigToOpenAPISegmenterConfig(segmenterConfiguration *_segmenters.SegmenterConfiguration) (*schema.Segmenter, error) {
	formatOptions := func(opts map[string]*_segmenters.SegmenterValue) (schema.SegmenterOptions, error) {
		segmenterOpts := schema.SegmenterOptions{AdditionalProperties: map[string]interface{}{}}
		for k, v := range opts {
			val := typeToVal(v)
			if val == nil {
				return segmenterOpts, errors.New("segmenters: invalid type conversion for options")
			}
			segmenterOpts.AdditionalProperties[k] = val
		}
		return segmenterOpts, nil
	}
	segmenterOptions, err := formatOptions(segmenterConfiguration.GetOptions())
	if err != nil {
		return nil, err
	}

	// Format Type
	segmenterType := schema.SegmenterType(strings.ToLower(segmenterConfiguration.GetType().String()))

	// Format Constraints
	var segmenterConstraints []schema.Constraint
	for _, val := range segmenterConfiguration.GetConstraints() {
		var allowedValues []schema.SegmenterValues
		for _, allowedVal := range val.GetAllowedValues().GetValues() {
			val := typeToVal(allowedVal)
			if val == nil {
				return nil, errors.New("segmenters: invalid type conversion for allowed_values")
			}
			allowedValues = append(allowedValues, schema.SegmenterValues(val))
		}

		var prereqs []schema.PreRequisite
		for _, prereq := range val.GetPreRequisites() {
			var segmenterValues []schema.SegmenterValues
			for _, segmenterVal := range prereq.GetSegmenterValues().GetValues() {
				val := typeToVal(segmenterVal)
				if val == nil {
					return nil, errors.New("segmenters: invalid type conversion for pre_requisites")
				}
				segmenterValues = append(segmenterValues, schema.SegmenterValues(val))
			}

			prereqs = append(prereqs, schema.PreRequisite{
				SegmenterName:   prereq.GetSegmenterName(),
				SegmenterValues: segmenterValues,
			})
		}

		constraint := schema.Constraint{
			AllowedValues: allowedValues,
			PreRequisites: prereqs,
		}

		// Format constraint-specific options if exists
		if len(val.Options) > 0 {
			opts, err := formatOptions(val.Options)
			if err != nil {
				return nil, err
			}
			constraint.Options = &opts
		}

		segmenterConstraints = append(segmenterConstraints, constraint)
	}

	// Format Options
	segmenterDescription := segmenterConfiguration.GetDescription()

	var segmenterTreatmentRequestFields [][]string
	// TreatmentRequestFields is an array of array holding possible combination of variables that segmenter can derive from
	listExperimentVariables := segmenterConfiguration.GetTreatmentRequestFields()
	for _, experimentVariables := range listExperimentVariables.Values {
		segmenterTreatmentRequestFields = append(segmenterTreatmentRequestFields, experimentVariables.Value)
	}

	modelConfig := &schema.Segmenter{
		Name:                   segmenterConfiguration.GetName(),
		MultiValued:            segmenterConfiguration.GetMultiValued(),
		Constraints:            segmenterConstraints,
		Options:                segmenterOptions,
		TreatmentRequestFields: segmenterTreatmentRequestFields,
		Type:                   segmenterType,
		Required:               segmenterConfiguration.GetRequired(),
		Description:            &segmenterDescription,
	}

	return modelConfig, nil
}

func typeToVal(value *_segmenters.SegmenterValue) interface{} {
	strType := strings.Split(value.String(), ":")[0]

	switch strType {
	case "string":
		return value.GetString_()
	case "integer":
		return value.GetInteger()
	case "real":
		return value.GetReal()
	case "bool":
		return value.GetBool()
	}
	return nil
}
