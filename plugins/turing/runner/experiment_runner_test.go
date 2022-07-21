package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gojek/turing/engines/experiment/pkg/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojek/xp/plugins/turing/config"
	"github.com/gojek/xp/plugins/turing/internal/testutils"
)

func TestNewExperimentRunner(t *testing.T) {
	tests := map[string]struct {
		props json.RawMessage
		err   string
	}{
		"success": {
			props: json.RawMessage(`{
				"endpoint": "http://test-endpoint",
				"project_id": 10,
				"passkey": "abc",
				"timeout": "500ms",
				"request_parameters": [
					{
						"parameter": "country",
						"field": "countryValue",
						"field_source": "payload"
					}
				]
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
				"Key: 'ExperimentRunnerConfig.Endpoint' Error:",
				"Field validation for 'Endpoint' failed on the 'required' tag\n",
				"Key: 'ExperimentRunnerConfig.ProjectID' Error:",
				"Field validation for 'ProjectID' failed on the 'required' tag\n",
				"Key: 'ExperimentRunnerConfig.Passkey' Error:",
				"Field validation for 'Passkey' failed on the 'required' tag\n",
				"Key: 'ExperimentRunnerConfig.Timeout' Error:",
				"Field validation for 'Timeout' failed on the 'required' tag\n",
				"Key: 'ExperimentRunnerConfig.RequestParameters' Error:",
				"Field validation for 'RequestParameters' failed on the 'required' tag",
			),
		},
		"failure | bad timeout": {
			props: json.RawMessage(`{
				"endpoint": "http://test-endpoint",
				"project_id": 10,
				"passkey": "abc",
				"timeout": "500ss",
				"request_parameters": [
					{
						"parameter": "country",
						"field": "countryValue",
						"field_source": "payload"
					}
				]
			}`),
			err: "XP runner timeout 500ss is invalid",
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
		parameters []config.RequestParameter
		payload    string
		header     http.Header
		expected   map[string]string
		err        string
	}{
		"failure | field not found in payload": {
			parameters: []config.RequestParameter{
				{
					Parameter: "X",
					Field:     "Y",
					FieldSrc:  request.PayloadFieldSource,
				},
			},
			header:   make(http.Header),
			expected: make(map[string]string),
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
