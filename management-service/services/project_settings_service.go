package services

import (
	"time"

	"github.com/golang-collections/collections/set"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/utils"
)

const PASSKEY_LENGTH = 32

type CreateProjectSettingsRequestBody struct {
	EnableS2idClustering *bool                    `json:"enable_s2id_clustering,omitempty"`
	RandomizationKey     string                   `json:"randomization_key" validate:"required,notBlank"`
	Segmenters           models.ProjectSegmenters `json:"segmenters" validate:"required"`
	TreatmentSchema      *models.TreatmentSchema  `json:"treatment_schema" validate:"omitempty"`
	ValidationUrl        *string                  `json:"validation_url" validate:"omitempty,url"`
	Username             string                   `json:"username" validate:"required,notBlank"`
}

type UpdateProjectSettingsRequestBody struct {
	EnableS2idClustering *bool                    `json:"enable_s2id_clustering,omitempty"`
	RandomizationKey     string                   `json:"randomization_key" validate:"required,notBlank"`
	Segmenters           models.ProjectSegmenters `json:"segmenters" validate:"required,notBlank"`
	TreatmentSchema      *models.TreatmentSchema  `json:"treatment_schema" validate:"omitempty"`
	ValidationUrl        *string                  `json:"validation_url" validate:"omitempty,url"`
}

type ProjectSettingsService interface {
	ListProjects() (*[]models.Project, error)

	GetProjectSettings(projectId int64) (*models.Settings, error)
	GetExperimentVariables(projectId int64) (*[]string, error)
	CreateProjectSettings(projectId int64, settings CreateProjectSettingsRequestBody) (*models.Settings, error)
	UpdateProjectSettings(projectId int64, settings UpdateProjectSettingsRequestBody) (*models.Settings, error)

	GetDBRecord(projectId models.ID) (*models.Settings, error)
}

type projectSettingsService struct {
	services *Services
	db       *gorm.DB
}

func NewProjectSettingsService(services *Services, db *gorm.DB) ProjectSettingsService {
	return &projectSettingsService{
		services: services,
		db:       db,
	}
}

func (svc *projectSettingsService) ListProjects() (*[]models.Project, error) {
	var dbRecords []*models.Settings
	err := svc.query().Find(&dbRecords).Error
	if err != nil {
		return nil, err
	}

	// Convert to the format expected by the API
	projectsResponse := []models.Project{}
	for _, r := range dbRecords {
		projectsResponse = append(projectsResponse, models.Project{
			CreatedAt:        r.CreatedAt,
			Id:               r.ProjectID.ToApiSchema(),
			RandomizationKey: r.Config.RandomizationKey,
			Segmenters:       r.Config.Segmenters.Names,
			UpdatedAt:        r.UpdatedAt,
			Username:         r.Username,
		})
	}
	return &projectsResponse, nil
}

func (svc *projectSettingsService) GetProjectSettings(projectId int64) (*models.Settings, error) {
	dbRecord, err := svc.GetDBRecord(models.ID(projectId))
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}
	return dbRecord, nil
}

func (svc *projectSettingsService) GetExperimentVariables(projectId int64) (*[]string, error) {
	dbRecord, err := svc.GetDBRecord(models.ID(projectId))
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}

	// use to track if a variable was appended. assumption is variable name is unique
	segmenterSet := set.New()
	var segmenterParams []string
	// Combine all the variables of project segmenters into a single array
	for _, variables := range dbRecord.Config.Segmenters.Variables {
		for _, variable := range variables {
			if segmenterSet.Has(variable) {
				continue
			}
			segmenterSet.Insert(variable)
			segmenterParams = append(segmenterParams, variable)
		}
	}
	// Add randomization key
	segmenterParams = append(segmenterParams, dbRecord.Config.RandomizationKey)
	return &segmenterParams, nil
}

func (svc *projectSettingsService) CreateProjectSettings(
	projectId int64,
	settings CreateProjectSettingsRequestBody,
) (*models.Settings, error) {
	// Check that the given project doesn't already have settings in the DB
	_, err := svc.GetDBRecord(models.ID(projectId))
	if err == nil {
		return nil, errors.Newf(errors.BadInput, "Project has already been set up")
	}

	// Validate settings data
	err = svc.services.ValidationService.Validate(settings)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Validate segmenter are recognized and experiment variable mapping are accepted as system allowed
	err = svc.services.SegmenterService.ValidateExperimentVariables(projectId, settings.Segmenters)
	if err != nil {
		return nil, err
	}

	// Verify required segmenters are provided
	err = svc.services.SegmenterService.ValidateRequiredSegmenters(projectId, settings.Segmenters.Names)
	if err != nil {
		return nil, err
	}

	// Verify dependent segmenters are provided
	err = svc.services.SegmenterService.ValidatePrereqSegmenters(projectId, settings.Segmenters.Names)
	if err != nil {
		return nil, err
	}

	// Generate random Passkey
	passkey, err := utils.GenerateRandomBase16String(PASSKEY_LENGTH)
	if err != nil {
		return nil, err
	}

	// Create the settings record
	settingsRecord := &models.Settings{
		ProjectID: models.ID(projectId),
		Username:  settings.Username,
		Passkey:   passkey,
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names:     settings.Segmenters.Names,
				Variables: settings.Segmenters.Variables,
			},
			RandomizationKey: settings.RandomizationKey,
		},
		TreatmentSchema: settings.TreatmentSchema,
		ValidationUrl:   settings.ValidationUrl,
	}
	if settings.EnableS2idClustering != nil {
		settingsRecord.Config.S2IDClusteringEnabled = *(settings.EnableS2idClustering)
	}

	// Save to DB
	dbRecord, err := svc.save(settingsRecord)
	if err != nil {
		return nil, err
	}

	// Convert to the format expected by the Message Queue
	protoExpResponse := dbRecord.ToProtoSchema()
	err = svc.services.PubSubPublisherService.PublishProjectSettingsMessage("create", &protoExpResponse)
	if err != nil {
		return nil, err
	}

	return dbRecord, nil
}

