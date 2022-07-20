package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gojek/xp/common/api/schema"
	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/segmenters"
	"github.com/golang-collections/collections/set"
)

// SegmenterValueType represents the possible types that segmenter values can take.
type SegmenterValueType string

const (
	SegmenterValueTypeString  SegmenterValueType = "STRING"
	SegmenterValueTypeBool    SegmenterValueType = "BOOL"
	SegmenterValueTypeInteger SegmenterValueType = "INTEGER"
	SegmenterValueTypeReal    SegmenterValueType = "REAL"
)

type PreRequisite struct {
	// segmenter_name is the name of the free segmenter. This must be single-valued.
	SegmenterName string `json:"segmenter_name"`
	// segmenter_values is the set of values of the pre-requisite segmenter, one
	// of which must be matched.
	SegmenterValues []interface{} `json:"segmenter_values"`
}

func (p *PreRequisite) ToApiSchema() schema.PreRequisite {
	var segmenterValues []schema.SegmenterValues
	for _, segmenter := range p.SegmenterValues {
		segmenterValues = append(segmenterValues, segmenter)
	}
	return schema.PreRequisite{
		SegmenterName:   p.SegmenterName,
		SegmenterValues: segmenterValues,
	}
}

type Constraints []Constraint

type Constraint struct {
	PreRequisites []PreRequisite `json:"pre_requisites"`
	AllowedValues []interface{}  `json:"allowed_values"`
	Options       *Options       `json:"options"`
}

func (c *Constraint) ToApiSchema() schema.Constraint {
	var preRequisites []schema.PreRequisite
	for _, preRequisite := range c.PreRequisites {
		preRequisites = append(preRequisites, preRequisite.ToApiSchema())
	}
	var allowedValues []schema.SegmenterValues
	for _, allowedValue := range c.AllowedValues {
		allowedValues = append(allowedValues, allowedValue)
	}
	var additionalProperties map[string]interface{}
	if c.Options != nil {
		additionalProperties = *c.Options
	}
	return schema.Constraint{
		PreRequisites: preRequisites,
		AllowedValues: allowedValues,
		Options: &schema.SegmenterOptions{
			AdditionalProperties: additionalProperties,
		},
	}
}

func (ct *Constraints) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &ct)
}

func (ct Constraints) Value() (driver.Value, error) {
	return json.Marshal(ct)
}

type Options map[string]interface{}

func (op *Options) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &op)
}

func (op Options) Value() (driver.Value, error) {
	return json.Marshal(op)
}

type CustomSegmenter struct {
	Model

	// ProjectID is the id of the MLP project
	ProjectID ID `json:"project_id" gorm:"primary_key"`

	// Name is the human-readable name of the segmenter. This must be unique across global and project segmenters.
	Name string `json:"name" gorm:"primary_key"`
	// Type of the segmenter's values. All values of a segmenter should be of the same type.
	Type SegmenterValueType `json:"type" validate:"notBlank"`
	// Additional information about segmenter
	Description *string `json:"description"`

	// Required represents whether the segmenter must be chosen in an experiment
	Required bool `json:"required"`
	// MultiValued represents whether multiple values of the segmenter can be
	// chosen in an experiment. Only single-valued segmenters can act as
	// pre-requisites.
	MultiValued bool `json:"multi_valued"`
	// A map of the segmenter values (human-readable name -> internal value)
	Options *Options `json:"options"`
	// Constraints captures an optional list of rules. Each constraint has one or
	// more pre-requisite conditions, which when satisfied, narrows the list of
	// available values for the current segmenter. If none of the constraints are
	// satisfied, all values of the segmenter described by the options field may
	// be applicable.
	Constraints *Constraints `json:"constraints"`
}

// NewCustomSegmenter creates a new CustomSegmenter object and ensures that its segmenter values are all of the
// correct type specified
func NewCustomSegmenter(
	projectId ID,
	name string,
	segmenterType SegmenterValueType,
	description *string,
	required bool,
	multiValued bool,
	options *Options,
	constraints *Constraints,
	segmenterTypes map[string]schema.SegmenterType,
) (*CustomSegmenter, error) {
	newCustomSegmenter := CustomSegmenter{
		ProjectID:   projectId,
		Name:        name,
		Type:        segmenterType,
		Description: description,
		Required:    required,
		MultiValued: multiValued,
		Options:     options,
		Constraints: constraints,
	}
	if err := newCustomSegmenter.ConvertToTypedValues(segmenterTypes); err != nil {
		return nil, err
	}
	if err := validateOptionsHaveUniqueValues(newCustomSegmenter.Options); err != nil {
		return nil, err
	}
	if err := newCustomSegmenter.ValidateSegmenterNotPreRequisiteOfItself(); err != nil {
		return nil, err
	}
	if err := newCustomSegmenter.ValidateConstraintValues(); err != nil {
		return nil, err
	}
	return &newCustomSegmenter, nil
}

