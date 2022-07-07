package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-collections/collections/set"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jinzhu/gorm"

	"github.com/gojek/xp/common/api/schema"
	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/segmenters"
	"github.com/gojek/xp/management-service/utils"
)

type SegmenterScope string

const (
	SegmenterScopeGlobal  SegmenterScope = "global"
	SegmenterScopeProject SegmenterScope = "project"
)

type SegmenterStatus string

const (
	SegmenterStatusActive   SegmenterStatus = "active"
	SegmenterStatusInactive SegmenterStatus = "inactive"
)

var (
	SegmenterScopeMap = map[string]SegmenterScope{
		"global":  SegmenterScopeGlobal,
		"project": SegmenterScopeProject,
	}
	SegmenterStatusMap = map[string]SegmenterStatus{
		"active":   SegmenterStatusActive,
		"inactive": SegmenterStatusInactive,
	}
)

type CreateCustomSegmenterRequestBody struct {
	Name        string              `json:"name" validate:"required,notBlank"`
	Type        string              `json:"type" validate:"notBlank"`
	Options     *models.Options     `json:"options"`
	MultiValued bool                `json:"multi_valued"`
	Constraints *models.Constraints `json:"constraints"`
	Required    bool                `json:"required"`
	Description *string             `json:"description,omitempty"`
}

type UpdateCustomSegmenterRequestBody struct {
	Options     *models.Options     `json:"options"`
	MultiValued bool                `json:"multi_valued"`
	Constraints *models.Constraints `json:"constraints"`
	Required    bool                `json:"required"`
	Description *string             `json:"description,omitempty"`
}

type ListSegmentersParams struct {
	Scope  *SegmenterScope  `json:"scope,omitempty"`
	Status *SegmenterStatus `json:"status,omitempty"`
	Search *string          `json:"search,omitempty"`
}

type SegmenterService interface {
	GetFormattedSegmenters(projectId int64, expSegment models.ExperimentSegmentRaw) (map[string]*[]interface{}, error)
	GetSegmenterConfigurations(projectId int64, segmenterNames []string) ([]*_segmenters.SegmenterConfiguration, error)
	ValidateExperimentSegment(projectId int64, userSegmenters []string, expSegment models.ExperimentSegmentRaw) error
	ValidateSegmentOrthogonality(
		projectId int64,
		userSegmenters []string,
		expSegment models.ExperimentSegmentRaw,
		allExps []models.Experiment,
	) error
	ValidatePrereqSegmenters(projectId int64, segmenters []string) error
	ValidateRequiredSegmenters(projectId int64, segmenters []string) error
	ValidateExperimentVariables(projectId int64, projectSegmenters models.ProjectSegmenters) error
	GetSegmenter(projectId int64, name string) (*schema.Segmenter, error)
	ListSegmenters(projectId int64, params ListSegmentersParams) ([]*schema.Segmenter, error)
	ListGlobalSegmenters() ([]*schema.Segmenter, error)
	GetCustomSegmenter(projectId int64, name string) (*models.CustomSegmenter, error)
	CreateCustomSegmenter(
		projectId int64,
		customSegmenterData CreateCustomSegmenterRequestBody,
	) (*models.CustomSegmenter, error)
	UpdateCustomSegmenter(
		projectId int64,
		name string,
		customSegmenterData UpdateCustomSegmenterRequestBody,
	) (*models.CustomSegmenter, error)
	DeleteCustomSegmenter(projectId int64, name string) error
	GetDBRecord(projectId models.ID, name string) (*models.CustomSegmenter, error)
	GetSegmenterTypes(projectId int64) (map[string]schema.SegmenterType, error)
}

type segmenterService struct {
	globalSegmenters map[string]segmenters.Segmenter
	services         *Services
	db               *gorm.DB
}

func NewSegmenterService(
	services *Services,
	cfg map[string]interface{},
	db *gorm.DB) (SegmenterService, error) {

	globalSegmenters := make(map[string]segmenters.Segmenter)

	for name := range segmenters.Segmenters {
		if _, ok := cfg[name]; ok {
			configJSON, err := json.Marshal(cfg[name])
			if err != nil {
				return nil, err
			}

			m, err := segmenters.Get(name, configJSON)
			if err != nil {
				return nil, err
			}
			globalSegmenters[name] = m
			continue
		}
		m, err := segmenters.Get(name, nil)
		if err != nil {
			return nil, err
		}
		globalSegmenters[name] = m
	}

	return &segmenterService{
		globalSegmenters: globalSegmenters,
		db:               db,
		services:         services}, nil
}

