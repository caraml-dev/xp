package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/gojek/xp/common/api/schema"
	"github.com/gojek/xp/common/pubsub"
	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/treatment-service/api"
	"github.com/gojek/xp/treatment-service/appcontext"
	"github.com/gojek/xp/treatment-service/config"
	"github.com/gojek/xp/treatment-service/instrumentation"
	"github.com/gojek/xp/treatment-service/models"
	"github.com/gojek/xp/treatment-service/monitoring"
	"github.com/gojek/xp/treatment-service/services"
)

type TreatmentController struct {
	*appcontext.AppContext
	Config *config.Config
}

func NewTreatmentController(ctx appcontext.AppContext, cfg config.Config) *TreatmentController {
	return &TreatmentController{AppContext: &ctx, Config: &cfg}
}

func (t TreatmentController) FetchTreatment(w http.ResponseWriter, r *http.Request, projectId_ int64, params api.FetchTreatmentParams) {
	w.Header().Set("ProjectId", strconv.Itoa(int(projectId_)))

	projectId := models.NewProjectId(projectId_)
	requestId := uuid.New().String()

	// Initialize metric / log variables
	begin := time.Now()
	treatment := schema.SelectedTreatment{}
	statusCode := http.StatusBadRequest
	var requestFilter map[string][]*_segmenters.SegmenterValue
	var filterParams api.FetchTreatmentRequestBody
	var err error

	defer func() {
		if requestFilter == nil {
			requestFilter = map[string][]*_segmenters.SegmenterValue{}
		}
		t.logFetchTreatmentMetrics(begin, projectId, treatment, requestFilter, statusCode)
		if statusCode == http.StatusInternalServerError && err != nil {
			// This is typically a problem with the experiment configuration that should not have been allowed
			// by the Management Service, or other unexpected errors. Log the response to console, for tracking.
			logFetchTreatmentError(projectId, statusCode, err, filterParams, requestFilter)
		}

	}()

	var filteredExperiment *pubsub.Experiment
	var selectedTreatment *pubsub.ExperimentTreatment
	var lookupRequestFilters []models.SegmentFilter
	var errorLog *monitoring.ErrorResponseLog
	if t.AppContext.AssignedTreatmentLogger != nil {
		defer func() {
			// Capture potential errors from other calls to service layer and prevent it from
			// slipping pass subsequent JSON marshaling errors
			if err != nil {
				errorLog = &monitoring.ErrorResponseLog{Code: statusCode, Error: err.Error()}
			}

			headerJson, err := json.Marshal(r.Header)
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
				RequestID:  requestId,
				Experiment: filteredExperiment,
				Treatment:  selectedTreatment,
				Request:    requestJson,
				Segmenters: requestFilters,
			}

			if errorLog != nil {
				assignedTreatmentLog.Error = errorLog
			}

			_ = t.AppContext.AssignedTreatmentLogger.Append(assignedTreatmentLog)
		}()
	}

	passkeyValue, passkeyPresent := r.Header["Pass-Key"]
	if !passkeyPresent {
		passkeyError := errors.New("pass-key header was not provided")
		errorLog = &monitoring.ErrorResponseLog{Code: statusCode, Error: passkeyError.Error()}
		ErrorResponse(w, statusCode, passkeyError, &requestId)
		return
	}
	err = t.SchemaService.ValidatePasskey(projectId, passkeyValue[0])
	if err != nil {
		ErrorResponse(w, statusCode, err, &requestId)
		return
	}

	filterParams = api.FetchTreatmentRequestBody{}
	err = json.NewDecoder(r.Body).Decode(&filterParams)
	if err != nil {
		ErrorResponse(w, statusCode, err, &requestId)
		return
	}

	err = t.SchemaService.ValidateSchema(projectId, filterParams.AdditionalProperties)
	if err != nil {
		switch err.(type) {
		default:
			ErrorResponse(w, statusCode, err, &requestId)
			return
		case *services.ProjectSettingsNotFoundError:
			statusCode = http.StatusNotFound
			ErrorResponse(w, statusCode, err, &requestId)
			return
		}
	}

	// Use the S2ID at the max configured level (most granular level) to generate the filter
	requestFilter, err = t.SchemaService.GetRequestFilter(projectId, filterParams.AdditionalProperties)
	if err != nil {
		switch err.(type) {
		default:
			ErrorResponse(w, statusCode, err, &requestId)
			return
		case *services.ProjectSettingsNotFoundError:
			statusCode = http.StatusNotFound
			ErrorResponse(w, statusCode, err, &requestId)
			return
		}
	}
	lookupRequestFilters, filteredExperiment, err = t.ExperimentService.GetExperiment(projectId, requestFilter)
	if err != nil {
		statusCode = http.StatusInternalServerError
		ErrorResponse(w, statusCode, err, &requestId)
		return
	}
	experimentLookupLabels := t.MetricService.GetProjectNameLabel(projectId)
	t.MetricService.LogLatencyHistogram(begin, experimentLookupLabels, instrumentation.ExperimentLookupDurationMs)

	// Fetch treatment
	if filteredExperiment == nil {
		Ok(w, api.FetchTreatmentSuccess{
			Data: nil,
		}, &requestId)
		return
	}

	randomizationKeyValue, err := t.SchemaService.GetRandomizationKeyValue(
		projectId, filterParams.AdditionalProperties,
	)
	if err != nil {
		ErrorResponse(w, statusCode, err, &requestId)
		return
	}

	selectedTreatment, err = t.TreatmentService.GetTreatment(filteredExperiment, randomizationKeyValue)
	if err != nil {
		statusCode = http.StatusInternalServerError
		ErrorResponse(w, statusCode, err, &requestId)
		return
	}

	treatmentRepr := models.ExperimentTreatmentToOpenAPITreatment(selectedTreatment)

	// Marshal and return response
	treatment = schema.SelectedTreatment{
		ExperimentId:   filteredExperiment.Id,
		ExperimentName: filteredExperiment.Name,
		Treatment:      treatmentRepr,
	}
	response := api.FetchTreatmentSuccess{
		Data: &treatment,
	}
	statusCode = http.StatusOK

	Ok(w, response, &requestId)
}

