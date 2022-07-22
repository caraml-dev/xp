package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gojek/turing/engines/experiment/log"
	"github.com/gojek/turing/engines/experiment/pkg/request"
	inproc "github.com/gojek/turing/engines/experiment/plugin/inproc/runner"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/pkg/errors"

	xpclient "github.com/gojek/xp/clients/treatment"
	"github.com/gojek/xp/plugins/turing/config"
)

// init ensures this runner is registered when the package is imported.
func init() {
	err := inproc.Register("xp", NewExperimentRunner)
	if err != nil {
		log.Panicf("failed to register xp experiment runner: %v", err)
	}
}

// experimentRunner implements runner.ExperimentRunner
type experimentRunner struct {
	httpClient *xpclient.ClientWithResponses
	projectID  int
	passkey    string
	parameters []config.RequestParameter
}

func (er *experimentRunner) GetTreatmentForRequest(
	reqHeader http.Header,
	body []byte,
	options runner.GetTreatmentOptions,
) (*runner.Treatment, error) {
	logger := log.With("turing_req_id", options.TuringRequestID)

	// Get the request parameters for the current request
	requestParams := er.getRequestParams(logger, reqHeader, body)

	// Create Request Payload
	reqPayload, err := json.Marshal(requestParams)
	if err != nil {
		return nil, err
	}

	treatmentResponse, err := er.httpClient.FetchTreatmentWithBodyWithResponse(
		context.Background(),
		int64(er.projectID),
		&xpclient.FetchTreatmentParams{PassKey: er.passkey},
		"application/json",
		bytes.NewReader(reqPayload),
	)
	treatmentErrTpl := "Error retrieving treatment for the given request: %s"

	// Check for errors
	if err != nil {
		return nil, fmt.Errorf(treatmentErrTpl, err.Error())
	}
	if treatmentResponse.JSON400 != nil {
		return nil, fmt.Errorf(treatmentErrTpl, treatmentResponse.JSON400.Message)
	}
	if treatmentResponse.JSON500 != nil {
		return nil, fmt.Errorf(treatmentErrTpl, treatmentResponse.JSON500.Message)
	}
	if treatmentResponse.JSON200 == nil {
		return nil, fmt.Errorf(treatmentErrTpl, "empty response body")
	}

	// Marshal Response Body
	rawConfig, err := json.Marshal(treatmentResponse.JSON200.Data)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling the treatment config: %s", err.Error())
	}

	// Return treatment info
	var expName, treatmentName string
	if treatmentResponse.JSON200.Data != nil {
		expName = treatmentResponse.JSON200.Data.ExperimentName
		treatmentName = treatmentResponse.JSON200.Data.Treatment.Name
	}
	return &runner.Treatment{
		ExperimentName: expName,
		Name:           treatmentName,
		Config:         rawConfig,
	}, nil
}

func (er *experimentRunner) getRequestParams(
	logger log.Logger,
	reqHeader http.Header,
	body []byte,
) map[string]string {
	// Get the request parameters for the current request
	requestParams := map[string]string{}
	for _, param := range er.parameters {
		val, err := request.GetValueFromRequest(reqHeader, body, param.FieldSrc, param.Field)
		if err != nil {
			logger.Errorf(err.Error())
		} else {
			requestParams[param.Parameter] = val
		}
	}
	return requestParams
}

// NewExperimentRunner creates an instance of ExperimentRunner with the provided JSON config.
func NewExperimentRunner(jsonCfg json.RawMessage) (runner.ExperimentRunner, error) {
	// Ensure valid schema for the JSON config.
	var config config.ExperimentRunnerConfig
	err := json.Unmarshal(jsonCfg, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not parse the XP runner configuration")
	}

	// Validate that all required configs exist.
	err = config.Validate()
	if err != nil {
		return nil, err
	}

	// Ensure timeout set in config has a valid duration format.
	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("XP runner timeout %s is invalid", config.Timeout)
	}

	// Create XP client
	client, err := xpclient.NewClientWithResponses(
		config.Endpoint,
		xpclient.WithHTTPClient(&http.Client{Timeout: timeout}),
	)
	if err != nil {
		return nil, fmt.Errorf("Unable to create XP runner client: %s", err.Error())
	}

	// Return new XP Runner
	r := &experimentRunner{
		httpClient: client,
		projectID:  config.ProjectID,
		passkey:    config.Passkey,
		parameters: config.RequestParameters,
	}
	return r, nil
}