// ToApiSchema converts the configured segmenter DB model to a format compatible with the OpenAPI specifications.
func (s *CustomSegmenter) ToApiSchema() schema.Segmenter {
	var constraints []schema.Constraint
	if s.Constraints != nil {
		for _, constraint := range *s.Constraints {
			constraints = append(constraints, constraint.ToApiSchema())
		}
	}

	var additionalProperties map[string]interface{}
	if s.Options != nil {
		additionalProperties = *s.Options
	}

	return schema.Segmenter{
		Name:        s.Name,
		Type:        schema.SegmenterType(strings.ToLower(string(s.Type))),
		Description: s.Description,
		Required:    s.Required,
		MultiValued: s.MultiValued,
		Options: schema.SegmenterOptions{
			AdditionalProperties: additionalProperties,
		},
		Constraints: constraints,
		CreatedAt:   &s.CreatedAt,
		UpdatedAt:   &s.UpdatedAt,
		TreatmentRequestFields: [][]string{
			{s.Name},
		},
	}
}

func (s *CustomSegmenter) GetName() string {
	return s.Name
}

func (s *CustomSegmenter) GetType() _segmenters.SegmenterValueType {
	return _segmenters.SegmenterValueType(_segmenters.SegmenterValueType_value[string(s.Type)])
}

func (s *CustomSegmenter) GetConfiguration() (*_segmenters.SegmenterConfiguration, error) {
	baseSegmenter, err := s.GetBaseSegmenter()
	if err != nil {
		return nil, err
	}
	config, err := baseSegmenter.GetConfiguration()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *CustomSegmenter) GetExperimentVariables() *_segmenters.ListExperimentVariables {
	return &_segmenters.ListExperimentVariables{
		Values: []*_segmenters.ExperimentVariables{
			{Value: []string{s.Name}},
		},
	}
}