func (svc *projectSettingsService) UpdateProjectSettings(
	projectId int64,
	settings UpdateProjectSettingsRequestBody,
) (*models.Settings, error) {
	// Get the existing settings from the DB
	dbRecord, err := svc.GetDBRecord(models.ID(projectId))
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}

	// Validate settings data
	err = svc.services.ValidationService.Validate(settings)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Validate segmenter are recognized and experiment variable mapping are accepted as system allowed
	err = svc.services.SegmenterService.ValidateExperimentVariables(projectId, settings.Segmenters)
	if err != nil {
		return nil, err
	}

	// Verify Segmenter names are recognised
	_, err = svc.services.SegmenterService.GetSegmenterConfigurations(projectId, settings.Segmenters.Names)
	if err != nil {
		return nil, err
	}

	// Verify required segmenters are provided
	err = svc.services.SegmenterService.ValidateRequiredSegmenters(projectId, settings.Segmenters.Names)
	if err != nil {
		return nil, err
	}

	// Verify dependent segmenters are provided
	err = svc.services.SegmenterService.ValidatePrereqSegmenters(projectId, settings.Segmenters.Names)
	if err != nil {
		return nil, err
	}

	// Verify pairwise orthogonality checks are valid for all experiments
	err = svc.validateProjectSettingsUpdate(projectId, dbRecord.Config.Segmenters.Names, settings.Segmenters.Names)
	if err != nil {
		return nil, err
	}

	// Set the configurable fields
	if settings.EnableS2idClustering != nil {
		dbRecord.Config.S2IDClusteringEnabled = *(settings.EnableS2idClustering)
	}
	dbRecord.Config.RandomizationKey = settings.RandomizationKey
	dbRecord.Config.Segmenters = settings.Segmenters
	dbRecord.TreatmentSchema = settings.TreatmentSchema
	dbRecord.ValidationUrl = settings.ValidationUrl

	// Save to the DB
	dbRecord, err = svc.save(dbRecord)
	if err != nil {
		return nil, err
	}

	// Convert to the format expected by the Message Queue
	protoExpResponse := dbRecord.ToProtoSchema()
	err = svc.services.PubSubPublisherService.PublishProjectSettingsMessage("update", &protoExpResponse)
	if err != nil {
		return nil, err
	}

	return dbRecord, nil
}

func (svc *projectSettingsService) GetDBRecord(projectId models.ID) (*models.Settings, error) {
	var settings models.Settings
	query := svc.query().
		Where("project_id = ?", projectId).
		First(&settings)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (svc *projectSettingsService) query() *gorm.DB {
	return svc.db
}

func (svc *projectSettingsService) save(settings *models.Settings) (*models.Settings, error) {
	if err := svc.query().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(settings).Error; err != nil {
		return nil, err
	}
	return svc.GetDBRecord(settings.ProjectID)
}

func (svc *projectSettingsService) validateProjectSettingsUpdate(
	projectId int64,
	currentSegmenters []string,
	updatedSegmenters []string,
) error {
	// Perform orthogonality checks when there are removed segmenter(s)
	currentSegmentersSet := utils.StringSliceToSet(currentSegmenters)
	updatedSegmentersSet := utils.StringSliceToSet(updatedSegmenters)
	hasSetDifferences := currentSegmentersSet.Difference(updatedSegmentersSet)

	status := models.ExperimentStatusActive
	startTime := time.Now()
	endTime := time.Now().Add(855360 * time.Hour)
	listExpParams := ListExperimentsParams{StartTime: &startTime, EndTime: &endTime, Status: &status}

	if hasSetDifferences.Len() > 0 {
		id := models.ID(projectId)
		exps, err := svc.services.ExperimentService.ListAllExperiments(id, listExpParams)
		if err != nil {
			return err
		}

		err = svc.services.ExperimentService.ValidatePairwiseExperimentOrthogonality(projectId, exps, updatedSegmenters)
		if err != nil {
			return err
		}
		// Check if the set of updated segmenters contains all the segments specified by all the experiments
		err = svc.services.ExperimentService.ValidateProjectExperimentSegmentersExist(projectId, exps, updatedSegmenters)
		if err != nil {
			return err
		}
	}

	return nil
}
