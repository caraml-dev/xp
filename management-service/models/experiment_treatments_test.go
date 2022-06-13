package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojek/turing-experiments/common/api/schema"
)

var testTreatmentsTraffic30 int32 = 30
var testTreatments = ExperimentTreatments([]ExperimentTreatment{
	{
		Configuration: map[string]interface{}{
			"config-1": "value",
			"config-2": 100.5,
			"config-3": map[string]interface{}{
				"x": "y",
				"z": nil,
			},
			"config-4": []interface{}{true, false, true},
		},
		Name:    "control",
		Traffic: &testTreatmentsTraffic30,
	},
})
var testExperimentTreatments = ExperimentTreatments([]ExperimentTreatment{
	{
		Configuration: map[string]interface{}{
			"config-1": "value",
			"config-2": 100.5,
			"config-3": map[string]interface{}{
				"x": "y",
				"z": nil,
			},
			"config-4": []interface{}{true, false, true},
		},
		Name:    "control",
		Traffic: &testTreatmentsTraffic30,
	},
})

func TestTreatmentsValue(t *testing.T) {
	value, err := testTreatments.Value()
	// Convert to string for comparison
	byteValue, ok := value.([]byte)
	assert.True(t, ok)
	// Validate
	assert.NoError(t, err)
	assert.JSONEq(t, `
	[
		{
			"configuration": {
				"config-1": "value",
				"config-2": 100.5,
				"config-3": {
					"x": "y",
					"z": null
				},
				"config-4": [true, false, true]
			},
			"name": "control",
			"traffic": 30
		}
	]
	`, string(byteValue))
}

func TestTreatmentsScan(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		errString string
		expected  ExperimentTreatments
	}{
		{
			name: "success",
			value: []byte(`
			[
				{
					"configuration": {
						"config-1": "value",
						"config-2": 100.5,
						"config-3": {
							"x": "y",
							"z": null
						},
						"config-4": [true, false, true]
					},
					"name": "control",
					"traffic": 30
				}
			]
			`),
			expected: testTreatments,
		},
		{
			name:      "failure | invalid value",
			value:     100,
			errString: "type assertion to []byte failed",
		},
	}

	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			var treatments ExperimentTreatments
			err := treatments.Scan(data.value)
			if data.errString == "" {
				// Success
				require.NoError(t, err)
				assert.Equal(t, data.expected, treatments)
			} else {
				assert.EqualError(t, err, data.errString)
			}
		})
	}
}

func TestTreatmentsToApiSchema(t *testing.T) {
	assert.Equal(t, []schema.ExperimentTreatment{
		{
			Configuration: map[string]interface{}{
				"config-1": "value",
				"config-2": 100.5,
				"config-3": map[string]interface{}{
					"x": "y",
					"z": nil,
				},
				"config-4": []interface{}{true, false, true},
			},
			Name:    "control",
			Traffic: &testTreatmentsTraffic30,
		},
	}, testExperimentTreatments.ToApiSchema())
}

func TestTreatmentsToProtoSchema(t *testing.T) {
	protoRecord, err := testExperimentTreatments.ToProtoSchema()
	require.NoError(t, err)
	jsonData, err := json.Marshal(protoRecord)
	require.NoError(t, err)
	assert.JSONEq(t, `[
		{
			"config": {
				"config-1": "value",
				"config-2": 100.5,
				"config-3": {
					"x": "y",
					"z": null
				},
				"config-4": [true, false, true]
			},
			"name": "control",
			"traffic": 30
		}
	]`, string(jsonData))
}
