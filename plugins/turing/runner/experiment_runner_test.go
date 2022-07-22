package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/gojek/turing/engines/experiment/pkg/request"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	xpclient "github.com/gojek/xp/clients/treatment"
	"github.com/gojek/xp/common/api/schema"
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

func TestFetchTreatment(t *testing.T) {
	treatmentClient := &xpclient.ClientWithResponses{}
	monkey.PatchInstanceMethod(
		reflect.TypeOf(treatmentClient),
		"FetchTreatmentWithBodyWithResponse",
		func(tc *xpclient.ClientWithResponses, ctx context.Context, projectId int64, params *xpclient.FetchTreatmentParams,
			contentType string, body io.Reader, reqEditors ...xpclient.RequestEditorFn,
		) (*xpclient.FetchTreatmentResponse, error) {
			var json200Struct struct {
				Data *schema.SelectedTreatment `json:"data,omitempty"`
			}

			if projectId == 1 {
				return &xpclient.FetchTreatmentResponse{
					Body:    []byte{},
					JSON200: &json200Struct,
				}, nil
			}
			if projectId == 2 {
				traffic := int32(50)
				selectedTreatment := &schema.SelectedTreatment{
					ExperimentId:   712,
					ExperimentName: "test_experiment",
					Treatment: schema.SelectedTreatmentData{
						Configuration: map[string]interface{}{"foo": "bar"},
						Name:          "test_experiment-control",
						Traffic:       &traffic,
					},
				}
				json200Struct.Data = selectedTreatment

				return &xpclient.FetchTreatmentResponse{
					Body:    []byte{},
					JSON200: &json200Struct,
				}, nil
			}
			return nil, fmt.Errorf("Unexpected ProjectID: %d", projectId)
		},
	)
	defer monkey.UnpatchAll()

	expRunner := experimentRunner{
		httpClient: treatmentClient,
		projectID:  1,
		passkey:    "abc",
		parameters: []config.RequestParameter{
			{Parameter: "country", Field: "Country", FieldSrc: "header"},
			{Parameter: "latitude", Field: "pos.lat", FieldSrc: "payload"},
			{Parameter: "longitude", Field: "pos.lng", FieldSrc: "payload"},
			{Parameter: "geo_area", Field: "geo-area", FieldSrc: "payload"},
			{Parameter: "order_id", Field: "order-id", FieldSrc: "payload"},
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
