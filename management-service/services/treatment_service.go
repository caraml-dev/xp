package services

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/golang-collections/collections/set"
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"

	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
)

type CreateTreatmentRequestBody struct {
	Config    map[string]interface{} `json:"configuration" validate:"required"`
	Name      string                 `json:"name" validate:"required,notBlank"`
	UpdatedBy *string                `json:"updated_by,omitempty"`
}

type UpdateTreatmentRequestBody struct {
	Config    map[string]interface{} `json:"configuration" validate:"required,notBlank"`
	UpdatedBy *string                `json:"updated_by,omitempty"`
}

type ListTreatmentsParams struct {
	pagination.PaginationOptions
	UpdatedBy *string                  `json:"updated_by,omitempty"`
	Search    *string                  `json:"search,omitempty"`
	Fields    *[]models.TreatmentField `json:"fields,omitempty"`
}

type TreatmentService interface {
	ListTreatments(
		projectId int64,
		params ListTreatmentsParams,
	) ([]*models.Treatment, *pagination.Paging, error)
	GetTreatment(projectId int64, treatmentId int64) (*models.Treatment, error)
	CreateTreatment(settings models.Settings, treatmentData CreateTreatmentRequestBody) (*models.Treatment, error)
	UpdateTreatment(settings models.Settings, treatmentId int64, treatmentData UpdateTreatmentRequestBody) (*models.Treatment, error)
	DeleteTreatment(projectId int64, treatmentId int64) error

	GetDBRecord(projectId models.ID, treatmentId models.ID) (*models.Treatment, error)
	RunCustomValidation(
		treatmentConfig map[string]interface{},
		settings models.Settings,
		context ValidationContext,
		operationType OperationType,
	) error
}

type treatmentService struct {
	services *Services
	db       *gorm.DB
}

func NewTreatmentService(
	services *Services,
	db *gorm.DB,
) TreatmentService {
	return &treatmentService{
		services: services,
		db:       db,
	}
}

func (svc *treatmentService) ListTreatments(
	projectId int64,
	params ListTreatmentsParams,
) ([]*models.Treatment, *pagination.Paging, error) {
	var err error
	var treatments []*models.Treatment

	query := svc.query()
	if params.Fields != nil && len(*params.Fields) != 0 {
		err = validateListTreatmentFieldNames(*params.Fields)
		if err != nil {
			return nil, nil, err
		}
		// query.Select only accepts []string
		var fieldNames []string
		for _, field := range *params.Fields {
			fieldNames = append(fieldNames, string(field))
		}
		query = query.Select(fieldNames)
	}

	query = query.
		Where("project_id = ?", projectId).
		Order("updated_at desc")

	// Handle optional parameters
	if params.UpdatedBy != nil {
		query = query.Where(
			fmt.Sprintf("updated_by ILIKE '%%%s%%'", *params.UpdatedBy),
		)
	}
	if params.Search != nil {
		query = query.Where(
			fmt.Sprintf("name ILIKE '%%%s%%'", *params.Search),
		)
	}

	// Pagination
	var pagingResponse *pagination.Paging
	var count int
	if params.Fields == nil || params.Page != nil || params.PageSize != nil {
		err = pagination.ValidatePaginationParams(params.Page, params.PageSize)
		if err != nil {
			return nil, nil, err
		}
		pageOpts := pagination.NewPaginationOptions(params.Page, params.PageSize)
		// Count total
		query.Model(&treatments).Count(&count)
		// Add offset and limit
		query = query.Offset((*pageOpts.Page - 1) * *pageOpts.PageSize)
		query = query.Limit(*pageOpts.PageSize)
		// Format opts into paging response
		pagingResponse = pagination.ToPaging(pageOpts, count)
		if pagingResponse.Page > 1 && pagingResponse.Pages < pagingResponse.Page {
			// Invalid query - total pages is less than the requested page
			return nil, nil, errors.Newf(errors.BadInput,
				"Requested page number %d exceeds total pages: %d.", pagingResponse.Page, pagingResponse.Pages)
		}
	}

	// Filter treatments
	err = query.Find(&treatments).Error
	if err != nil {
		return nil, nil, err
	}

	return treatments, pagingResponse, nil
}

func (svc *treatmentService) GetTreatment(projectId int64, treatmentId int64) (*models.Treatment, error) {
	treatment, err := svc.GetDBRecord(models.ID(projectId), models.ID(treatmentId))
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}

	return treatment, nil
}

func (svc *treatmentService) CreateTreatment(
	settings models.Settings,
	treatmentData CreateTreatmentRequestBody,
) (*models.Treatment, error) {
	// Validate treatment data
	err := svc.services.ValidationService.Validate(treatmentData)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Create the treatment record
	treatment := &models.Treatment{
		ProjectID:     settings.ProjectID,
		Name:          treatmentData.Name,
		Configuration: treatmentData.Config,
		UpdatedBy:     *treatmentData.UpdatedBy,
	}

	// Validate the treatment against the project settings' treatment schema and validation url
	err = svc.RunCustomValidation(treatment.Configuration, settings, ValidationContext{}, OperationTypeCreate)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Save to DB
	treatmentDBRecord, err := svc.save(treatment)
	if err != nil {
		return nil, err
	}

	return treatmentDBRecord, nil
}

