package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"bou.ke/monkey"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/xp/treatment-service/appcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/xp/plugins/turing/config"
	"github.com/caraml-dev/xp/plugins/turing/internal/testutils"

	_config "github.com/caraml-dev/xp/treatment-service/config"
)

func TestNewExperimentRunner(t *testing.T) {
	dummyAppCtx := appcontext.AppContext{}
	// Patch appcontext
	monkey.Patch(appcontext.NewAppContext,
		func(treatmentServiceConfig *_config.Config) (*appcontext.AppContext, error) {
			return &dummyAppCtx, nil
		},
	)

	tests := map[string]struct {
		props json.RawMessage
		err   string
	}{
		"success": {
			props: json.RawMessage(`{
				"request_parameters": [
					{
						"name": "country",
						"field": "countryValue",
						"field_source": "payload"
					}
				],
				"treatment_service_config": {
					"assigned_treatment_logger": {
						"bq_config": {
							"dataset": "xp_dataset",
							"project": "xp_project",
							"table": "xp_table"
						},
						"kind": "bq",
						"queue_length": 100000
					},
					"debug_config": {
						"output_path": "/tmp"
					},
					"pub_sub": {
						"project": "dev",
						"topic_name": "xp-update",
						"pub_sub_timeout_seconds": 30
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
					"project_ids": ["1"],
					"swagger_config": {
						"enabled": false
					}
				}
			}`),
		},
		"failure | bad config": {
			props: json.RawMessage(`test`),
			err: fmt.Sprint(
				"Could not parse the XP runner configuration: ",
				"invalid character 'e' in literal true (expecting 'r')",
			),
		},
		"failure | missing config": {
			props: json.RawMessage(`{}`),
			err: fmt.Sprint(
				"Key: 'ExperimentRunnerConfig.RequestParameters' Error:",
				"Field validation for 'RequestParameters' failed on the 'required' tag\n",
				"Key: 'ExperimentRunnerConfig.TreatmentServiceConfig' Error:",
				"Field validation for 'TreatmentServiceConfig' failed on the 'required' tag",
			),
		},
		"failure | bad treatment service config": {
			props: json.RawMessage(`{
				"request_parameters": [
					{
						"name": "country",
						"field": "countryValue",
						"field_source": "payload"
					}
				],
				"treatment_service_config": {
										"assigned_treatment_logger": {
						"bq_config": {
							"dataset": "xp_dataset",
							"project": "xp_project",
							"table": "xp_table"
						},
						"kind": "bq",
						"queue_length": 100000
					},
					"debug_config": {
						"output_path": "/tmp"
					},
					"pub_sub": {
						"project": "dev",
						"topic_name": "xp-update",
						"pub_sub_timeout_seconds": 30
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
					"project_ids": ["1"],
					"swagger_config": {
						"enabled": false
					}
				}
			}`),
			err: "Key: 'ExperimentRunnerConfig.TreatmentServiceConfig.Port' Error:" +
				"Field validation for 'Port' failed on the 'required' tag",
		},
		"failure | 0 project ids specified": {
			props: json.RawMessage(`{
				"request_parameters": [
					{
						"name": "country",
						"field": "countryValue",
						"field_source": "payload"
					}
				],
				"treatment_service_config": {
					"assigned_treatment_logger": {
						"bq_config": {
							"dataset": "xp_dataset",
							"project": "xp_project",
							"table": "xp_table"
						},
						"kind": "bq",
						"queue_length": 100000
					},
					"debug_config": {
						"output_path": "/tmp"
					},
					"pub_sub": {
						"project": "dev",
						"topic_name": "xp-update",
						"pub_sub_timeout_seconds": 30
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
			err: "One and only one project id must be specified",
		},
		"failure | project id specified cannot be parsed into an int64 data type": {
			props: json.RawMessage(`{
				"request_parameters": [
					{
						"name": "country",
						"field": "countryValue",
						"field_source": "payload"
					}
				],
				"treatment_service_config": {
					"assigned_treatment_logger": {
						"bq_config": {
							"dataset": "xp_dataset",
							"project": "xp_project",
							"table": "xp_table"
						},
						"kind": "bq",
						"queue_length": 100000
					},
					"debug_config": {
						"output_path": "/tmp"
					},
					"pub_sub": {
						"project": "dev",
						"topic_name": "xp-update",
						"pub_sub_timeout_seconds": 30
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
					"project_ids": ["abc"],
					"swagger_config": {
						"enabled": false
					}
				}
			}`),
			err: "Error parsing project id string into int64: strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			expRunner, err := NewExperimentRunner(data.props)
			if data.err == "" {
				assert.NotNil(t, expRunner)
				assert.Nil(t, err)
			} else {
				assert.Nil(t, expRunner)
				if err != nil {
					assert.Equal(t, data.err, err.Error())
				}
			}
		})
	}
}

func TestMissingRequestValue(t *testing.T) {
	tests := map[string]struct {
		parameters []config.Variable
		payload    string
		header     http.Header
		expected   map[string]interface{}
		err        string
	}{
		"failure | field not found in payload": {
			parameters: []config.Variable{
				{
					Name:        "X",
					Field:       "Y",
					FieldSource: config.FieldSource(request.PayloadFieldSource),
				},
			},
			header:   make(http.Header),
			expected: make(map[string]interface{}),
			err:      "Field Y not found in the request payload: Key path not found",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a logger with a memory sink for testing error log
			logger, sink, err := testutils.NewLoggerWithMemorySink()
			require.NoError(t, err)

			expRunner := experimentRunner{parameters: test.parameters}
			// Get request params and compare
			actual := expRunner.getRequestParams(logger, test.header, []byte(test.payload))
			assert.Equal(t, test.expected, actual)

			if test.err != "" {
				var logObj struct {
					Msg string `json:"msg"`
				}
				err = json.Unmarshal(sink.Bytes(), &logObj)
				require.NoError(t, err)
				assert.Equal(t, test.err, logObj.Msg)
			}
		})
	}
}
