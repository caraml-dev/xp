package services

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
)

type ListExperimentHistoryParams struct {
	pagination.PaginationOptions
}

type ExperimentHistoryService interface {
	ListExperimentHistory(experimentId int64, params ListExperimentHistoryParams) ([]*models.ExperimentHistory, *pagination.Paging, error)
	GetExperimentHistory(experimentId int64, version int64) (*models.ExperimentHistory, error)
	CreateExperimentHistory(*models.Experiment) (*models.ExperimentHistory, error)
	GetDBRecord(experimentId models.ID, version int64) (*models.ExperimentHistory, error)
}

type experimentHistoryService struct {
	db *gorm.DB
}

func NewExperimentHistoryService(db *gorm.DB) ExperimentHistoryService {
	return &experimentHistoryService{
		db: db,
	}
}

func (svc *experimentHistoryService) ListExperimentHistory(
	experimentId int64,
	params ListExperimentHistoryParams,
) ([]*models.ExperimentHistory, *pagination.Paging, error) {
	var history []*models.ExperimentHistory
	query := svc.query().
		Where("experiment_id = ?", experimentId).
		Order("updated_at desc")

	// Pagination
	var pagingResponse *pagination.Paging
	var count int64
	err := pagination.ValidatePaginationParams(params.Page, params.PageSize)
	if err != nil {
		return nil, nil, err
	}
	pageOpts := pagination.NewPaginationOptions(params.Page, params.PageSize)
	// Count total
	query.Model(&history).Count(&count)
	// Add offset and limit
	query = query.Offset(int((*pageOpts.Page - 1) * *pageOpts.PageSize))
	query = query.Limit(int(*pageOpts.PageSize))
	// Format opts into paging response
	pagingResponse = pagination.ToPaging(pageOpts, int(count))
	if pagingResponse.Page > 1 && pagingResponse.Pages < pagingResponse.Page {
		// Invalid query - total pages is less than the requested page
		return nil, nil, errors.Newf(errors.BadInput,
			"Requested page number %d exceeds total pages: %d.", pagingResponse.Page, pagingResponse.Pages)
	}

	// Filter experiments
	err = query.Find(&history).Error
	if err != nil {
		return nil, nil, err
	}

	return history, pagingResponse, nil
}

func (svc *experimentHistoryService) GetExperimentHistory(
	experimentId int64,
	version int64,
) (*models.ExperimentHistory, error) {
	history, err := svc.GetDBRecord(models.ID(experimentId), version)
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}

	return history, nil
}

func (svc *experimentHistoryService) CreateExperimentHistory(experiment *models.Experiment) (*models.ExperimentHistory, error) {
	return svc.save(&models.ExperimentHistory{
		Model: models.Model{
			CreatedAt: experiment.UpdatedAt,
		},
		ExperimentID: experiment.ID,
		Version:      experiment.Version,
		Description:  experiment.Description,
		EndTime:      experiment.EndTime,
		Interval:     experiment.Interval,
		Name:         experiment.Name,
		Segment:      experiment.Segment,
		Status:       experiment.Status,
		Treatments:   experiment.Treatments,
		Tier:         experiment.Tier,
		Type:         experiment.Type,
		StartTime:    experiment.StartTime,
		UpdatedBy:    experiment.UpdatedBy,
	})
}

func (svc *experimentHistoryService) GetDBRecord(
	experimentId models.ID,
	version int64,
) (*models.ExperimentHistory, error) {
	var history models.ExperimentHistory
	query := svc.query().
		Where("experiment_id = ?", experimentId).
		Where("version = ?", version).
		First(&history)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (svc *experimentHistoryService) query() *gorm.DB {
	return svc.db
}

func (svc *experimentHistoryService) save(history *models.ExperimentHistory) (*models.ExperimentHistory, error) {
	if err := svc.query().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(history).Error; err != nil {
		return nil, err
	}
	return svc.GetDBRecord(history.ExperimentID, history.Version)
}