func (svc *segmenterService) GetSegmenter(projectId int64, name string) (*schema.Segmenter, error) {
	// Get Active Segmenters
	activeSegmenterNames, err := svc.getActiveSegmenterNames(projectId)
	if err != nil {
		return nil, err
	}
	activeSegmenterSet := set.New()
	for _, activeSegmenterName := range activeSegmenterNames {
		activeSegmenterSet.Insert(activeSegmenterName)
	}

	// Check global segmenters if a segmenter with a matching name exists
	if globalSegmenter, err := svc.getGlobalSegmenter(name); globalSegmenter != nil && err == nil {
		formattedGlobalSegmenter, err := formatSegmenter(*globalSegmenter, activeSegmenterSet, schema.SegmenterScopeGlobal)
		if err != nil {
			return nil, err
		}
		return formattedGlobalSegmenter, nil
	}
	// Check custom segmenters if a segmenter with a matching name exists
	customSegmenter, err := svc.GetCustomSegmenter(projectId, name)
	if err != nil {
		return nil, err
	}
	baseSegmenter, err := customSegmenter.GetBaseSegmenter()
	if err != nil {
		return nil, err
	}
	formattedCustomSegmenter, err := formatSegmenter(baseSegmenter, activeSegmenterSet, schema.SegmenterScopeProject)
	if err != nil {
		return nil, err
	}
	formattedCustomSegmenter.UpdatedAt = &customSegmenter.UpdatedAt
	formattedCustomSegmenter.CreatedAt = &customSegmenter.CreatedAt

	return formattedCustomSegmenter, nil
}

// GetBaseSegmenter retrieves the global/custom segmenter with the given name, formatted as a BaseSegmenter
func (svc *segmenterService) GetBaseSegmenter(projectId int64, name string) (*segmenters.Segmenter, error) {
	// Check global segmenters if a segmenter with a matching name exists
	if globalSegmenter, err := svc.getGlobalSegmenter(name); globalSegmenter != nil && err == nil {
		return globalSegmenter, nil
	}
	// Check custom segmenters if a segmenter with a matching name exists
	var customSegmenter segmenters.Segmenter
	customSegmenter, err := svc.GetCustomSegmenter(projectId, name)
	if err != nil {
		return nil, err
	}

	return &customSegmenter, nil
}

func (svc *segmenterService) ListSegmenters(
	projectId int64,
	params ListSegmentersParams,
) ([]*schema.Segmenter, error) {
	allSegmenters := make([]*schema.Segmenter, 0)

	// Get Active Segmenters
	activeSegmenterNames, err := svc.getActiveSegmenterNames(projectId)
	if err != nil {
		return nil, err
	}
	activeSegmenterSet := set.New()
	for _, activeSegmenterName := range activeSegmenterNames {
		activeSegmenterSet.Insert(activeSegmenterName)
	}

	// Get the list of segmenter types for all segmenters
	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return nil, err
	}

	// Get all global segmenters
	if params.Scope == nil || *params.Scope == SegmenterScopeGlobal {
		globalSegmenters := svc.getGlobalSegmenters()
		for _, globalSegmenter := range globalSegmenters {
			formattedGlobalSegmenter, err := formatSegmenter(globalSegmenter, activeSegmenterSet, schema.SegmenterScopeGlobal)
			if err != nil {
				return nil, err
			}
			// Add segmenter to response
			if params.Status == nil || string(*params.Status) == string(*formattedGlobalSegmenter.Status) {
				allSegmenters = append(allSegmenters, formattedGlobalSegmenter)
			}
		}
	}

	// Get all custom segmenters
	if params.Scope == nil || *params.Scope == SegmenterScopeProject {
		customSegmenters, err := svc.getCustomSegmenters(projectId)
		if err != nil {
			return nil, err
		}
		for _, customSegmenter := range customSegmenters {
			if err := customSegmenter.FromStorageSchema(segmenterTypes); err != nil {
				return nil, err
			}
			baseSegmenter, err := customSegmenter.GetBaseSegmenter()
			if err != nil {
				return nil, err
			}
			formattedCustomSegmenter, err := formatSegmenter(baseSegmenter, activeSegmenterSet, schema.SegmenterScopeProject)
			if err != nil {
				return nil, err
			}
			// UpdatedAt and CreatedAt fields are manually updated for custom segmenters but not global segmenters since
			// global segmenters do not contain these fields
			formattedCustomSegmenter.UpdatedAt = &customSegmenter.UpdatedAt
			formattedCustomSegmenter.CreatedAt = &customSegmenter.CreatedAt
			if params.Status == nil || string(*params.Status) == string(*formattedCustomSegmenter.Status) {
				allSegmenters = append(allSegmenters, formattedCustomSegmenter)
			}
		}
	}

	// Search
	filteredResp := make([]*schema.Segmenter, 0)
	if params.Search != nil {
		for _, segmenter := range allSegmenters {
			if strings.Contains(segmenter.Name, *params.Search) {
				filteredResp = append(filteredResp, segmenter)
			}
		}
		return filteredResp, nil
	}

	return allSegmenters, nil
}

