//go:build integration

package services_test

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/gojek/xp/common/api/schema"
	tu "github.com/gojek/xp/management-service/internal/testutils"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/pagination"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

type SegmentServiceTestSuite struct {
	suite.Suite
	services.SegmentService

	*mocks.SegmenterService
	*mocks.ValidationService
	*mocks.SegmentHistoryService

	CleanUpFunc func()

	Settings models.Settings
	Segments []*models.Segment
}

func (s *SegmentServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up SegmentServiceTestSuite")

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB()
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	// Init services
	s.SegmenterService = &mocks.SegmenterService{}
	s.ValidationService = &mocks.ValidationService{}
	// Init segment history svc, mock calls will be set up during the test
	s.SegmentHistoryService = &mocks.SegmentHistoryService{}

	allServices := &services.Services{
		SegmenterService:      s.SegmenterService,
		ValidationService:     s.ValidationService,
		SegmentHistoryService: s.SegmentHistoryService,
	}

	// Init segment service
	s.SegmentService = services.NewSegmentService(allServices, db)

	// Create test data
	s.Settings, s.Segments, err = createTestSegments(db)
	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func (s *SegmentServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up SegmentServiceTestSuite")
	s.CleanUpFunc()
}

func TestSegmentService(t *testing.T) {
	suite.Run(t, new(SegmentServiceTestSuite))
}

func (s *SegmentServiceTestSuite) TestSegmentServiceGetIntegration() {
	segmentResponse, err := s.SegmentService.GetSegment(1, 1)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.Segments[0], segmentResponse)
}

func (s *SegmentServiceTestSuite) TestSegmentServiceListCreateUpdateDeleteIntegration() {
	// Test list segments first, since the create/update/delete of segments
	// could affect the results
	testListSegments(s)
	testCreateUpdateDeleteSegment(s)
}

func testListSegments(s *SegmentServiceTestSuite) {
	t := s.Suite.T()
	svc := s.SegmentService

	testSearchTerm := "-1"
	testUpdatedBy := "test-user-1"
	testPage := int32(1)
	testPageSize := int32(2)

	tests := []struct {
		name             string
		projectId        int64
		params           services.ListSegmentsParams
		expectedPaging   *pagination.Paging
		expectedSegments []*models.Segment
		expectedError    string
	}{
		{
			name:             "no params",
			projectId:        1,
			expectedPaging:   &pagination.Paging{Page: 1, Pages: 1, Total: 2},
			expectedSegments: []*models.Segment{s.Segments[1], s.Segments[0]},
		},
		{
			name:             "no segments filtered",
			projectId:        2,
			expectedPaging:   &pagination.Paging{Page: 1, Pages: 0, Total: 0},
			expectedSegments: []*models.Segment{},
		},
		{
			name:             "search term parameter",
			projectId:        1,
			params:           services.ListSegmentsParams{Search: &testSearchTerm},
			expectedPaging:   &pagination.Paging{Page: 1, Pages: 1, Total: 1},
			expectedSegments: []*models.Segment{s.Segments[0]},
		},
		{
			name:      "fields parameter without pagination",
			projectId: 1,
			params: services.ListSegmentsParams{
				Fields: &[]models.SegmentField{models.SegmentFieldId},
			},
			expectedSegments: []*models.Segment{{ID: s.Segments[1].ID}, {ID: s.Segments[0].ID}},
		},
		{
			name:      "all parameters",
			projectId: 1,
			params: services.ListSegmentsParams{
				Search: &testSearchTerm,
				PaginationOptions: pagination.PaginationOptions{
					Page:     &testPage,
					PageSize: &testPageSize,
				},
				Fields:    &[]models.SegmentField{models.SegmentFieldName},
				UpdatedBy: &testUpdatedBy,
			},
			expectedPaging:   &pagination.Paging{Page: 1, Pages: 1, Total: 1},
			expectedSegments: []*models.Segment{{Name: s.Segments[0].Name}},
		},
	}

	// Run test
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			segments, paging, err := svc.ListSegments(data.projectId, data.params)
			if data.expectedError == "" {
				assert.NoError(t, err)
				tu.AssertEqualValues(t, data.expectedSegments, segments)
				tu.AssertEqualValues(t, data.expectedPaging, paging)
			} else {
				assert.EqualError(t, err, data.expectedError)
			}
		})
	}
}

