package models

import (
	"testing"
	"time"

	"github.com/gojek/xp/common/segmenters"
	"github.com/stretchr/testify/assert"
)

var testDescription1 = "test-custom-segmenter: string"

var testSegmenters = []CustomSegmenter{
	{
		ProjectID:   ID(1),
		Name:        "test-custom-segmenter-1",
		Type:        SegmenterValueTypeString,
		Description: &testDescription1,
		Required:    true,
		MultiValued: true,
		Options: &Options{
			"option_1": "string_1",
			"option_2": "string_2",
		},
		Model: Model{
			CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
			UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
		},
	},
}

func TestGetName(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		expectedName    string
	}{
		"success | test-custom-segmenter-1": {
			testSegmenters[0],
			"test-custom-segmenter-1",
		},
	}

	for _, data := range tests {
		assert.Equal(t, data.expectedName, data.customSegmenter.GetName())
	}
}

func TestGetType(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		expectedName    segmenters.SegmenterValueType
	}{
		"success | test-custom-segmenter-1 is of type STRING": {
			testSegmenters[0],
			segmenters.SegmenterValueType_STRING,
		},
	}

	for _, data := range tests {
		assert.Equal(t, data.expectedName, data.customSegmenter.GetType())
	}
}

func TestGetConfiguration(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		expectedConfig  *segmenters.SegmenterConfiguration
		errString       string
	}{
		"failure | unable to retrieve base segmenter due to invalid type": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueType("INVALID_TYPE"),
			},
			errString: "error getting a segmenter value type corresponding to: INVALID_TYPE",
		},
		"success | test-custom-segmenter-1 is of type STRING": {
			customSegmenter: testSegmenters[0],
			expectedConfig: &segmenters.SegmenterConfiguration{
				Name: "test-custom-segmenter-1",
				Type: 0,
				Options: map[string]*segmenters.SegmenterValue{
					"option_1": {Value: &segmenters.SegmenterValue_String_{String_: "string_1"}},
					"option_2": {Value: &segmenters.SegmenterValue_String_{String_: "string_2"}},
				},
				MultiValued: true,
				TreatmentRequestFields: &segmenters.ListExperimentVariables{
					Values: []*segmenters.ExperimentVariables{
						{
							Value: []string{"test-custom-segmenter-1"},
						},
					},
				},
				Constraints: []*segmenters.Constraint(nil),
				Required:    true,
				Description: "test-custom-segmenter: string",
			},
		},
	}

	for _, data := range tests {
		actual, err := data.customSegmenter.GetConfiguration()
		if data.errString == "" {
			assert.NoError(t, err)
			assert.Equal(t, data.expectedConfig, actual)
		} else {
			assert.EqualError(t, err, data.errString)
		}
	}
}

func TestGetExperimentVariables(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		expected        *segmenters.ListExperimentVariables
	}{
		"success | test-custom-segmenter-1 is of type STRING": {
			customSegmenter: testSegmenters[0],
			expected: &segmenters.ListExperimentVariables{
				Values: []*segmenters.ExperimentVariables{
					{
						Value: []string{"test-custom-segmenter-1"},
					},
				},
			},
		},
	}

	for _, data := range tests {
		actual := data.customSegmenter.GetExperimentVariables()
		assert.Equal(t, data.expected, actual)
	}
}