// ListGlobalSegmenters is a temporary method introduced to return global segmenters without the need for a project
// to have been set up, which was necessary because the UI couldn't query the available global segmenters when creating
// a new project. This method is essentially what supported the original '/segmenters' endpoint that was removed, and
// we can perhaps consider reusing this method if we are reintroducing that endpoint in the end.
func (svc *segmenterService) ListGlobalSegmenters() ([]*schema.Segmenter, error) {
	var formattedSegmenters []*schema.Segmenter
	globalSegmenters := svc.getGlobalSegmenters()
	for _, globalSegmenter := range globalSegmenters {
		config, err := globalSegmenter.GetConfiguration()
		if err != nil {
			return nil, err
		}
		// Format segmenters that comply with OpenAPI format
		formattedSegmenter, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(config)
		if err != nil {
			return nil, err
		}
		formattedSegmenters = append(formattedSegmenters, formattedSegmenter)
	}
	return formattedSegmenters, nil
}

func (svc *segmenterService) getGlobalSegmenter(name string) (*segmenters.Segmenter, error) {
	segmenter, ok := svc.globalSegmenters[name]
	if !ok {
		return nil, fmt.Errorf("unknown segmenter: %s", name)
	}
	return &segmenter, nil
}

func (svc *segmenterService) getGlobalSegmenters() []segmenters.Segmenter {
	var globalSegmenters []segmenters.Segmenter
	for _, v := range svc.globalSegmenters {
		globalSegmenters = append(globalSegmenters, v)
	}

	return globalSegmenters
}

func (svc *segmenterService) GetCustomSegmenter(projectId int64, name string) (*models.CustomSegmenter, error) {
	dbCustomSegmenter, err := svc.GetDBRecord(models.ID(projectId), name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("unknown segmenter: %s", name)
		}
		return nil, err
	}
	// Convert custom segmenter from DB schema
	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return nil, err
	}
	if err := dbCustomSegmenter.FromStorageSchema(segmenterTypes); err != nil {
		return nil, err
	}
	return dbCustomSegmenter, nil
}

func (svc *segmenterService) getCustomSegmenters(projectId int64) ([]models.CustomSegmenter, error) {
	var customSegmenters []models.CustomSegmenter

	query := svc.query().
		Where("project_id = ?", projectId).
		Order("updated_at desc").
		Find(&customSegmenters)
	if err := query.Error; err != nil {
		return nil, err
	}

	return customSegmenters, nil
}

