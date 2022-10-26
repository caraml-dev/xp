package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/plugins/turing/config"
)

func TestNewExperimentManagerImplementsCustomExperimentManagerInterface(t *testing.T) {
	em := &experimentManager{}
	// Test that the custom experiment manager interface is satisfied
	assert.Implements(t, (*manager.CustomExperimentManager)(nil), em)
}

func TestNewExperimentManager(t *testing.T) {
	reset := testSetupEnvForGoogleCredentials(t)
	defer reset()

	// Define test cases
	tests := map[string]struct {
		input json.RawMessage
		err   string
	}{
		"failure | bad data": {
			input: json.RawMessage(`[1, 2]`),
			err: strings.Join([]string{"failed to create XP experiment manager:",
				"json: cannot unmarshal array into Go value of type config.ExperimentManagerConfig"}, " "),
		},
		"success": {
			input: json.RawMessage(`{
				"base_url": "http://xp-management:8080/v1",
				"home_page_url": "/turing/projects/{{projectId}}/experiments",
				"remote_ui": {
					"config": "/xp/app.config.js",
					"name": "xp",
					"url": "/xp/remoteEntry.js"
				},
				"runner_defaults": {
					"endpoint": "http://xp-treatment.global.io/v1",
					"timeout": "5s"
				},
				"treatment_service_plugin_config": {
					"assigned_treatment_logger": {
						"bq_config": {
							"dataset": "xp_dataset",
							"project": "xp_project",
							"table": "xp_table"
						},
						"kind": "bq",
						"queue_length": 100000
					},
					"deployment_config": {
						"environment_type": "dev",
						"max_go_routines": 200
					},
					"management_service": {
						"authorization_enabled": true,
						"url": "http://xp-management.global.io/api/xp/v1"
					},
					"monitoring_config": {
						"kind": "prometheus",
						"metric_labels": [
							"country",
							"service"
						]
					},
					"port": 8080,
					"swagger_config": {
						"enabled": false
					}
				}
			}`),
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewExperimentManager(data.input)

			// Validate
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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
						"name": "country",
						"field": "countryID",
						"field_source": "header"
					},
					{
						"name": "geo_area",
						"field": "gArea",
						"field_source": "payload"
					}
				],
				"treatment_service_config":null
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
	monkey.PatchInstanceMethod(
		reflect.TypeOf(em),
		"GetTreatmentServicePluginConfig",
		func(em *experimentManager) (*schema.TreatmentServicePluginConfig, error) {
			return nil, nil
		},
	)
	monkey.PatchInstanceMethod(
		reflect.TypeOf(em),
		"MakeTreatmentServiceConfig",
		func(em *experimentManager, treatmentServicePluginConfig *schema.TreatmentServicePluginConfig) (
			*config.TreatmentServiceConfig, error) {
			return nil, nil
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
		"success | all values": {
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
		"success | missing field": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field_source": "none"
					}
				]
			}`),
		},
		"failure | missing field source": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1"
					}
				]
			}`),
			err: strings.Join([]string{"Key: 'ExperimentConfig.Variables[0].FieldSource' Error:",
				"Field validation for 'FieldSource' failed on the 'required' tag",
			}, ""),
		},
		"failure | field is unset when field source is not none": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field_source": "header"
					}
				]
			}`),
			err: strings.Join([]string{"Key: 'ExperimentConfig.Variables[0].Field' ",
				"Error:Field validation for 'Field' failed on the 'Value must be set if FieldSource is not none' tag",
			}, ""),
		},
		"failure | field is set when field source is none": {
			cfg: json.RawMessage(`{
				"project_id": 1,
				"variables": [
					{
						"name": "var-1",
						"field": "var1",
						"field_source": "none"
					}
				]
			}`),
			err: strings.Join([]string{"Key: 'ExperimentConfig.Variables[0].Field' ",
				"Error:Field validation for 'Field' failed on the 'Value must not be set if FieldSource is none' tag",
			}, ""),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			em := experimentManager{validate: config.NewValidator()}
			err := em.ValidateExperimentConfig(data.cfg)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// testSetupEnvForGoogleCredentials creates a temporary file containing dummy service account JSON
// then set the environment variable GOOGLE_APPLICATION_CREDENTIALS to point to the file.
// This is useful for tests that assume Google Cloud Client libraries can automatically find
// the service account credentials in any environment.
// At the end of the test, the returned function can be called to perform cleanup.
func testSetupEnvForGoogleCredentials(t *testing.T) (reset func()) {
	serviceAccountKey := []byte(`{
		"type": "service_account",
		"project_id": "foo",
		"private_key_id": "bar",
		"private_key": "baz",
		"client_email": "foo@example.com",
		"client_id": "bar_client_id",
		"auth_uri": "https://oauth2.googleapis.com/auth",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`)

	file, err := os.CreateTemp("", "dummy-service-account")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(file.Name(), serviceAccountKey, 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", file.Name())
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		err := os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		if err != nil {
			t.Log("Cleanup failed", err)
		}
		err = os.Remove(file.Name())
		if err != nil {
			t.Log("Cleanup failed", err)
		}
	}
}