func (svc *treatmentService) UpdateTreatment(
	settings models.Settings,
	treatmentId int64,
	treatmentData UpdateTreatmentRequestBody,
) (*models.Treatment, error) {
	// Validate treatment data
	err := svc.services.ValidationService.Validate(treatmentData)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Get current treatment
	curTreatment, err := svc.GetDBRecord(settings.ProjectID, models.ID(treatmentId))
	if err != nil {
		return nil, err
	}

	// Validate the treatment against the project settings' treatment schema and validation url
	err = svc.RunCustomValidation(treatmentData.Config, settings, ValidationContext{CurrentData: curTreatment},
		OperationTypeUpdate)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Copy current treatment's contents as treatment history
	_, err = svc.services.TreatmentHistoryService.CreateTreatmentHistory(curTreatment)
	if err != nil {
		return nil, err
	}

	// Update current treatment and save to DB
	treatmentDBRecord, err := svc.save(&models.Treatment{
		// Copy the ID and the fixed fields
		ID:        curTreatment.ID,
		ProjectID: curTreatment.ProjectID,
		Name:      curTreatment.Name,
		// Add the new data
		Configuration: treatmentData.Config,
		UpdatedBy:     *treatmentData.UpdatedBy,
	})
	if err != nil {
		return nil, err
	}

	return treatmentDBRecord, nil
}

func (svc *treatmentService) DeleteTreatment(projectId int64, treatmentId int64) error {
	// Explicitly delete all treatment history first
	err := svc.services.TreatmentHistoryService.DeleteTreatmentHistory(treatmentId)
	if err != nil {
		return err
	}

	query := svc.query().
		Where("project_id = ?", projectId).
		Where("id = ?", treatmentId).
		Unscoped().
		Delete(models.Treatment{})
	if err = query.Error; err != nil {
		return err
	}
	return nil
}

func (svc *treatmentService) GetDBRecord(projectId models.ID, treatmentId models.ID) (*models.Treatment, error) {
	var treatment models.Treatment
	query := svc.query().
		Where("project_id = ?", projectId).
		Where("id = ?", treatmentId).
		First(&treatment)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &treatment, nil
}

func (svc *treatmentService) query() *gorm.DB {
	return svc.db
}

func (svc *treatmentService) save(treatment *models.Treatment) (*models.Treatment, error) {
	var err error
	if svc.db.NewRecord(treatment) {
		err = svc.db.Create(treatment).Error
	} else {
		err = svc.db.Save(treatment).Error
	}
	if err != nil {
		return nil, err
	}
	return svc.GetDBRecord(treatment.ProjectID, treatment.ID)
}

func validateListTreatmentFieldNames(fields []models.TreatmentField) error {
	allowedFieldList := []interface{}{models.TreatmentFieldId, models.TreatmentFieldName}
	allowedFields := set.New(allowedFieldList...)
	for _, field := range fields {
		if !allowedFields.Has(field) {
			return fmt.Errorf("field %s is not supported, fields should only be name and/or id", field)
		}
	}
	return nil
}

// RunCustomValidation validates the given treatment by running it against the treatment schema AND the validation
// url given in the settings concurrently; if either of them return an error, this method returns an error
func (svc *treatmentService) RunCustomValidation(
	treatmentConfig map[string]interface{},
	settings models.Settings,
	context ValidationContext,
	operationType OperationType,
) error {
	g := new(errgroup.Group)

	g.Go(func() error {
		return ValidateTreatmentConfigWithTreatmentSchema(treatmentConfig, settings.TreatmentSchema)
	})
	g.Go(func() error {
		return svc.services.ValidationService.ValidateEntityWithExternalUrl(
			operationType,
			EntityTypeTreatment,
			treatmentConfig,
			context,
			settings.ValidationUrl,
		)
	})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// ValidateTreatmentConfigWithTreatmentSchema validates the given treatment config by running it against all the
// rules in the treatment schema concurrently; if any one of these rules return an error, this method returns an error
func ValidateTreatmentConfigWithTreatmentSchema(
	treatmentConfig map[string]interface{},
	treatmentSchema *models.TreatmentSchema,
) error {
	if treatmentSchema == nil {
		return nil
	}

	g := new(errgroup.Group)
	for _, rule := range treatmentSchema.Rules {
		rule := rule
		g.Go(func() error {
			return validateTreatmentConfigWithTemplateRule(treatmentConfig, rule)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// validateTreatmentConfigWithTemplateRule validates the given treatment config by running it against the given rule;
// if the rule expression evaluates to false, neither true nor false, or if an error occurs with parsing the expression,
// this method returns an error, otherwise it returns nil
func validateTreatmentConfigWithTemplateRule(
	treatmentConfig map[string]interface{},
	rule models.Rule,
) error {
	var output bytes.Buffer
	// The rule used is assumed to have been validated before
	t := template.Must(template.New(rule.Name).Funcs(sprig.FuncMap()).Parse(rule.Predicate))
	if err := t.Execute(&output, treatmentConfig); err != nil {
		return errors.Newf(errors.BadInput, "Error validating Go template rule %s: %v", rule.Name, err.Error())
	}

	switch output.String() {
	case "true":
		return nil
	case "false":
		return errors.Newf(errors.BadInput, "Go template rule %s returns false", rule.Name)
	default:
		return errors.Newf(errors.BadInput,
			"Go template rule %s returns a value that is neither 'true' nor 'false': %v",
			rule.Name,
			output.String())
	}
}