func (svc *segmenterService) CreateCustomSegmenter(
	projectId int64,
	customSegmenterData CreateCustomSegmenterRequestBody,
) (*models.CustomSegmenter, error) {
	// Check all segmenters to ensure a segmenter with a matching name does not exist
	if segmenter, err := svc.GetSegmenter(projectId, customSegmenterData.Name); segmenter != nil && err == nil {
		return nil, errors.Newf(errors.BadInput, "a segmenter with the name %s already exists", customSegmenterData.Name)
	}

	// Get the list of segmenter types for all segmenters
	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return nil, err
	}

	// Validate new custom segmenter
	newCustomSegmenter, err := models.NewCustomSegmenter(
		models.ID(projectId),
		customSegmenterData.Name,
		models.SegmenterValueType(customSegmenterData.Type),
		customSegmenterData.Description,
		customSegmenterData.Required,
		customSegmenterData.MultiValued,
		customSegmenterData.Options,
		customSegmenterData.Constraints,
		segmenterTypes,
	)
	if err != nil {
		return nil, err
	}

	// Convert custom segmenter to DB schema
	err = newCustomSegmenter.ToStorageSchema(segmenterTypes)
	if err != nil {
		return nil, err
	}

	// Save to DB
	customSegmenterDBRecord, err := svc.save(
		newCustomSegmenter,
	)
	if err != nil {
		return nil, err
	}

	// Convert custom segmenter from DB schema
	if err := customSegmenterDBRecord.FromStorageSchema(segmenterTypes); err != nil {
		return nil, err
	}

	// Get SegmenterConfiguration expected by the Message Queue
	protoSegmenterConfig, err := customSegmenterDBRecord.GetConfiguration()
	if err != nil {
		return nil, err
	}
	if err = svc.services.PubSubPublisherService.PublishProjectSegmenterMessage("create", protoSegmenterConfig, projectId); err != nil {
		return nil, err
	}

	return customSegmenterDBRecord, nil
}

func (svc *segmenterService) UpdateCustomSegmenter(
	projectId int64,
	name string,
	customSegmenterData UpdateCustomSegmenterRequestBody,
) (*models.CustomSegmenter, error) {
	// Check custom segmenters to ensure a segmenter with a matching name exists
	curCustomSegmenter, err := svc.GetCustomSegmenter(projectId, name)
	if err != nil {
		return nil, err
	}

	// Get the list of segmenter types for all segmenters
	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return nil, err
	}

	// Validate updated custom segmenter
	updatedCustomSegmenter, err := models.NewCustomSegmenter(
		curCustomSegmenter.ProjectID,
		curCustomSegmenter.Name,
		curCustomSegmenter.Type,
		customSegmenterData.Description,
		customSegmenterData.Required,
		customSegmenterData.MultiValued,
		customSegmenterData.Options,
		customSegmenterData.Constraints,
		segmenterTypes,
	)
	if err != nil {
		return nil, err
	}

	// Convert custom segmenter to DB schema
	err = updatedCustomSegmenter.ToStorageSchema(segmenterTypes)
	if err != nil {
		return nil, err
	}

	// Save to DB
	customSegmenterDBRecord, err := svc.save(
		updatedCustomSegmenter,
	)
	if err != nil {
		return nil, err
	}

	// Convert custom segmenter from DB schema
	if err := customSegmenterDBRecord.FromStorageSchema(segmenterTypes); err != nil {
		return nil, err
	}

	// Get SegmenterConfiguration expected by the Message Queue
	protoSegmenterConfig, err := customSegmenterDBRecord.GetConfiguration()
	if err != nil {
		return nil, err
	}
	if err = svc.services.PubSubPublisherService.PublishProjectSegmenterMessage("update", protoSegmenterConfig, projectId); err != nil {
		return nil, err
	}
	return customSegmenterDBRecord, nil
}

func (svc *segmenterService) DeleteCustomSegmenter(projectId int64, name string) error {
	// Check custom segmenters if a segmenter with a matching name exists
	customSegmenter, err := svc.GetCustomSegmenter(projectId, name)
	if err != nil {
		return err
	}
	// Check if selected custom segmenter is currently in use in the project settings
	activeProjectSegmenterNames, err := svc.getActiveSegmenterNames(projectId)
	if err != nil {
		return err
	}
	for _, activeProjectSegmenterName := range activeProjectSegmenterNames {
		if activeProjectSegmenterName == customSegmenter.Name {
			return errors.Newf(
				errors.BadInput,
				"custom segmenter: %s is currently in use in the project settings and cannot be deleted",
				name,
			)
		}
	}

	query := svc.query().
		Where("project_id = ?", projectId).
		Where("name = ?", name).
		Unscoped().
		Delete(models.CustomSegmenter{})
	if err := query.Error; err != nil {
		return err
	}
	// Get SegmenterConfiguration expected by the Message Queue
	protoSegmenterConfig, err := customSegmenter.GetConfiguration()
	if err != nil {
		return err
	}
	if err = svc.services.PubSubPublisherService.PublishProjectSegmenterMessage("delete", protoSegmenterConfig, projectId); err != nil {
		return err
	}
	return nil
}

