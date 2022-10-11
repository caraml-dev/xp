package services

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
)

type ListSegmentHistoryParams struct {
	pagination.PaginationOptions
}

type SegmentHistoryService interface {
	ListSegmentHistory(segmentId int64, params ListSegmentHistoryParams) ([]*models.SegmentHistory, *pagination.Paging, error)
	GetSegmentHistory(segmentId int64, version int64) (*models.SegmentHistory, error)
	CreateSegmentHistory(*models.Segment) (*models.SegmentHistory, error)
	GetDBRecord(segmentId models.ID, version int64) (*models.SegmentHistory, error)
	DeleteSegmentHistory(segmentId int64) error
}

type segmentHistoryService struct {
	db *gorm.DB
}

func NewSegmentHistoryService(db *gorm.DB) SegmentHistoryService {
	return &segmentHistoryService{
		db: db,
	}
}

func (svc *segmentHistoryService) ListSegmentHistory(
	segmentId int64,
	params ListSegmentHistoryParams,
) ([]*models.SegmentHistory, *pagination.Paging, error) {
	var history []*models.SegmentHistory
	query := svc.query().
		Where("segment_id = ?", segmentId).
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

	// Filter segments
	err = query.Find(&history).Error
	if err != nil {
		return nil, nil, err
	}

	return history, pagingResponse, nil
}

func (svc *segmentHistoryService) GetSegmentHistory(
	segmentId int64,
	version int64,
) (*models.SegmentHistory, error) {
	history, err := svc.GetDBRecord(models.ID(segmentId), version)
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}

	return history, nil
}

func (svc *segmentHistoryService) CreateSegmentHistory(segment *models.Segment) (*models.SegmentHistory, error) {
	var history []*models.SegmentHistory
	var count int64
	// Begin transaction - so that getting the current count and creating the new record are
	// done in a single transaction.
	tx := svc.db.Begin()
	// Get the count of the existing segment history records
	svc.query().Where("segment_id = ?", segment.ID).Model(&history).Count(&count)
	// Create the new history record
	newHistoryRecord, err := svc.save(&models.SegmentHistory{
		Model: models.Model{
			CreatedAt: segment.UpdatedAt,
		},
		SegmentID: segment.ID,
		Version:   count + 1,
		Name:      segment.Name,
		Segment:   segment.Segment,
		UpdatedBy: segment.UpdatedBy,
	})
	if err != nil {
		return nil, err
	}
	return newHistoryRecord, tx.Commit().Error
}

func (svc *segmentHistoryService) DeleteSegmentHistory(segmentId int64) error {
	query := svc.query().
		Where("segment_id = ?", segmentId).
		Unscoped().
		Delete(models.SegmentHistory{})
	if err := query.Error; err != nil {
		return err
	}
	return nil
}

func (svc *segmentHistoryService) GetDBRecord(
	segmentId models.ID,
	version int64,
) (*models.SegmentHistory, error) {
	var history models.SegmentHistory
	query := svc.query().
		Where("segment_id = ?", segmentId).
		Where("version = ?", version).
		First(&history)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (svc *segmentHistoryService) query() *gorm.DB {
	return svc.db
}

func (svc *segmentHistoryService) save(history *models.SegmentHistory) (*models.SegmentHistory, error) {
	if err := svc.query().Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(history).Error; err != nil {
		return nil, err
	}
	return svc.GetDBRecord(history.SegmentID, history.Version)
}
