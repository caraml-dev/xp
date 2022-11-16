package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/caraml-dev/turing/engines/experiment/log"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	inproc "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/instrumentation"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/pkg/errors"

	xpclient "github.com/caraml-dev/xp/clients/treatment"
	"github.com/caraml-dev/xp/plugins/turing/config"
	"github.com/caraml-dev/xp/treatment-service/appcontext"
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
	parameters []config.Variable
	appContext *appcontext.AppContext
}

func (er *experimentRunner) GetTreatmentForRequest(
	reqHeader http.Header,
	body []byte,
	options runner.GetTreatmentOptions,
) (*runner.Treatment, error) {
	logger := log.With("turing_req_id", options.TuringRequestID)

	// Get the request parameters for the current request
	requestParams := er.getRequestParams(logger, reqHeader, body)

	projectId := models.NewProjectId(int64(er.projectID))

	// Initialize metric / log variables
	begin := time.Now()

	var requestFilter map[string][]*_segmenters.SegmenterValue

	var filteredExperiment *pubsub.Experiment
	var selectedTreatment *pubsub.ExperimentTreatment
	var switchbackWindowId *int64

	err := er.appContext.SchemaService.ValidatePasskey(projectId, er.passkey)
	if err != nil {
		return nil, err
	}

	// Use the S2ID at the max configured level (most granular level) to generate the filter
	requestFilter, err = er.appContext.SchemaService.GetRequestFilter(projectId, requestParams)
	if err != nil {
		return nil, err
	}
	_, filteredExperiment, err = er.appContext.ExperimentService.GetExperiment(projectId, requestFilter)
	if err != nil {
		return nil, err
	}
	experimentLookupLabels := er.appContext.MetricService.GetProjectNameLabel(projectId)
	er.appContext.MetricService.LogLatencyHistogram(begin, experimentLookupLabels, instrumentation.ExperimentLookupDurationMs)

	// Fetch treatment
	if filteredExperiment == nil {
		return &runner.Treatment{
			Config: nil,
		}, nil
	}

	randomizationKeyValue, err := er.appContext.SchemaService.GetRandomizationKeyValue(projectId, requestParams)
	if err != nil {
		return nil, err
	}

	selectedTreatment, switchbackWindowId, err = er.appContext.TreatmentService.GetTreatment(filteredExperiment, randomizationKeyValue)
	if err != nil {
		return nil, err
	}

	treatmentRepr := models.ExperimentTreatmentToOpenAPITreatment(selectedTreatment)

	// Marshal and return response
	treatment := schema.SelectedTreatment{
		ExperimentId:   filteredExperiment.Id,
		ExperimentName: filteredExperiment.Name,
		Treatment:      treatmentRepr,
		Metadata: schema.SelectedTreatmentMetadata{
			ExperimentVersion:  filteredExperiment.Version,
			ExperimentType:     models.ProtobufExperimentTypeToOpenAPI(filteredExperiment.Type),
			SwitchbackWindowId: switchbackWindowId,
		},
	}

	// Marshal Response Body
	rawConfig, err := json.Marshal(treatment)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling the treatment config: %s", err.Error())
	}

	return &runner.Treatment{
		ExperimentName: filteredExperiment.Name,
		Name:           selectedTreatment.Name,
		Config:         rawConfig,
	}, nil
}

func (er *experimentRunner) getRequestParams(
	logger log.Logger,
	reqHeader http.Header,
	body []byte,
) map[string]interface{} {
	// Get the request parameters for the current request
	requestParams := map[string]interface{}{}
	for _, param := range er.parameters {
		if param.FieldSource == "none" || param.Field == "" {
			// Parameter not configured
			continue
		}
		val, err := request.GetValueFromRequest(reqHeader, body, request.FieldSource(param.FieldSource), param.Field)
		if err != nil {
			logger.Errorf(err.Error())
		} else {
			requestParams[param.Name] = val
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

	// Init AppContext
	log.Debug(fmt.Sprint(config.TreatmentServiceConfig))
	appCtx, err := appcontext.NewAppContext(config.TreatmentServiceConfig)
	if err != nil {
		log.Panicf("Failed initializing application appcontext: %v", err)
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
		appContext: appCtx,
	}
	return r, nil
}
