package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/xp/clients/testutils/mocks"
	treatmentClient "github.com/caraml-dev/xp/clients/treatment"
	"github.com/caraml-dev/xp/plugins/turing/config"
	"github.com/caraml-dev/xp/plugins/turing/internal/testutils"
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
						"name": "country",
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
						"name": "country",
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
		parameters []config.Variable
		payload    string
		header     http.Header
		expected   map[string]string
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

func TestFetchTreatment(t *testing.T) {
	mockTreatmentClientInterface := mocks.TreatmentClientInterface{}
	mockTreatmentClient := treatmentClient.ClientWithResponses{ClientInterface: &mockTreatmentClientInterface}
	mockTreatmentClientInterface.On("FetchTreatmentWithBody",
		context.Background(),
		int64(1),
		&treatmentClient.FetchTreatmentParams{PassKey: "abc"},
		"application/json",
		mock.Anything,
	).Return(
		&http.Response{
			StatusCode: 200,
			Header:     map[string][]string{"Content-Type": {"json"}},
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
		}, nil)

	mockTreatmentClientInterface.On("FetchTreatmentWithBody",
		context.Background(),
		int64(2),
		&treatmentClient.FetchTreatmentParams{PassKey: "abc"},
		"application/json",
		mock.Anything,
	).Return(
		&http.Response{
			StatusCode: 200,
			Header:     map[string][]string{"Content-Type": {"json"}},
			Body: ioutil.NopCloser(
				bytes.NewBufferString(
					`{
							"data" : {
								"experiment_id": 712,
								"experiment_name": "test_experiment",
								"treatment": {
									"configuration": {
										"foo": "bar"
									},
									"name": "test_experiment-control",
									"traffic": 50
								}
							}
						}`,
				),
			),
		}, nil)

	expRunner := experimentRunner{
		httpClient: &mockTreatmentClient,
		projectID:  1,
		passkey:    "abc",
		parameters: []config.Variable{
			{Name: "country", Field: "Country", FieldSource: "header"},
			{Name: "latitude", Field: "pos.lat", FieldSource: "payload"},
			{Name: "longitude", Field: "pos.lng", FieldSource: "payload"},
			{Name: "geo_area", Field: "geo-area", FieldSource: "payload"},
			{Name: "order_id", Field: "order-id", FieldSource: "payload"},
		},
	}

	// Define tests
	testHeader := http.Header{
		http.CanonicalHeaderKey("Country"): []string{"SG"},
	}
	tests := map[string]struct {
		projectId int
		payload   string
		expected  *runner.Treatment
	}{
		"nil experiment": {
			projectId: 1,
			payload: `{
				"geo-area": "100",
				"pos": {
					"lat": "1.234",
					"lng": "103.5678"
				},
				"order-id": "12345"
			}`,
			expected: &runner.Treatment{
				Config: json.RawMessage("null"),
			},
		},
		"success": {
			projectId: 2,
			payload: `{
				"pos": {"lat": "1.2485558597961544", "lng": "103.54947567634105"},
				"geo-area": "50",
				"order-id": "12345"
			}`,
			expected: &runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "test_experiment-control",
				Config: json.RawMessage(`{
					"experiment_id": 712,
					"experiment_name": "test_experiment",
					"treatment": {
						"configuration": {"foo":"bar"},
						"name": "test_experiment-control",
						"traffic": 50
					}
				}`),
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Run experiment
			expRunner.projectID = data.projectId
			actual, err := expRunner.GetTreatmentForRequest(
				testHeader,
				[]byte(data.payload),
				runner.GetTreatmentOptions{})
			require.NoError(t, err)

			// Validate
			assert.Equal(t, data.expected.ExperimentName, actual.ExperimentName)
			assert.Equal(t, data.expected.Name, actual.Name)
			assert.JSONEq(t, string(data.expected.Config), string(actual.Config))
		})
	}
}