func TestIsValidType(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		values          []*segmenters.SegmenterValue
		success         bool
	}{
		"failure | invalid type; need bool but given real": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "dummy-segmenter",
				Type:      SegmenterValueTypeBool,
			},
			values: []*segmenters.SegmenterValue{
				{Value: &segmenters.SegmenterValue_Real{Real: 0.5}},
			},
		},
		"failure | invalid type; need real but given integer": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "dummy-segmenter",
				Type:      SegmenterValueTypeReal,
			},
			values: []*segmenters.SegmenterValue{
				{Value: &segmenters.SegmenterValue_Integer{Integer: 10}},
			},
		},
		"failure | invalid type; need real but given bool": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "dummy-segmenter",
				Type:      SegmenterValueTypeReal,
			},
			values: []*segmenters.SegmenterValue{
				{Value: &segmenters.SegmenterValue_Bool{Bool: false}},
			},
		},
		"failure | invalid type; need integer but given string": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "dummy-segmenter",
				Type:      SegmenterValueTypeInteger,
			},
			values: []*segmenters.SegmenterValue{
				{Value: &segmenters.SegmenterValue_String_{String_: "0"}},
			},
		},
		"failure | mixed types": {
			customSegmenter: testSegmenters[0],
			values: []*segmenters.SegmenterValue{
				{Value: &segmenters.SegmenterValue_Integer{Integer: 10}},
				{Value: &segmenters.SegmenterValue_Real{Real: 0.5}},
			},
		},
		"success | empty list": {
			customSegmenter: testSegmenters[0],
			values:          []*segmenters.SegmenterValue{},
			success:         true,
		},
		"success | valid values": {
			customSegmenter: testSegmenters[0],
			values: []*segmenters.SegmenterValue{
				{Value: &segmenters.SegmenterValue_String_{String_: "100"}},
			},
			success: true,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			success := data.customSegmenter.IsValidType(data.values)
			assert.Equal(t, data.success, success)
		})
	}
}

func TestValidateSegmenterAndConstraints(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		values          map[string]*segmenters.ListSegmenterValue
		errString       string
	}{
		"failure | unable to retrieve base segmenter due to invalid type": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueType("INVALID_TYPE"),
			},
			values:    map[string]*segmenters.ListSegmenterValue{},
			errString: "error getting a segmenter value type corresponding to: INVALID_TYPE",
		},
		"success": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "valid-segmenter",
				Type:      SegmenterValueTypeInteger,
			},
			values: map[string]*segmenters.ListSegmenterValue{},
		},
	}

	for _, data := range tests {
		err := data.customSegmenter.ValidateSegmenterAndConstraints(data.values)
		if data.errString == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, data.errString)
		}
	}
}

func TestValidatePreRequisiteSegmenters(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		errString       string
	}{
		"failure | prerequisites contain the name of the actual segmenter": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						PreRequisites: []PreRequisite{
							{
								SegmenterName:   "invalid-segmenter",
								SegmenterValues: []interface{}{2},
							},
						},
						AllowedValues: []interface{}{2},
					},
				},
			},
			errString: "segmenter invalid-segmenter cannot be a prerequisite of itself",
		},
		"success | no constraints": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "valid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
			},
		},
		"success": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "valid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						PreRequisites: []PreRequisite{
							{
								SegmenterName:   "some-other-segmenter",
								SegmenterValues: []interface{}{"VALID"},
							},
						},
						AllowedValues: []interface{}{2},
					},
				},
			},
		},
	}

	for _, data := range tests {
		err := data.customSegmenter.ValidateSegmenterNotPreRequisiteOfItself()
		if data.errString == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, data.errString)
		}
	}
}