func (t TreatmentController) logFetchTreatmentMetrics(
	begin time.Time,
	projectId models.ProjectId,
	treatment schema.SelectedTreatment,
	requestFilter map[string][]*_segmenters.SegmenterValue,
	statusCode int,
) {
	labels := t.MetricService.GetLabels(
		projectId,
		treatment,
		statusCode,
		t.Config.MonitoringConfig.MetricLabels,
		requestFilter,
		false,
	)
	t.MetricService.LogLatencyHistogram(begin, labels, instrumentation.FetchTreatmentRequestDurationMs)

	labels = t.MetricService.GetLabels(
		projectId,
		treatment,
		statusCode,
		t.Config.MonitoringConfig.MetricLabels,
		requestFilter,
		true,
	)
	if treatment.ExperimentName != "" || treatment.Treatment.Name != "" {
		t.MetricService.LogRequestCount(labels, instrumentation.FetchTreatmentRequestCount)
	} else {
		delete(labels, "experiment_name")
		delete(labels, "treatment_name")
		t.MetricService.LogRequestCount(labels, instrumentation.NoMatchingExperimentRequestCount)
	}
}

func logFetchTreatmentError(
	projectId uint32,
	statusCode int,
	err error,
	request api.FetchTreatmentRequestBody,
	filters map[string][]*_segmenters.SegmenterValue) {
	logRecord := struct {
		ProjectID uint32                                   `json:"project_id,omitempty"`
		Error     string                                   `json:"error,omitempty"`
		Request   api.FetchTreatmentRequestBody            `json:"request,omitempty"`
		Filters   map[string][]*_segmenters.SegmenterValue `json:"filters,omitempty"`
	}{
		ProjectID: projectId,
		Error:     fmt.Sprintf("Code: %d, Message: %s", statusCode, err.Error()),
		Request:   request,
		Filters:   filters,
	}
	// Convert to JSON so that pointer data is encoded correctly, and log.
	bytes, _ := json.Marshal(logRecord)
	log.Println(string(bytes))
}