func (svc *segmenterService) GetDBRecord(projectId models.ID, name string) (*models.CustomSegmenter, error) {
	var customSegmenter models.CustomSegmenter
	query := svc.query().
		Where("project_id = ?", projectId).
		Where("name = ?", name).
		First(&customSegmenter)
	if err := query.Error; err != nil {
		return nil, err
	}

	return &customSegmenter, nil
}

func (svc *segmenterService) GetSegmenterTypes(projectId int64) (map[string]schema.SegmenterType, error) {
	segmenterTypes := map[string]schema.SegmenterType{}

	for key, val := range svc.globalSegmenters {
		switch val.GetType() {
		case _segmenters.SegmenterValueType_STRING:
			segmenterTypes[key] = schema.SegmenterTypeString
		case _segmenters.SegmenterValueType_INTEGER:
			segmenterTypes[key] = schema.SegmenterTypeInteger
		case _segmenters.SegmenterValueType_REAL:
			segmenterTypes[key] = schema.SegmenterTypeReal
		case _segmenters.SegmenterValueType_BOOL:
			segmenterTypes[key] = schema.SegmenterTypeBool
		}
	}

	customSegmenters, err := svc.getCustomSegmenters(projectId)
	if err != nil {
		return nil, err
	}

	for _, segmenter := range customSegmenters {
		switch segmenter.GetType() {
		case _segmenters.SegmenterValueType_STRING:
			segmenterTypes[segmenter.GetName()] = schema.SegmenterTypeString
		case _segmenters.SegmenterValueType_INTEGER:
			segmenterTypes[segmenter.GetName()] = schema.SegmenterTypeInteger
		case _segmenters.SegmenterValueType_REAL:
			segmenterTypes[segmenter.GetName()] = schema.SegmenterTypeReal
		case _segmenters.SegmenterValueType_BOOL:
			segmenterTypes[segmenter.GetName()] = schema.SegmenterTypeBool
		}
	}

	return segmenterTypes, nil
}

func (svc *segmenterService) GetSegmenterConfigurations(
	projectId int64,
	segmenterNames []string,
) ([]*_segmenters.SegmenterConfiguration, error) {
	// Convert to a generic interface map with formatted values
	segmenterConfigList := []*_segmenters.SegmenterConfiguration{}

	for _, segmenterName := range segmenterNames {
		segmenter, err := svc.GetBaseSegmenter(projectId, segmenterName)
		if err != nil {
			return segmenterConfigList, err
		}

		config, err := (*segmenter).GetConfiguration()
		if err != nil {
			return segmenterConfigList, err
		}
		segmenterConfigList = append(segmenterConfigList, config)
	}

	return segmenterConfigList, nil
}

func (svc *segmenterService) GetFormattedSegmenters(
	projectId int64,
	expSegment models.ExperimentSegmentRaw,
) (map[string]*[]interface{}, error) {
	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return nil, err
	}

	inputSegmenters, err := segmenters.ToProtoValues(expSegment, segmenterTypes)
	if err != nil {
		return nil, err
	}

	// Convert to a generic interface map with formatted values
	formattedMap := map[string]*[]interface{}{}

	for segmenterName, values := range inputSegmenters {
		segmenter, err := svc.GetBaseSegmenter(projectId, segmenterName)
		if err != nil {
			return formattedMap, err
		}

		if values != nil {
			// Format the segmenter values and add to the map
			segmenterType := (*segmenter).GetType()
			formattedValues := []interface{}{}
			for _, val := range values.GetValues() {
				switch segmenterType {
				case _segmenters.SegmenterValueType_STRING:
					// Quote the string
					formattedValues = append(formattedValues, fmt.Sprintf("%q", val.GetString_()))
				case _segmenters.SegmenterValueType_BOOL:
					formattedValues = append(formattedValues, val.GetBool())
				case _segmenters.SegmenterValueType_INTEGER:
					formattedValues = append(formattedValues, val.GetInteger())
				case _segmenters.SegmenterValueType_REAL:
					formattedValues = append(formattedValues, val.GetReal())
				}
			}
			formattedMap[segmenterName] = &formattedValues
		}
	}
	return formattedMap, nil
}

