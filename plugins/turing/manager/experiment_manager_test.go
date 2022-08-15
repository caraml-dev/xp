package manager

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/plugins/turing/config"
)

func TestNewExperimentManager(t *testing.T) {
	em := &experimentManager{}
	// Test that the custom experiment manager interface is satisfied
	assert.Implements(t, (*manager.CustomExperimentManager)(nil), em)
}

func TestGetEngineInfo(t *testing.T) {
	em := &experimentManager{
		RemoteUI: config.RemoteUI{
			Name:   "xp",
			URL:    "http://example.com",
			Config: "http://example.com/app.config.js",
		},
	}

	actual, err := em.GetEngineInfo()
	assert.NoError(t, err)
	assert.Equal(t, manager.Engine{
		Name:        "xp",
		DisplayName: "Turing Experiments",
		Type:        manager.CustomExperimentManagerType,
		CustomExperimentManagerConfig: &manager.CustomExperimentManagerConfig{
			RemoteUI: manager.RemoteUI{
				Name:   "xp",
				URL:    "http://example.com",
				Config: "http://example.com/app.config.js",
			},
			ExperimentConfigSchema: xpExperimentConfigSchema,
		},
	}, actual)
}

func TestGetExperimentRunnerConfig(t *testing.T) {
	// Define test cases
	tests := map[string]struct {
		input    json.RawMessage
		expected string
		err      string
	}{
		"failure | bad data": {
			input: json.RawMessage(`[1, 2]`),
			err: strings.Join([]string{"Error creating experiment runner config:",
				"json: cannot unmarshal array into Go value of type config.ExperimentConfig"}, " "),
		},
		"success": {
			input: json.RawMessage(`{
				"project_id": 10,
				"variables": [
					{
						"name": "country",
						"field": "countryID",
						"field_source": "header"
					},
										{
						"name": "geo_area",
						"field": "gArea",
						"field_source": "payload"
					}
				]
			}`),
			expected: `{
				"endpoint": "test-endpoint",
				"project_id": 10,
				"passkey": "test-passkey",
				"timeout": "12s",
				"request_parameters": [
					{
						"parameter": "country",
						"field": "countryID",
						"field_source": "header"
					},
					{
						"parameter": "geo_area",
						"field": "gArea",
						"field_source": "payload"
					}
				]
			}`,
		},
	}

	// Patch method to get passkey
	// TODO: Generate mock client and use it here instead of patching
	em := &experimentManager{RunnerDefaults: config.RunnerDefaults{Endpoint: "test-endpoint", Timeout: "12s"}}
	monkey.PatchInstanceMethod(
		reflect.TypeOf(em),
		"GetProject",
		func(em *experimentManager, projectId int) (*schema.ProjectSettings, error) {
			if projectId == 10 {
				return &schema.ProjectSettings{Passkey: "test-passkey"}, nil
			}
			return nil, fmt.Errorf("Unexpected ProjectID: %d", projectId)
		},
	)
	defer monkey.UnpatchAll()

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := em.GetExperimentRunnerConfig(data.input)

			// Validate
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.JSONEq(t, data.expected, string(result))
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateExperimentConfig(t *testing.T) {
	tests := map[string]struct {
		cfg json.RawMessage
		err string
	}{
		"success": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field": "Country",
						"field_source": "header"
					}
				]
			}`),
		},
		"failure | missing values": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field_source": "header"
					}
				]
			}`),
			err: strings.Join([]string{"Key: 'ExperimentConfig.Variables[0].Field' Error:",
				"Field validation for 'Field' failed on the 'required' tag",
			}, ""),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			em := experimentManager{validate: validator.New()}
			err := em.ValidateExperimentConfig(data.cfg)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
