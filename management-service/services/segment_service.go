package services

import (
	"fmt"

	"github.com/golang-collections/collections/set"
	"github.com/jinzhu/gorm"

	"github.com/gojek/xp/management-service/errors"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
)

type CreateSegmentRequestBody struct {
	Segment   models.ExperimentSegmentRaw `json:"segment" validate:"required,notBlank"`
	Name      string                      `json:"name" validate:"required,notBlank"`
	UpdatedBy *string                     `json:"updated_by,omitempty"`
}

type UpdateSegmentRequestBody struct {
	Segment   models.ExperimentSegmentRaw `json:"segment" validate:"required,notBlank"`
	UpdatedBy *string                     `json:"updated_by,omitempty"`
}

type ListSegmentsParams struct {
	pagination.PaginationOptions
	UpdatedBy *string                `json:"updated_by,omitempty"`
	Search    *string                `json:"search,omitempty"`
	Fields    *[]models.SegmentField `json:"fields,omitempty"`
}

type SegmentService interface {
	ListSegments(
		projectId int64,
		params ListSegmentsParams,
	) ([]*models.Segment, *pagination.Paging, error)
	GetSegment(projectId int64, segmentId int64) (*models.Segment, error)
	CreateSegment(settings models.Settings, segmentData CreateSegmentRequestBody) (*models.Segment, error)
	UpdateSegment(settings models.Settings, segmentId int64, segmentData UpdateSegmentRequestBody) (*models.Segment, error)
	DeleteSegment(projectId int64, segmentId int64) error

	GetDBRecord(projectId models.ID, segmentId models.ID) (*models.Segment, error)
}

type segmentService struct {
	services *Services
	db       *gorm.DB
}

func NewSegmentService(services *Services, db *gorm.DB) SegmentService {
	return &segmentService{
		services: services,
		db:       db,
	}
}

func (svc *segmentService) ListSegments(
	projectId int64,
	params ListSegmentsParams,
) ([]*models.Segment, *pagination.Paging, error) {
	var err error
	var segments []*models.Segment

	query := svc.query()
	if params.Fields != nil && len(*params.Fields) != 0 {
		err = validateListSegmentFieldNames(*params.Fields)
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
		query.Model(&segments).Count(&count)
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

	// Filter segments
	err = query.Find(&segments).Error
	if err != nil {
		return nil, nil, err
	}

	return segments, pagingResponse, nil
}

func (svc *segmentService) GetSegment(projectId int64, segmentId int64) (*models.Segment, error) {
	segment, err := svc.GetDBRecord(models.ID(projectId), models.ID(segmentId))
	if err != nil {
		return nil, errors.Newf(errors.NotFound, err.Error())
	}

	return segment, nil
}

func (svc *segmentService) CreateSegment(
	settings models.Settings,
	segmentData CreateSegmentRequestBody,
) (*models.Segment, error) {
	// Validate segment data
	err := svc.services.ValidationService.Validate(segmentData)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Validate segmenters
	err = svc.services.SegmenterService.ValidateExperimentSegment(
		settings.Config.Segmenters.Names,
		segmentData.Segment,
	)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Create the segment record
	segmenterStorageSchema, err := segmentData.Segment.ToStorageSchema(svc.services.SegmenterService.GetSegmenterTypes())
	if err != nil {
		return nil, err
	}
	segment := &models.Segment{
		ProjectID: settings.ProjectID,
		Name:      segmentData.Name,
		Segment:   segmenterStorageSchema,
		UpdatedBy: *segmentData.UpdatedBy,
	}

	// Save to DB
	segmentDBRecord, err := svc.save(segment)
	if err != nil {
		return nil, err
	}

	return segmentDBRecord, nil
}

func (svc *segmentService) UpdateSegment(
	settings models.Settings,
	segmentId int64,
	segmentData UpdateSegmentRequestBody,
) (*models.Segment, error) {
	// Validate segment data
	err := svc.services.ValidationService.Validate(segmentData)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Validate segmenters
	err = svc.services.SegmenterService.ValidateExperimentSegment(settings.Config.Segmenters.Names, segmentData.Segment)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Get current segment
	curSegment, err := svc.GetDBRecord(settings.ProjectID, models.ID(segmentId))
	if err != nil {
		return nil, err
	}

	// Validate segmenters
	err = svc.services.SegmenterService.ValidateExperimentSegment(settings.Config.Segmenters.Names, segmentData.Segment)
	if err != nil {
		return nil, errors.Newf(errors.BadInput, err.Error())
	}

	// Copy current segment's contents as segment history
	_, err = svc.services.SegmentHistoryService.CreateSegmentHistory(curSegment)
	if err != nil {
		return nil, err
	}

	// Update current segment and save to DB
	segmenterStorageSchema, err := segmentData.Segment.ToStorageSchema(svc.services.SegmenterService.GetSegmenterTypes())
	if err != nil {
		return nil, err
	}
	segmentDBRecord, err := svc.save(&models.Segment{
		// Copy the ID and the fixed fields
		ID:        curSegment.ID,
		ProjectID: curSegment.ProjectID,
		Name:      curSegment.Name,
		// Add the new data
		Segment:   segmenterStorageSchema,
		UpdatedBy: *segmentData.UpdatedBy,
	})
	if err != nil {
		return nil, err
	}

	return segmentDBRecord, nil
}

func (svc *segmentService) DeleteSegment(projectId int64, segmentId int64) error {
	// TODO: Explicitly delete all segment history first
	query := svc.query().
		Where("project_id = ?", projectId).
		Where("id = ?", segmentId).
		Unscoped().
		Delete(models.Segment{})
	if err := query.Error; err != nil {
		return err
	}
	return nil
}

func (svc *segmentService) GetDBRecord(projectId models.ID, segmentId models.ID) (*models.Segment, error) {
	var segment models.Segment
	query := svc.query().
		Where("project_id = ?", projectId).
		Where("id = ?", segmentId).
		First(&segment)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &segment, nil
}

func (svc *segmentService) query() *gorm.DB {
	return svc.db
}

func (svc *segmentService) save(segment *models.Segment) (*models.Segment, error) {
	var err error
	if svc.db.NewRecord(segment) {
		err = svc.db.Create(segment).Error
	} else {
		err = svc.db.Save(segment).Error
	}
	if err != nil {
		return nil, err
	}
	return svc.GetDBRecord(segment.ProjectID, segment.ID)
}

func validateListSegmentFieldNames(fields []models.SegmentField) error {
	allowedFields := set.New([]interface{}{models.SegmentFieldId, models.SegmentFieldName}...)
	for _, field := range fields {
		if !allowedFields.Has(field) {
			return fmt.Errorf("field %s is not supported, fields should only be name and/or id", field)
		}
	}
	return nil
}