func TestValidateConstraintValues(t *testing.T) {
	tests := map[string]struct {
		customSegmenter CustomSegmenter
		errString       string
	}{
		"failure | constraint contains empty allowed values array": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						AllowedValues: []interface{}{},
					},
				},
			},
			errString: "allowed values cannot be an empty array",
		},
		"failure | unable to find constraint allowed values in options": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						AllowedValues: []interface{}{3},
					},
				},
			},
			errString: "allowed value %!s(int=3) is not specified within segmenter options",
		},
		"failure | unable to find constraint options values in options": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						AllowedValues: []interface{}{1},
						Options: &Options{
							"option_1_new": 3,
						},
					},
				},
			},
			errString: "segmenter name option_1_new with value %!s(int=3) is not specified within segmenter options",
		},
		"failure | constraint options values are not unique": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						AllowedValues: []interface{}{1},
						Options: &Options{
							"option_1_new": 1,
							"option_2_new": 1,
						},
					},
				},
			},
			errString: "options mappings cannot contain different names for the same value %!s(int=1)",
		},
		"failure | constraint contains has allowed values that do not correspond to the constraint option values": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "invalid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						AllowedValues: []interface{}{1},
						Options: &Options{
							"option_1_new": 1,
							"option_2_new": 2,
						},
					},
				},
			},
			errString: "segmenter values in constraint options do not match those in the allowed values",
		},
		"success | no options": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "valid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Constraints: &Constraints{
					{
						AllowedValues: []interface{}{3},
					},
				},
			},
		},
		"success | no constraints": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "valid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
			},
		},
		"success": {
			customSegmenter: CustomSegmenter{
				ProjectID: ID(1),
				Name:      "valid-segmenter",
				Type:      SegmenterValueTypeInteger,
				Options: &Options{
					"option_1": 1,
					"option_2": 2,
				},
				Constraints: &Constraints{
					{
						AllowedValues: []interface{}{1},
					},
				},
			},
		},
	}

	for _, data := range tests {
		err := data.customSegmenter.ValidateConstraintValues()
		if data.errString == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, data.errString)
		}
	}
}

func TestConvertToTypedSegmenterValue(t *testing.T) {
	tests := map[string]struct {
		segmenterValue interface{}
		typeName       SegmenterValueType
		expected       interface{}
		errString      string
	}{
		"failure | invalid type; unrecognised segmenter value type": {
			segmenterValue: interface{}(1),
			typeName:       SegmenterValueType("INVALID_TYPE"),
			errString:      "segmenter value type not recognised: INVALID_TYPE",
		},
		"failure | invalid type; need bool but given real": {
			segmenterValue: interface{}(1.1),
			typeName:       SegmenterValueTypeBool,
			errString:      "received wrong type of segmenter value; %!s(float64=1.1) expects type bool",
		},
		"failure | invalid type; need real but given bool": {
			segmenterValue: interface{}(false),
			typeName:       SegmenterValueTypeReal,
			errString:      "received wrong type of segmenter value; %!s(bool=false) expects type real",
		},
		"failure | invalid type; need integer but given string": {
			segmenterValue: interface{}("invalid"),
			typeName:       SegmenterValueTypeInteger,
			errString:      "received wrong type of segmenter value; invalid expects type integer",
		},
		"failure | invalid type; need string but given integer": {
			segmenterValue: interface{}(1),
			typeName:       SegmenterValueTypeString,
			errString:      "received wrong type of segmenter value; %!s(int=1) expects type string",
		},
		"success | bool": {
			segmenterValue: interface{}(true),
			typeName:       SegmenterValueTypeBool,
			expected:       true,
		},
		"success | real": {
			segmenterValue: interface{}(1.1),
			typeName:       SegmenterValueTypeReal,
			expected:       1.1,
		},
		"success | integer": {
			segmenterValue: interface{}(float64(1)),
			typeName:       SegmenterValueTypeInteger,
			expected:       int64(1),
		},
		"success | string": {
			segmenterValue: interface{}("valid"),
			typeName:       SegmenterValueTypeString,
			expected:       "valid",
		},
	}

	for _, data := range tests {
		actual, err := convertToTypedSegmenterValue(data.segmenterValue, data.typeName)
		if data.errString == "" {
			assert.NoError(t, err)
			assert.Equal(t, data.expected, actual)
		} else {
			assert.EqualError(t, err, data.errString)
		}
	}
}