func (s *CustomSegmenter) IsValidType(inputValues []*_segmenters.SegmenterValue) bool {
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

func (s *CustomSegmenter) ValidateSegmenterAndConstraints(segment map[string]*_segmenters.ListSegmenterValue) error {
	baseSegmenter, err := s.GetBaseSegmenter()
	if err != nil {
		return err
	}
	return baseSegmenter.ValidateSegmenterAndConstraints(segment)
}

// GetBaseSegmenter returns a BaseSegmenter object (just as how global segmenters are registered) constructed using
// data contained in the custom segmenter
func (s *CustomSegmenter) GetBaseSegmenter() (segmenters.Segmenter, error) {
	segmenterValueType, ok := _segmenters.SegmenterValueType_value[strings.ToUpper(string(s.Type))]
	if !ok {
		return nil, fmt.Errorf("error getting a segmenter value type corresponding to: %s", string(s.Type))
	}

	description := ""
	if s.Description != nil {
		description = *s.Description
	}

	config := _segmenters.SegmenterConfiguration{
		Name: s.Name,
		Type: _segmenters.SegmenterValueType(segmenterValueType),
		TreatmentRequestFields: &_segmenters.ListExperimentVariables{
			Values: []*_segmenters.ExperimentVariables{
				{Value: []string{s.Name}},
			},
		},
		Options:     formatOptions(s.Options),
		MultiValued: s.MultiValued,
		Constraints: formatConstraints(s.Constraints),
		Required:    s.Required,
		Description: description,
	}
	return segmenters.NewBaseSegmenter(&config), nil
}

// validateOptionsHaveUniqueValues checks if all the values in the options argument are unique
func validateOptionsHaveUniqueValues(options *Options) error {
	if options == nil {
		return nil
	}
	valueSet := set.New()

	for _, value := range *options {
		if valueSet.Has(value) {
			return fmt.Errorf("options mappings cannot contain different names for the same value %s", value)
		}
		valueSet.Insert(value)
	}
	return nil
}

// ValidateSegmenterNotPreRequisiteOfItself checks if all the segmenter names of all the prerequisites specified for all the
// constraints in the segmenter do not refer to the actual segmenter's name.
func (s *CustomSegmenter) ValidateSegmenterNotPreRequisiteOfItself() error {
	if s.Constraints == nil {
		return nil
	}

	for _, constraint := range *s.Constraints {
		for _, preRequisite := range constraint.PreRequisites {
			if preRequisite.SegmenterName == s.Name {
				return fmt.Errorf("segmenter %s cannot be a prerequisite of itself", preRequisite.SegmenterName)
			}
		}
	}
	return nil
}

// ValidateConstraintValues checks if all the allowed values and options of all the constraints in the segmenter are
// present within the options specified. If no options or no constraints are specified, this check passes.
func (s *CustomSegmenter) ValidateConstraintValues() error {
	if s.Options == nil || s.Constraints == nil {
		return nil
	}
	var optionsSlice []interface{}
	for _, value := range *s.Options {
		optionsSlice = append(optionsSlice, value)
	}
	segmenterValueSet := set.New(optionsSlice...)

	for _, constraint := range *s.Constraints {
		// checks if all the allowed values are present in the segmenter options
		if len(constraint.AllowedValues) == 0 {
			return fmt.Errorf("allowed values cannot be an empty array")
		}
		for _, allowedValue := range constraint.AllowedValues {
			if !segmenterValueSet.Has(allowedValue) {
				return fmt.Errorf("allowed value %s is not specified within segmenter options", allowedValue)
			}
		}

		// checks if all the constraint options values specified are present in the segmenter options and that they
		// match every value in the allowed values field
		if constraint.Options != nil {
			err := validateOptionsHaveUniqueValues(constraint.Options)
			if err != nil {
				return err
			}

			var constraintOptionsSlice []interface{}
			for constraintName, constraintValue := range *constraint.Options {
				if !segmenterValueSet.Has(constraintValue) {
					return fmt.Errorf("segmenter name %s with value %s is not specified within segmenter options",
						constraintName, constraintValue)
				}
				constraintOptionsSlice = append(constraintOptionsSlice, constraintValue)
			}

			allowedValuesSet := set.New(constraint.AllowedValues...)
			constraintOptionsSet := set.New(constraintOptionsSlice...)

			if constraintOptionsSet.Difference(allowedValuesSet).Len() != 0 ||
				allowedValuesSet.Difference(constraintOptionsSet).Len() != 0 {
				return fmt.Errorf("segmenter values in constraint options do not match those in the allowed values")
			}
		}
	}
	return nil
}

// ConvertToTypedValues converts a CustomSegmenter's segmenter values that are untyped to the type specified in
// the Type field. As this method also indirectly validates the type of each value, it is also used to validate
// unknown segmenter value types (i.e. validate values passed in as user input with respect to the specified type)
func (s *CustomSegmenter) ConvertToTypedValues(segmenterTypes map[string]schema.SegmenterType) error {
	return s.ConvertCustomSegmenterValues(segmenterTypes, convertToTypedSegmenterValue)
}

// ToStorageSchema converts a CustomSegmenter's segmenter value types to strings for storing in the database
func (s *CustomSegmenter) ToStorageSchema(segmenterTypes map[string]schema.SegmenterType) error {
	return s.ConvertCustomSegmenterValues(segmenterTypes, convertSegmenterValueToString)
}

// FromStorageSchema converts a CustomSegmenter's segmenter value types from strings to the type specified
func (s *CustomSegmenter) FromStorageSchema(segmenterTypes map[string]schema.SegmenterType) error {
	return s.ConvertCustomSegmenterValues(segmenterTypes, convertSegmenterValueFromString)
}

func (s *CustomSegmenter) ConvertCustomSegmenterValues(
	segmenterTypes map[string]schema.SegmenterType,
	conversionFunction func(interface{}, SegmenterValueType) (interface{}, error),
) error {
	var err error
	if s.Options != nil {
		var newOptions Options
		newOptions, err = convertOptions(*s.Options, s.Type, conversionFunction)
		if err != nil {
			return err
		}
		s.Options = &newOptions
	}

	if s.Constraints != nil {
		newConstraints := Constraints{}
		for _, constraint := range *s.Constraints {
			newConstraint := Constraint{}
			newConstraint.PreRequisites, err = convertToTypedPreRequisites(constraint.PreRequisites, segmenterTypes,
				conversionFunction)
			if err != nil {
				return err
			}
			newConstraint.AllowedValues, err = convertSegmenterValues(constraint.AllowedValues, s.Type, conversionFunction)
			if err != nil {
				return err
			}
			if constraint.Options != nil {
				newConstraint.Options = &Options{}
				*newConstraint.Options, err = convertOptions(*constraint.Options, s.Type, conversionFunction)
				if err != nil {
					return err
				}
			}
			newConstraints = append(newConstraints, newConstraint)
		}
		s.Constraints = &newConstraints
	}
	return nil
}

func convertToTypedPreRequisites(
	preRequisites []PreRequisite,
	segmenterTypes map[string]schema.SegmenterType,
	conversionFunction func(interface{}, SegmenterValueType) (interface{}, error),
) ([]PreRequisite, error) {
	var newPreRequisites []PreRequisite
	for _, preRequisite := range preRequisites {
		preRequisiteType, ok := segmenterTypes[preRequisite.SegmenterName]
		if !ok {
			return nil, fmt.Errorf("segmenter type not found for pre-requisite segmenter: %s", preRequisite.SegmenterName)
		}
		newVal, err := convertSegmenterValues(
			preRequisite.SegmenterValues,
			SegmenterValueType(strings.ToUpper(string(preRequisiteType))),
			conversionFunction,
		)
		if err != nil {
			return nil, err
		}
		newPreRequisites = append(
			newPreRequisites,
			PreRequisite{
				SegmenterName:   preRequisite.SegmenterName,
				SegmenterValues: newVal,
			},
		)
	}
	return newPreRequisites, nil
}

func convertOptions(
	options map[string]interface{},
	typeName SegmenterValueType,
	conversionFunction func(interface{}, SegmenterValueType) (interface{}, error),
) (map[string]interface{}, error) {
	convertedOptions := make(map[string]interface{})
	for key, val := range options {
		newVal, err := conversionFunction(val, typeName)
		if err != nil {
			return nil, err
		}
		convertedOptions[key] = newVal
	}
	return convertedOptions, nil
}

func convertSegmenterValues(
	segmenterValues []interface{},
	typeName SegmenterValueType,
	conversionFunction func(interface{}, SegmenterValueType) (interface{}, error),
) ([]interface{}, error) {
	var newSegmenterValues []interface{}
	for _, segmenterValue := range segmenterValues {
		newVal, err := conversionFunction(segmenterValue, typeName)
		if err != nil {
			return nil, err
		}
		newSegmenterValues = append(newSegmenterValues, newVal)
	}
	return newSegmenterValues, nil
}

func convertToTypedSegmenterValue(segmenterValue interface{}, typeName SegmenterValueType) (interface{}, error) {
	errTmpl := fmt.Sprintf("received wrong type of segmenter value; %s expects type", segmenterValue)
	switch typeName {
	case SegmenterValueTypeString:
		stringVal, ok := segmenterValue.(string)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeString)
		}
		return stringVal, nil
	case SegmenterValueTypeInteger:
		// uses float64 conversion as golang json conversion reads untyped numbers as float by default; the check below
		// ensures that val is minimally a number-like variable
		floatVal, ok := segmenterValue.(float64)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeInteger)
		}
		return int64(floatVal), nil
	case SegmenterValueTypeReal:
		floatVal, ok := segmenterValue.(float64)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeReal)
		}
		return floatVal, nil
	case SegmenterValueTypeBool:
		boolVal, ok := segmenterValue.(bool)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeBool)
		}
		return boolVal, nil
	default:
		return nil, fmt.Errorf("segmenter value type not recognised: %s", typeName)
	}
}