func (svc *segmenterService) ValidateExperimentSegment(
	projectId int64,
	userSegmenters []string,
	expSegment models.ExperimentSegmentRaw,
) error {
	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return err
	}

	inputSegmenters, err := segmenters.ToProtoValues(expSegment, segmenterTypes)
	if err != nil {
		return err
	}
	// For each user segmenter, check the detailed segmenter config
	for _, s := range userSegmenters {
		segmenter, err := svc.GetBaseSegmenter(projectId, s)
		if err != nil {
			return err
		}
		err = (*segmenter).ValidateSegmenterAndConstraints(inputSegmenters)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateSegmentOrthogonality checks that the given experiment's segment does not overlap
// with other given experiments. A segment is considered to overlap with another if each
// segmenter has one or more common values. The reverse makes them orthogonal - at least
// one segmenter has no common values.
func (svc *segmenterService) ValidateSegmentOrthogonality(
	projectId int64,
	userSegmenters []string,
	expSegment models.ExperimentSegmentRaw,
	allExps []models.Experiment,
) error {
	expSegmentFormatted, err := svc.GetFormattedSegmenters(projectId, expSegment)
	if err != nil {
		return err
	}

	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return err
	}

	for _, exp := range allExps {
		rawSegments, err := exp.Segment.ToRawSchema(segmenterTypes)
		if err != nil {
			return err
		}
		otherSegmentFormatted, err := svc.GetFormattedSegmenters(projectId, rawSegments)
		if err != nil {
			return err
		}

		// Check that the current experiment segment and the other are orthogonal
		segmentsOverlap := true
		for _, name := range userSegmenters {
			isCurrValEmpty, isOtherValEmpty := false, false
			currValues, ok := expSegmentFormatted[name]
			if !ok || currValues == nil || len(*currValues) == 0 {
				isCurrValEmpty = true
			}
			otherValues, ok := otherSegmentFormatted[name]
			if !ok || otherValues == nil || len(*otherValues) == 0 {
				isOtherValEmpty = true
			}

			// If both values non-empty, check overlap.
			// If only one of the values is empty, we can skip further checks.
			// If both empty, nothing to do.
			if !isCurrValEmpty && !isOtherValEmpty {
				currentSet := set.New(*currValues...)
				otherSet := set.New(*otherValues...)
				if currentSet.Intersection(otherSet).Len() == 0 {
					// At least one segmenter does not overlap, we can terminate the check for
					// this other experiment.
					segmentsOverlap = false
					break
				}
			} else if !isCurrValEmpty || !isOtherValEmpty {
				segmentsOverlap = false
				break
			}
		}

		if segmentsOverlap {
			return fmt.Errorf("Segment Orthogonality check failed against experiment ID %d", exp.ID)
		}
	}

	return nil
}

func (svc *segmenterService) ValidateRequiredSegmenters(projectId int64, segmenterNames []string) error {
	providedSegmenterNames := utils.StringSliceToSet(segmenterNames)

	// Get the list of segmenter types for all segmenters
	segmenterTypes, err := svc.GetSegmenterTypes(projectId)
	if err != nil {
		return err
	}

	// Validate required global segmenters are selected
	for k, v := range svc.globalSegmenters {
		config, err := v.GetConfiguration()
		if err != nil {
			return err
		}
		if config.Required {
			if !providedSegmenterNames.Has(k) {
				return fmt.Errorf("segmenter %s is a required segmenter that must be chosen", k)
			}
		}
	}
	// Validate required custom segmenters are selected
	customSegmenters, err := svc.getCustomSegmenters(projectId)
	if err != nil {
		return err
	}
	for _, customSegmenter := range customSegmenters {
		if err := customSegmenter.FromStorageSchema(segmenterTypes); err != nil {
			return err
		}
		config, err := customSegmenter.GetConfiguration()
		if err != nil {
			return err
		}
		if config.Required {
			baseSegmenterName := customSegmenter.GetName()
			if !providedSegmenterNames.Has(baseSegmenterName) {
				return fmt.Errorf("segmenter %s is a required segmenter that must be chosen", baseSegmenterName)
			}
		}
	}
	return nil
}

