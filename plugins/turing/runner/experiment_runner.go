package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/experiment/log"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	inproc "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	routerMetrics "github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/api"
	"github.com/caraml-dev/xp/treatment-service/controller"
	"github.com/caraml-dev/xp/treatment-service/instrumentation"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/monitoring"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

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
	projectID  int64
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

	projectId := models.NewProjectId(er.projectID)

	// Initialize metric / log variables
	begin := time.Now()

	var requestFilter map[string][]*_segmenters.SegmenterValue

	var filteredExperiment *pubsub.Experiment
	var selectedTreatment *pubsub.ExperimentTreatment
	var treatment schema.SelectedTreatment
	var switchbackWindowId *int64
	var err error

	statusCode := http.StatusBadRequest
	filterParams := api.FetchTreatmentRequestBody{AdditionalProperties: requestParams}

	defer func() {
		if requestFilter == nil {
			requestFilter = map[string][]*_segmenters.SegmenterValue{}
		}
		er.appContext.MetricService.LogFetchTreatmentMetrics(begin, projectId, treatment, requestFilter, statusCode)
		if statusCode == http.StatusInternalServerError && err != nil {
			// This is typically a problem with the experiment configuration that should not have been allowed
			// by the Management Service, or other unexpected errors. Log the response to console, for tracking.
			controller.LogFetchTreatmentError(
				projectId,
				statusCode,
				err,
				filterParams,
				requestFilter,
			)
		}
	}()

	var lookupRequestFilters []models.SegmentFilter
	var errorLog *monitoring.ErrorResponseLog

	if er.appContext.AssignedTreatmentLogger != nil {
		defer func() {
			// Capture potential errors from other calls to service layer and prevent it from
			// slipping pass subsequent JSON marshaling errors
			if err != nil {
				errorLog = &monitoring.ErrorResponseLog{Code: statusCode, Error: err.Error()}
			}

			headerJson, err := json.Marshal(reqHeader)
			if err != nil {
				errorLog = &monitoring.ErrorResponseLog{Code: statusCode, Error: err.Error()}
			}
			bodyJson, err := json.Marshal(filterParams)
			if err != nil {
				errorLog = &monitoring.ErrorResponseLog{Code: statusCode, Error: err.Error()}
			}
			requestJson := &monitoring.Request{
				Header: string(headerJson),
				Body:   string(bodyJson),
			}

			var requestFilters []models.SegmentFilter
			if errorLog == nil {
				requestFilters = lookupRequestFilters
			}

			assignedTreatmentLog := &monitoring.AssignedTreatmentLog{
				ProjectID:  projectId,
				RequestID:  uuid.New().String(),
				Experiment: filteredExperiment,
				Treatment:  selectedTreatment,
				Request:    requestJson,
				Segmenters: requestFilters,
			}
			if filteredExperiment != nil {
				assignedTreatmentLog.TreatmentMetadata = &monitoring.TreatmentMetadata{
					ExperimentVersion:  filteredExperiment.Version,
					ExperimentType:     string(models.ProtobufExperimentTypeToOpenAPI(filteredExperiment.Type)),
					SwitchbackWindowId: switchbackWindowId,
				}
			}

			if errorLog != nil {
				assignedTreatmentLog.Error = errorLog
			}

			_ = er.appContext.AssignedTreatmentLogger.Append(assignedTreatmentLog)
		}()
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
		statusCode = http.StatusOK
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
	treatment = schema.SelectedTreatment{
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

	statusCode = http.StatusOK

	return &runner.Treatment{
		ExperimentName: filteredExperiment.Name,
		Name:           selectedTreatment.Name,
		Config:         rawConfig,
	}, nil
}

func (er *experimentRunner) RegisterMetricsCollector(
	collector metrics.Collector,
	metricsRegistrationHelper runner.MetricsRegistrationHelper,
) error {
	er.appContext.MetricService.SetMetricsCollector(collector)
	err := metricsRegistrationHelper.Register([]routerMetrics.Metric{
		{
			Name:        string(instrumentation.FetchTreatmentRequestDurationMs),
			Type:        routerMetrics.HistogramMetricType,
			Description: instrumentation.FetchTreatmentRequestDurationMsHelpString,
			Buckets:     instrumentation.RequestLatencyBuckets,
			Labels:      instrumentation.FetchTreatmentRequestDurationMsLabels,
		},
		{
			Name:        string(instrumentation.ExperimentLookupDurationMs),
			Type:        routerMetrics.HistogramMetricType,
			Description: instrumentation.ExperimentLookupDurationMsHelpString,
			Buckets:     instrumentation.RequestLatencyBuckets,
			Labels:      instrumentation.ExperimentLookupDurationMsLabels,
		},
		{
			Name:        string(instrumentation.FetchTreatmentRequestCount),
			Type:        routerMetrics.CounterMetricType,
			Description: instrumentation.FetchTreatmentRequestCountHelpString,
			Labels: append(
				er.appContext.MetricService.GetMetricLabels(),
				instrumentation.AdditionalFetchTreatmentRequestCountLabels...,
			),
		},
		{
			Name:        string(instrumentation.NoMatchingExperimentRequestCount),
			Type:        routerMetrics.CounterMetricType,
			Description: instrumentation.NoMatchingExperimentRequestCountHelpString,
			Labels: append(
				er.appContext.MetricService.GetMetricLabels(),
				instrumentation.AdditionalNoMatchingExperimentRequestCountLabels...,
			),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (er *experimentRunner) startBackgroundServices(
	errChannel chan error,
) {
	backgroundSvcCtx := context.Background()
	if er.appContext.MessageQueueService != nil {
		go func() {
			err := er.appContext.MessageQueueService.SubscribeToManagementService(backgroundSvcCtx)
			if err != nil {
				errChannel <- err
			}
		}()
	}
	if er.appContext.PollerService != nil {
		er.appContext.PollerService.Start()
	}
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
		val, err := request.GetValueFromHTTPRequest(reqHeader, body, request.FieldSource(param.FieldSource), param.Field)
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

	// Init AppContext
	appCtx, err := appcontext.NewAppContext(config.TreatmentServiceConfig)
	if err != nil {
		log.Panicf("Failed initializing application appcontext: %v", err)
	}

	// Retrieve project ID
	if len(config.TreatmentServiceConfig.ProjectIds) != 1 {
		return nil, fmt.Errorf("One and only one project id must be specified")
	}
	projectId, err := strconv.ParseInt(config.TreatmentServiceConfig.ProjectIds[0], 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing project id string into int64")
	}

	// Return new XP Runner
	r := &experimentRunner{
		projectID:  projectId,
		parameters: config.RequestParameters,
		appContext: appCtx,
	}

	// TODO: To find a way to handle errors from the errChannel in the future
	errChannel := make(chan error, 1)
	r.startBackgroundServices(errChannel)
	return r, nil
}