func convertSegmenterValueToString(segmenterValue interface{}, typeName SegmenterValueType) (interface{}, error) {
	errTmpl := fmt.Sprintf("received wrong type of segmenter value; %s expects type", segmenterValue)
	switch typeName {
	case SegmenterValueTypeString:
		stringVal, ok := segmenterValue.(string)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeString)
		}
		return stringVal, nil
	case SegmenterValueTypeInteger:
		intVal, ok := segmenterValue.(int64)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeInteger)
		}
		return strconv.Itoa(int(intVal)), nil
	case SegmenterValueTypeReal:
		floatVal, ok := segmenterValue.(float64)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeReal)
		}
		return strconv.FormatFloat(floatVal, 'f', -1, 64), nil
	case SegmenterValueTypeBool:
		boolVal, ok := segmenterValue.(bool)
		if !ok {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeBool)
		}
		return strconv.FormatBool(boolVal), nil
	default:
		return nil, fmt.Errorf("segmenter value type not recognised: %s", typeName)
	}
}

func convertSegmenterValueFromString(segmenterValue interface{}, typeName SegmenterValueType) (interface{}, error) {
	stringVal, ok := segmenterValue.(string)
	if !ok {
		return nil, fmt.Errorf("segmenter value is not a string: %s", segmenterValue)
	}
	errTmpl := fmt.Sprintf("received wrong type of segmenter value; %s expects type", segmenterValue)
	switch typeName {
	case SegmenterValueTypeString:
		return stringVal, nil
	case SegmenterValueTypeInteger:
		intVal, err := strconv.ParseInt(stringVal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeInteger)
		}
		return intVal, nil
	case SegmenterValueTypeReal:
		floatVal, err := strconv.ParseFloat(stringVal, 64)
		if err != nil {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeReal)
		}
		return floatVal, nil
	case SegmenterValueTypeBool:
		boolVal, err := strconv.ParseBool(stringVal)
		if err != nil {
			return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeBool)
		}
		return boolVal, nil
	default:
		return nil, fmt.Errorf("segmenter value type not recognised: %s", typeName)
	}
}