func (svc *segmenterService) ValidatePrereqSegmenters(projectId int64, segmenterNames []string) error {
	providedSegmenterNames := set.New()
	for _, segmenterName := range segmenterNames {
		providedSegmenterNames.Insert(segmenterName)
	}

	// Validate pre-requisite segmenters are selected
	for _, segmenterName := range segmenterNames {
		segmenter, err := svc.GetBaseSegmenter(projectId, segmenterName)
		if err != nil {
			return err
		}
		config, err := (*segmenter).GetConfiguration()
		if err != nil {
			return err
		}
		for _, constraint := range config.Constraints {
			prereqs := constraint.GetPreRequisites()
			for _, prereq := range prereqs {
				if providedSegmenterNames.Has(prereq.SegmenterName) {
					continue
				} else {
					return fmt.Errorf("segmenter %s requires %s to also be chosen", segmenterName, prereq.SegmenterName)
				}
			}
		}
	}
	return nil
}

func (svc *segmenterService) ValidateExperimentVariables(projectId int64, projectSegmenters models.ProjectSegmenters) error {
	if len(projectSegmenters.Names) != len(projectSegmenters.Variables) {
		return fmt.Errorf("len of project segmenters does not match mapping of experiment variables")
	}
	for _, segmentersName := range projectSegmenters.Names {
		providedVariables, ok := projectSegmenters.Variables[segmentersName]
		if !ok {
			return fmt.Errorf("project segmenters does not match mapping of experiment variables")
		}
		segmenter, err := svc.GetBaseSegmenter(projectId, segmentersName)
		if err != nil {
			return err
		}
		config, err := (*segmenter).GetConfiguration()
		if err != nil {
			return err
		}
		// flag to check if segmenter has matching variables as per segmenters setting
		isValid := false
		treatmentRequestFields := config.TreatmentRequestFields.GetValues()
		less := func(a, b string) bool { return a < b }
		for _, supportedVariables := range treatmentRequestFields {
			if isValid {
				break
			}
			// sorts and compare the slice if they are equal. Returns "" if equal.\
			isValid = cmp.Diff(supportedVariables.Value, providedVariables, cmpopts.SortSlices(less)) == ""
		}
		if !isValid {
			return fmt.Errorf("segmenter (%s) does not have valid experiment variable(s) provided", segmentersName)
		}
	}
	return nil
}

func (svc *segmenterService) query() *gorm.DB {
	return svc.db
}

func (svc *segmenterService) save(customSegmenter *models.CustomSegmenter) (*models.CustomSegmenter, error) {
	var err error
	if svc.db.NewRecord(customSegmenter) {
		err = svc.db.Create(customSegmenter).Error
	} else {
		err = svc.db.Save(customSegmenter).Error
	}
	if err != nil {
		return nil, err
	}
	return svc.GetDBRecord(customSegmenter.ProjectID, customSegmenter.Name)
}

func (svc segmenterService) getActiveSegmenterNames(projectId int64) ([]string, error) {
	dbRecord, err := svc.services.ProjectSettingsService.GetDBRecord(models.ID(projectId))
	if err != nil {
		return nil, err
	}
	settings := dbRecord.ToApiSchema()

	return settings.Segmenters.Names, nil
}

func formatSegmenter(segmenter segmenters.Segmenter, activeSegmenterSet *set.Set, scope schema.SegmenterScope) (*schema.Segmenter, error) {
	config, err := segmenter.GetConfiguration()
	if err != nil {
		return nil, err
	}
	// Format segmenters that comply with OpenAPI format
	formattedSegmenter, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(config)
	if err != nil {
		return nil, err
	}
	// Label global segmenters
	formattedSegmenter.Scope = &scope
	var segmenterStatus schema.SegmenterStatus

	if activeSegmenterSet.Has(formattedSegmenter.Name) {
		segmenterStatus = schema.SegmenterStatusActive
	} else {
		segmenterStatus = schema.SegmenterStatusInactive
	}
	formattedSegmenter.Status = &segmenterStatus

	return formattedSegmenter, nil
}
