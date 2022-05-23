package services

import (
	"github.com/jinzhu/gorm"

	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
)

type ListTreatmentHistoryParams struct {
	pagination.PaginationOptions
}

type TreatmentHistoryService interface {
	ListTreatmentHistory(treatmentId int64, params ListTreatmentHistoryParams) ([]*models.TreatmentHistory, *pagination.Paging, error)
	GetTreatmentHistory(treatmentId int64, version int64) (*models.TreatmentHistory, error)
	CreateTreatmentHistory(*models.Treatment) (*models.TreatmentHistory, error)
	GetDBRecord(treatmentId models.ID, version int64) (*models.TreatmentHistory, error)
	DeleteTreatmentHistory(treatmentId int64) error
}

type treatmentHistoryService struct {
	db *gorm.DB
}

func NewTreatmentHistoryService(db *gorm.DB) TreatmentHistoryService {
	return &treatmentHistoryService{
		db: db,
	}
}

func (svc *treatmentHistoryService) ListTreatmentHistory(
	treatmentId int64,
	params ListTreatmentHistoryParams,
) ([]*models.TreatmentHistory, *pagination.Paging, error) {
	var history []*models.TreatmentHistory
	query := svc.query().
		Where("treatment_id = ?", treatmentId).
		Order("updated_at desc")

	// Pagination
	var pagingResponse *pagination.Paging
	var count int
	err := pagination.ValidatePaginationParams(params.Page, params.PageSize)
	if err != nil {
		return nil, nil, err
	}
	pageOpts := pagination.NewPaginationOptions(params.Page, params.PageSize)
	// Count total
	query.Model(&history).Count(&count)
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

	// Filter treatments
	err = query.Find(&history).Error
	if err != nil {
		return nil, nil, err
	}

	return history, pagingResponse, nil
}

func (svc *treatmentHistoryService) GetTreatmentHistory(
	treatmentId int64,
	version int64,
) (*models.TreatmentHistory, error) {
	history, err := svc.GetDBRecord(models.ID(treatmentId), version)
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}

	return history, nil
}

func (svc *treatmentHistoryService) CreateTreatmentHistory(treatment *models.Treatment) (*models.TreatmentHistory, error) {
	var history []*models.TreatmentHistory
	var count int64
	// Begin transaction - so that getting the current count and creating the new record are
	// done in a single transaction.
	tx := svc.db.Begin()
	// Get the count of the existing treatment history records
	svc.query().Where("treatment_id = ?", treatment.ID).Model(&history).Count(&count)
	// Create the new history record
	newHistoryRecord, err := svc.save(&models.TreatmentHistory{
		Model: models.Model{
			CreatedAt: treatment.UpdatedAt,
		},
		TreatmentID:   treatment.ID,
		Version:       count + 1,
		Name:          treatment.Name,
		Configuration: treatment.Configuration,
		UpdatedBy:     treatment.UpdatedBy,
	})
	if err != nil {
		return nil, err
	}
	return newHistoryRecord, tx.Commit().Error
}

func (svc *treatmentHistoryService) DeleteTreatmentHistory(treatmentId int64) error {
	query := svc.query().
		Where("treatment_id = ?", treatmentId).
		Unscoped().
		Delete(models.TreatmentHistory{})
	if err := query.Error; err != nil {
		return err
	}
	return nil
}

func (svc *treatmentHistoryService) GetDBRecord(
	treatmentId models.ID,
	version int64,
) (*models.TreatmentHistory, error) {
	var history models.TreatmentHistory
	query := svc.query().
		Where("treatment_id = ?", treatmentId).
		Where("version = ?", version).
		First(&history)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (svc *treatmentHistoryService) query() *gorm.DB {
	return svc.db
}

func (svc *treatmentHistoryService) save(history *models.TreatmentHistory) (*models.TreatmentHistory, error) {
	var err error
	if svc.db.NewRecord(history) {
		err = svc.db.Create(history).Error
	} else {
		err = svc.db.Save(history).Error
	}
	if err != nil {
		return nil, err
	}
	return svc.GetDBRecord(history.TreatmentID, history.Version)
}
