//go:build integration

package services_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"

	tu "github.com/caraml-dev/xp/management-service/internal/testutils"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
	"github.com/caraml-dev/xp/management-service/services"
)

type SegmentHistoryServiceTestSuite struct {
	suite.Suite
	services.SegmentHistoryService
	CleanUpFunc func()

	Segments       []*models.Segment
	SegmentHistory []*models.SegmentHistory
}

func (s *SegmentHistoryServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up SegmentHistoryServiceTestSuite")

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB()
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	// Init segment history service
	s.SegmentHistoryService = services.NewSegmentHistoryService(db)

	// Create test data
	s.Segments, s.SegmentHistory, err = createTestSegmentHistory(db)

	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func (s *SegmentHistoryServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up SegmentHistoryServiceTestSuite")
	s.CleanUpFunc()
}

func (s *SegmentHistoryServiceTestSuite) TestSegmentHistoryServiceGetIntegration() {
	// Successful get
	histResponse, err := s.SegmentHistoryService.GetSegmentHistory(1, 1)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.SegmentHistory[0], histResponse)
	// Invalid version
	_, err = s.SegmentHistoryService.GetSegmentHistory(1, 200)
	s.Suite.Require().EqualError(err, "record not found")
}

func (s *SegmentHistoryServiceTestSuite) TestSegmentHistoryServiceListCreateIntegration() {
	// Test list segment history first, since the create could affect the results
	testListSegmentHistory(s)
	testCreateSegmentHistory(s)
	testDeleteSegmentHistory(s)
}

func testListSegmentHistory(s *SegmentHistoryServiceTestSuite) {
	histResponse, paging, err := s.SegmentHistoryService.ListSegmentHistory(1, services.ListSegmentHistoryParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.SegmentHistory, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 1, Pages: 1, Total: 2}, paging)
	// Pagination
	var page, pageSize int32 = 2, 1
	histResponse, paging, err = s.SegmentHistoryService.ListSegmentHistory(1, services.ListSegmentHistoryParams{
		pagination.PaginationOptions{
			Page:     &page,
			PageSize: &pageSize,
		},
	})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), []*models.SegmentHistory{s.SegmentHistory[1]}, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 2, Pages: 2, Total: 2}, paging)
	// No history
	histResponse, paging, err = s.SegmentHistoryService.ListSegmentHistory(2, services.ListSegmentHistoryParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), []*models.SegmentHistory{}, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 1, Pages: 0, Total: 0}, paging)
}

func testCreateSegmentHistory(s *SegmentHistoryServiceTestSuite) {
	segment := s.Segments[1]
	segmentHist, err := s.SegmentHistoryService.CreateSegmentHistory(segment)

	s.Suite.Require().NoError(err)

	// Set up expected record
	expected := &models.SegmentHistory{
		ID:        models.ID(3),
		SegmentID: segment.ID,
		Model: models.Model{
			CreatedAt: segment.UpdatedAt,
			UpdatedAt: segmentHist.UpdatedAt, // Copy the updated_at value from the new record
		},
		Version:   1,
		Name:      segment.Name,
		Segment:   segment.Segment,
		UpdatedBy: segment.UpdatedBy,
	}
	expectedJSON, _ := json.Marshal(expected)

	// JSON Marshal to compare, to bypass timezone precision issues
	segmentHistJSON, _ := json.Marshal(segmentHist)
	s.Suite.Assert().JSONEq(string(expectedJSON), string(segmentHistJSON))

	// Get the newly created segment history and verify
	segmentHist, err = s.SegmentHistoryService.GetSegmentHistory(int64(s.Segments[1].ID), 1)
	s.Suite.Require().NoError(err)
	// JSON Marshal to compare, to bypass timezone precision issues
	segmentHistJSON, _ = json.Marshal(segmentHist)
	s.Suite.Assert().JSONEq(string(expectedJSON), string(segmentHistJSON))
}

func testDeleteSegmentHistory(s *SegmentHistoryServiceTestSuite) {
	// Delete SegmentHistory
	deletedSegmentId := int64(3)
	err := s.SegmentHistoryService.DeleteSegmentHistory(deletedSegmentId)
	s.Suite.Require().NoError(err)
}

func TestSegmentHistoryService(t *testing.T) {
	suite.Run(t, new(SegmentHistoryServiceTestSuite))
}

func createTestSegmentHistory(db *gorm.DB) ([]*models.Segment, []*models.SegmentHistory, error) {
	// Create segments (method reused from segment service test)
	_, segmentRecords, err := createTestSegments(db)
	if err != nil {
		return []*models.Segment{}, []*models.SegmentHistory{}, err
	}

	// Create segment history records, associated to the first segment
	segmentName := "segment-hist"
	updatedBy := "test-updated-by"
	stringSegmenter := []string{"seg-1", "seg-2"}
	integerSegmenter := []string{"1", "2"}
	experimentSegment := models.ExperimentSegment{"string_segmenter": stringSegmenter, "integer_segmenter": integerSegmenter}
	// Records have descending updated_at value, so the order would be maintained
	// by list segment history API
	segmentHistory := []models.SegmentHistory{
		{
			Model: models.Model{
				CreatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2018, 1, 1, 2, 3, 5, 0, time.UTC),
			},
			SegmentID: segmentRecords[0].ID,
			Version:   int64(1),
			Name:      segmentName,
			Segment:   experimentSegment,
			UpdatedBy: updatedBy,
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
			},
			SegmentID: segmentRecords[0].ID,
			Version:   int64(2),
			Name:      segmentName,
			Segment:   experimentSegment,
			UpdatedBy: updatedBy,
		},
	}

	// Create test segment history
	for _, hist := range segmentHistory {
		err := db.Create(&hist).Error
		if err != nil {
			return []*models.Segment{}, []*models.SegmentHistory{}, err
		}
	}

	// Set up expected segment history data
	historyRecords := []*models.SegmentHistory{
		&segmentHistory[0],
		&segmentHistory[1],
	}
	historyRecords[0].ID = models.ID(1)
	historyRecords[1].ID = models.ID(2)

	return segmentRecords, historyRecords, nil
}