func testCreateUpdateDeleteSegment(s *SegmentServiceTestSuite) {
	t := s.Suite.T()
	svc := s.SegmentService

	// Create Segment
	updatedBy := "test-user"
	projectId := int64(1)
	segmentId := int64(3)
	stringSegmenter := []string{"seg-1"}
	integerSegmenter := []string{"4", "5"}
	expSegment := models.ExperimentSegment{
		"string_segmenter":  stringSegmenter,
		"integer_segmenter": integerSegmenter,
	}
	rawStringSegmenter := []interface{}{"seg-1"}
	rawIntegerSegmenter := []interface{}{float64(4), float64(5)}
	expSegmentRaw := models.ExperimentSegmentRaw{
		"string_segmenter":  rawStringSegmenter,
		"integer_segmenter": rawIntegerSegmenter,
	}

	// Create Segment
	createSegmentBody := services.CreateSegmentRequestBody{
		Name:      "test-segment-create",
		Segment:   expSegmentRaw,
		UpdatedBy: &updatedBy,
	}
	s.ValidationService.On("Validate", createSegmentBody).Return(nil)
	s.SegmenterService.On("ValidateExperimentSegment", int64(1), mock.Anything, mock.Anything).Return(nil)
	s.SegmenterService.
		On("GetSegmenterTypes", int64(1)).
		Return(
			map[string]schema.SegmenterType{
				"string_segmenter":  schema.SegmenterTypeString,
				"integer_segmenter": schema.SegmenterTypeInteger,
			},
			nil,
		)
	segmentResponse, err := svc.CreateSegment(s.Settings, createSegmentBody)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(t, models.Segment{
		ID: models.ID(segmentId),
		Model: models.Model{
			CreatedAt: segmentResponse.CreatedAt,
			UpdatedAt: segmentResponse.UpdatedAt,
		},
		ProjectID: models.ID(projectId),
		Name:      "test-segment-create",
		Segment:   expSegment,
		UpdatedBy: updatedBy,
	}, *segmentResponse)

	// Update Segment
	s.SegmentHistoryService.On("CreateSegmentHistory", segmentResponse).Return(nil, nil)
	newExpSegmentRaw := models.ExperimentSegmentRaw{"string_segmenter": rawStringSegmenter}
	newExpSegment := models.ExperimentSegment{"string_segmenter": stringSegmenter}
	updateSegmentBody := services.UpdateSegmentRequestBody{
		Segment:   newExpSegmentRaw,
		UpdatedBy: &updatedBy,
	}
	s.ValidationService.On("Validate", updateSegmentBody).Return(nil)
	s.SegmenterService.On("ValidateExperimentSegment", s.Settings.Config.Segmenters, newExpSegment).Return(nil)
	segmentResponse, err = svc.UpdateSegment(s.Settings, segmentId, updateSegmentBody)
	s.Suite.Require().NoError(err)
	s.Suite.Assert().Equal(models.ID(3), segmentResponse.ID)
	s.Suite.Assert().Equal(newExpSegment, segmentResponse.Segment)

	// Delete Segment
	deletedSegmentId := int64(3)
	err = svc.DeleteSegment(projectId, deletedSegmentId)
	s.Suite.Require().NoError(err)
}

func createTestSegments(db *gorm.DB) (models.Settings, []*models.Segment, error) {
	// Create test project settings (with project_id=1)
	var settings models.Settings
	err := db.Create(&models.Settings{
		ProjectID: models.ID(1),
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"string_segmenter", "integer_segmenter"},
				Variables: map[string][]string{
					"string_segmenter":  {"string_segmenter"},
					"integer_segmenter": {"integer_segmenter"},
				},
			},
		},
	}).Error
	if err != nil {
		return settings, []*models.Segment{}, err
	}
	// Query the created settings
	query := db.Where("project_id = 1").First(&settings)
	if err := query.Error; err != nil {
		return settings, []*models.Segment{}, err
	}

	// Define test segments
	stringSegmenter := []string{"seg-1", "seg-2"}
	integerSegmenter := []string{"1", "2"}
	segments := []models.Segment{
		{
			Model: models.Model{
				CreatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
			},
			ProjectID: models.ID(1),
			Name:      "test-segment-1",
			Segment:   models.ExperimentSegment{"string_segmenter": stringSegmenter, "integer_segmenter": integerSegmenter},
			UpdatedBy: "test-user-1",
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2021, 12, 1, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2021, 12, 1, 4, 5, 7, 0, time.UTC),
			},
			ProjectID: models.ID(1),
			Name:      "test-segment-2",
			Segment:   models.ExperimentSegment{"string_segmenter": stringSegmenter, "integer_segmenter": integerSegmenter},
			UpdatedBy: "test-user-2",
		},
	}

	// Create test segments
	for _, segment := range segments {
		err := db.Create(&segment).Error
		if err != nil {
			return settings, []*models.Segment{}, err
		}
	}

	// Return expected segment responses
	segmentRecords := []*models.Segment{&segments[0], &segments[1]}
	segmentRecords[0].ID = models.ID(1)
	segmentRecords[1].ID = models.ID(2)

	return settings, segmentRecords, nil
}
