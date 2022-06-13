package services

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/golang-collections/collections/set"

	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
	"github.com/gojek/turing-experiments/treatment-service/models"
)

type SchemaService interface {
	// GetRandomizationKeyValue retrieves the value of Randomization key based on projectId
	GetRandomizationKeyValue(projectId models.ProjectId, filterParams map[string]interface{}) (*string, error)

	// GetRequestFilter retrieves required request parameters based on projectId and builds a typed filter
	// for matching experiments
	GetRequestFilter(
		projectId models.ProjectId,
		filterParams map[string]interface{},
	) (map[string][]*_segmenters.SegmenterValue, error)

	// ValidatePasskey validates whether required passkey is provided based on projectId
	ValidatePasskey(projectId models.ProjectId, passkey string) error
	// ValidateSchema validates whether required request parameters are provided based on projectId
	ValidateSchema(projectId models.ProjectId, filterParams map[string]interface{}) error
}

type schemaService struct {
	requestParams     map[models.ProjectId]map[string]bool
	randomizationKeys map[models.ProjectId]string

	ProjectSettingsStorage models.ProjectSettingsStorage
	segmenterService       SegmenterService
}

type ProjectSettingsNotFoundError struct {
	message string
}

func ProjectSettingsNotFound(message string) *ProjectSettingsNotFoundError {
	return &ProjectSettingsNotFoundError{
		message: message,
	}
}
func (e *ProjectSettingsNotFoundError) Error() string {
	return e.message
}

func NewSchemaService(
	projectSettingsStorage models.ProjectSettingsStorage,
	segmenterService SegmenterService,
) (SchemaService, error) {
	svc := &schemaService{
		requestParams:          make(map[models.ProjectId]map[string]bool),
		randomizationKeys:      make(map[models.ProjectId]string),
		ProjectSettingsStorage: projectSettingsStorage,
		segmenterService:       segmenterService,
	}

	return svc, nil
}

func (ss *schemaService) ValidatePasskey(projectId models.ProjectId, passkey string) error {
	settings := ss.ProjectSettingsStorage.FindProjectSettingsWithId(projectId)
	if settings == nil {
		return ProjectSettingsNotFound(fmt.Sprintf("unable to find project id %d", projectId))
	}
	if passkey != settings.GetPasskey() {
		return errors.New("incorrect passkey was provided")
	}

	return nil
}

func (ss *schemaService) ValidateSchema(projectId models.ProjectId, filterParams map[string]interface{}) error {
	receivedRequestParameters := make([]interface{}, len(filterParams))
	i := 0
	for k := range filterParams {
		receivedRequestParameters[i] = k
		i++
	}

	requiredRequestParams, err := ss.getRequestParams(projectId)
	if err != nil {
		return err
	}

	// Validate and inform clients missing request parameters (if any)
	isValid, missingRequestParameters := func() (bool, []string) {
		hasAllRequiredParams := true
		receivedRequestParametersSet := set.New(receivedRequestParameters...)
		missingRequestParameters := []string{}
		for _, val := range requiredRequestParams {
			if !receivedRequestParametersSet.Has(val) {
				hasAllRequiredParams = false
				missingRequestParameters = append(missingRequestParameters, val)
			}
		}
		// Ensure order is deterministic
		sort.Strings(missingRequestParameters)
		return hasAllRequiredParams, missingRequestParameters
	}()
	if !isValid {
		return fmt.Errorf("required request parameters are not provided: %s", missingRequestParameters)
	}

	return nil
}

func (ss *schemaService) GetRequestFilter(
	projectId models.ProjectId,
	filterParams map[string]interface{},
) (map[string][]*_segmenters.SegmenterValue, error) {
	// Retrieve Experiment variables for each Segmenter from Cached Project Settings in Storage
	projectSettings := ss.ProjectSettingsStorage.FindProjectSettingsWithId(projectId)
	if projectSettings == nil {
		return nil, ProjectSettingsNotFound(fmt.Sprintf("unable to find project id %d", projectId))
	}
	projectSegmenters := ss.ProjectSettingsStorage.FindProjectSettingsWithId(projectId).Segmenters

	allTransformations := map[string][]*_segmenters.SegmenterValue{}
	for _, k := range projectSegmenters.Names {
		transformation, err := ss.segmenterService.GetTransformation(k, filterParams, projectSegmenters.Variables[k].Value)
		if err != nil {
			return nil, err
		}
		allTransformations[k] = transformation
	}

	return allTransformations, nil
}

func (ss *schemaService) GetRandomizationKeyValue(
	projectId models.ProjectId,
	filterParams map[string]interface{},
) (*string, error) {
	randomizationKey := ss.ProjectSettingsStorage.FindProjectSettingsWithId(projectId).RandomizationKey
	randomizationValue := filterParams[randomizationKey]
	if randomizationValue == nil {
		return nil, nil
	}
	var randomizationStringValue string
	switch randomizationValue := randomizationValue.(type) {
	case string:
		randomizationStringValue = randomizationValue
	case float64:
		randomizationStringValue = strconv.FormatFloat(randomizationValue, 'f', -1, 64)
	}

	return &randomizationStringValue, nil
}

func (ss *schemaService) getRequestParams(projectId models.ProjectId) ([]string, error) {
	settings := ss.ProjectSettingsStorage.FindProjectSettingsWithId(projectId)
	if settings == nil {
		return nil, ProjectSettingsNotFound(fmt.Sprintf("unable to find project id %d", projectId))
	}
	params := []string{}
	for _, variables := range settings.Segmenters.Variables {
		params = append(params, variables.Value...)
	}
	// Should contain randomization key
	params = append(params, settings.RandomizationKey)
	return params, nil
}