// formatConstraints parses constraints into a format suitable for constructing a SegmenterConfiguration object
func formatConstraints(constraints *Constraints) []*_segmenters.Constraint {
	if constraints == nil {
		return nil
	}
	var configConstraints []*_segmenters.Constraint
	for _, constraint := range *constraints {
		configConstraints = append(
			configConstraints,
			&_segmenters.Constraint{
				PreRequisites: formatPreRequisites(constraint.PreRequisites),
				AllowedValues: formatSegmenterValues(constraint.AllowedValues),
				Options:       formatOptions(constraint.Options),
			},
		)
	}
	return configConstraints
}

// formatPreRequisites parses prerequisites into a format suitable for constructing a SegmenterConfiguration object
func formatPreRequisites(preRequisites []PreRequisite) []*_segmenters.PreRequisite {
	var configPreRequisites []*_segmenters.PreRequisite
	for _, preRequisite := range preRequisites {
		configPreRequisites = append(
			configPreRequisites,
			&_segmenters.PreRequisite{
				SegmenterName:   preRequisite.SegmenterName,
				SegmenterValues: formatSegmenterValues(preRequisite.SegmenterValues),
			},
		)
	}
	return configPreRequisites
}

// formatOptions parses arbitrary options into a format suitable for constructing a SegmenterConfiguration object
func formatOptions(options *Options) map[string]*_segmenters.SegmenterValue {
	if options == nil {
		return nil
	}
	configOptions := make(map[string]*_segmenters.SegmenterValue)
	for key, option := range *options {
		configOptions[key] = formatSegmenterValue(option)
	}
	return configOptions
}

// formatSegmenterValues parses arbitrary segmenter values into a format suitable for constructing a
// SegmenterConfiguration object
func formatSegmenterValues(allowedValues []interface{}) *_segmenters.ListSegmenterValue {
	var configAllowedValues []*_segmenters.SegmenterValue
	for _, allowedValue := range allowedValues {
		configAllowedValues = append(configAllowedValues, formatSegmenterValue(allowedValue))
	}
	return &_segmenters.ListSegmenterValue{
		Values: configAllowedValues,
	}
}

// formatSegmenterValue parses an arbitrary segmenter value into a format suitable for constructing a
// SegmenterConfiguration object
func formatSegmenterValue(segmenterValue interface{}) *_segmenters.SegmenterValue {
	switch (segmenterValue).(type) {
	case string:
		return &_segmenters.SegmenterValue{
			Value: &_segmenters.SegmenterValue_String_{String_: (segmenterValue).(string)},
		}
	case bool:
		return &_segmenters.SegmenterValue{
			Value: &_segmenters.SegmenterValue_Bool{Bool: (segmenterValue).(bool)},
		}
	case int64:
		return &_segmenters.SegmenterValue{
			Value: &_segmenters.SegmenterValue_Integer{Integer: (segmenterValue).(int64)},
		}
	case float64:
		return &_segmenters.SegmenterValue{
			Value: &_segmenters.SegmenterValue_Real{Real: (segmenterValue).(float64)},
		}
	default:
		return nil
	}
}