func TestConvertSegmenterValueToString(t *testing.T) {
	tests := map[string]struct {
		segmenterValue interface{}
		typeName       SegmenterValueType
		expected       string
		errString      string
	}{
		"failure | invalid type; unrecognised segmenter value type": {
			segmenterValue: 1,
			typeName:       SegmenterValueType("INVALID_TYPE"),
			errString:      "segmenter value type not recognised: INVALID_TYPE",
		},
		"failure | invalid type; need bool but given real": {
			segmenterValue: 1.1,
			typeName:       SegmenterValueTypeBool,
			errString:      "received wrong type of segmenter value; %!s(float64=1.1) expects type bool",
		},
		"failure | invalid type; need real but given bool": {
			segmenterValue: false,
			typeName:       SegmenterValueTypeReal,
			errString:      "received wrong type of segmenter value; %!s(bool=false) expects type real",
		},
		"failure | invalid type; need integer but given string": {
			segmenterValue: "invalid",
			typeName:       SegmenterValueTypeInteger,
			errString:      "received wrong type of segmenter value; invalid expects type integer",
		},
		"failure | invalid type; need string but given integer": {
			segmenterValue: 1,
			typeName:       SegmenterValueTypeString,
			errString:      "received wrong type of segmenter value; %!s(int=1) expects type string",
		},
		"success | bool": {
			segmenterValue: true,
			typeName:       SegmenterValueTypeBool,
			expected:       "true",
		},
		"success | real": {
			segmenterValue: 1.1,
			typeName:       SegmenterValueTypeReal,
			expected:       "1.1",
		},
		"success | integer": {
			segmenterValue: int64(1),
			typeName:       SegmenterValueTypeInteger,
			expected:       "1",
		},
		"success | string": {
			segmenterValue: "valid",
			typeName:       SegmenterValueTypeString,
			expected:       "valid",
		},
	}

	for _, data := range tests {
		actual, err := convertSegmenterValueToString(data.segmenterValue, data.typeName)
		if data.errString == "" {
			assert.NoError(t, err)
			assert.Equal(t, data.expected, actual)
		} else {
			assert.EqualError(t, err, data.errString)
		}
	}
}

func TestConvertSegmenterValueFromString(t *testing.T) {
	tests := map[string]struct {
		segmenterValue interface{}
		typeName       SegmenterValueType
		expected       interface{}
		errString      string
	}{
		"failure | input is not of string type": {
			segmenterValue: false,
			typeName:       SegmenterValueTypeBool,
			errString:      "segmenter value is not a string: %!s(bool=false)",
		},
		"failure | invalid type; unrecognised segmenter value type": {
			segmenterValue: "1",
			typeName:       SegmenterValueType("INVALID_TYPE"),
			errString:      "segmenter value type not recognised: INVALID_TYPE",
		},
		"failure | invalid type; need bool but given real": {
			segmenterValue: "1.1",
			typeName:       SegmenterValueTypeBool,
			errString:      "received wrong type of segmenter value; 1.1 expects type bool",
		},
		"failure | invalid type; need real but given bool": {
			segmenterValue: "false",
			typeName:       SegmenterValueTypeReal,
			errString:      "received wrong type of segmenter value; false expects type real",
		},
		"failure | invalid type; need integer but given string": {
			segmenterValue: "invalid",
			typeName:       SegmenterValueTypeInteger,
			errString:      "received wrong type of segmenter value; invalid expects type integer",
		},
		"success | bool": {
			segmenterValue: "true",
			typeName:       SegmenterValueTypeBool,
			expected:       true,
		},
		"success | real": {
			segmenterValue: "1.1",
			typeName:       SegmenterValueTypeReal,
			expected:       1.1,
		},
		"success | integer": {
			segmenterValue: "1",
			typeName:       SegmenterValueTypeInteger,
			expected:       int64(1),
		},
		"success | string": {
			segmenterValue: "valid",
			typeName:       SegmenterValueTypeString,
			expected:       "valid",
		},
	}

	for _, data := range tests {
		actual, err := convertSegmenterValueFromString(data.segmenterValue, data.typeName)
		if data.errString == "" {
			assert.NoError(t, err)
			assert.Equal(t, data.expected, actual)
		} else {
			assert.EqualError(t, err, data.errString)
		}
	}
}
