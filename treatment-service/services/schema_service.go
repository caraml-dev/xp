package services

import (
	"errors"
	"fmt"
	"strconv"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/models"
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

func (ss *schemaService) GetRequestFilter(
	projectId models.ProjectId,
	filterParams map[string]interface{},
) (map[string][]*_segmenters.SegmenterValue, error) {
	// Retrieve Experiment variables for each Segmenter from Cached Project Settings in Storage
	projectSettings := ss.ProjectSettingsStorage.FindProjectSettingsWithId(projectId)
	if projectSettings == nil {
		return nil, ProjectSettingsNotFound(fmt.Sprintf("unable to find project id %d", projectId))
	}
	projectSettingsSegmenters := projectSettings.Segmenters

	allTransformations := map[string][]*_segmenters.SegmenterValue{}
	for _, k := range projectSettingsSegmenters.Names {
		transformation, err := ss.segmenterService.GetTransformation(projectId, k, filterParams, projectSettingsSegmenters.Variables[k].Value)
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
