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

type ExperimentHistoryServiceTestSuite struct {
	suite.Suite
	services.ExperimentHistoryService
	CleanUpFunc func()

	Experiments       []*models.Experiment
	ExperimentHistory []*models.ExperimentHistory
}

func (s *ExperimentHistoryServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ExperimentHistoryServiceTestSuite")

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB()
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	// Init experiment history service
	s.ExperimentHistoryService = services.NewExperimentHistoryService(db)

	// Create test data
	s.Experiments, s.ExperimentHistory, err = createTestExperimentHistory(db)

	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func (s *ExperimentHistoryServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up ExperimentHistoryServiceTestSuite")
	s.CleanUpFunc()
}

func (s *ExperimentHistoryServiceTestSuite) TestExperimentHistoryServiceGetIntegration() {
	// Successful get
	histResponse, err := s.ExperimentHistoryService.GetExperimentHistory(1, 1)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.ExperimentHistory[0], histResponse)
	// Invalid version
	_, err = s.ExperimentHistoryService.GetExperimentHistory(1, 200)
	s.Suite.Require().EqualError(err, "record not found")
}

func (s *ExperimentHistoryServiceTestSuite) TestExperimentHistoryServiceListCreateIntegration() {
	// Test list experiment history first, since the create could affect the results
	testListExperimentHistory(s)
	testCreateExperimentHistory(s)
}

func testListExperimentHistory(s *ExperimentHistoryServiceTestSuite) {
	histResponse, paging, err := s.ExperimentHistoryService.ListExperimentHistory(1, services.ListExperimentHistoryParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.ExperimentHistory, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 1, Pages: 1, Total: 2}, paging)
	// Pagination
	var page, pageSize int32 = 2, 1
	histResponse, paging, err = s.ExperimentHistoryService.ListExperimentHistory(1, services.ListExperimentHistoryParams{
		pagination.PaginationOptions{
			Page:     &page,
			PageSize: &pageSize,
		},
	})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), []*models.ExperimentHistory{s.ExperimentHistory[1]}, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 2, Pages: 2, Total: 2}, paging)
	// No history
	histResponse, paging, err = s.ExperimentHistoryService.ListExperimentHistory(2, services.ListExperimentHistoryParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), []*models.ExperimentHistory{}, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 1, Pages: 0, Total: 0}, paging)
}

func testCreateExperimentHistory(s *ExperimentHistoryServiceTestSuite) {
	experiment := s.Experiments[1]
	expHist, err := s.ExperimentHistoryService.CreateExperimentHistory(experiment)

	s.Suite.Require().NoError(err)

	// Set up expected record
	expected := &models.ExperimentHistory{
		ID:           models.ID(3),
		ExperimentID: experiment.ID,
		Model: models.Model{
			CreatedAt: experiment.UpdatedAt,
			UpdatedAt: expHist.UpdatedAt, // Copy the updated_at value from the new record
		},
		Version:     1,
		Description: experiment.Description,
		EndTime:     experiment.EndTime,
		Interval:    experiment.Interval,
		Name:        experiment.Name,
		Segment:     experiment.Segment,
		Status:      experiment.Status,
		Tier:        experiment.Tier,
		Treatments:  experiment.Treatments,
		Type:        experiment.Type,
		StartTime:   experiment.StartTime,
		UpdatedBy:   experiment.UpdatedBy,
	}
	expectedJSON, _ := json.Marshal(expected)

	// JSON Marshal to compare, to bypasss timezone precision issues
	expHistJSON, _ := json.Marshal(expHist)
	s.Suite.Assert().JSONEq(string(expectedJSON), string(expHistJSON))

	// Get the newly created experiment history and verify
	expHist, err = s.ExperimentHistoryService.GetExperimentHistory(int64(s.Experiments[1].ID), 1)
	s.Suite.Require().NoError(err)
	// JSON Marshal to compare, to bypasss timezone precision issues
	expHistJSON, _ = json.Marshal(expHist)
	s.Suite.Assert().JSONEq(string(expectedJSON), string(expHistJSON))
}

func TestExperimentHistoryService(t *testing.T) {
	suite.Run(t, new(ExperimentHistoryServiceTestSuite))
}

func createTestExperimentHistory(db *gorm.DB) ([]*models.Experiment, []*models.ExperimentHistory, error) {
	// Create experiments (method reused from experiment service test)
	_, expRecords, err := createTestExperiments(db)
	if err != nil {
		return []*models.Experiment{}, []*models.ExperimentHistory{}, err
	}

	// Create experiment history records, associated to the first experiment
	var testDescription1 string = "exp history desc 1"
	var testDescription2 string = "exp history desc 2"
	// Records have descending updated_at value, so the order would be maintained
	// by list experiment history API
	expHistory := []models.ExperimentHistory{
		{
			Model: models.Model{
				CreatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2018, 1, 1, 2, 3, 5, 0, time.UTC),
			},
			ExperimentID: expRecords[0].ID,
			Version:      int64(1),
			Name:         "exp-hist",
			Description:  &testDescription1,
			Tier:         models.ExperimentTierDefault,
			Type:         models.ExperimentTypeAB,
			Treatments:   nil,
			Segment:      models.ExperimentSegment{},
			Status:       models.ExperimentStatusInactive,
			EndTime:      time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC),
			StartTime:    time.Date(2022, 2, 2, 1, 1, 1, 0, time.UTC),
			UpdatedBy:    "test-updated-by",
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
			},
			ExperimentID: expRecords[0].ID,
			Version:      int64(2),
			Name:         "exp-hist",
			Description:  &testDescription2,
			Type:         models.ExperimentTypeAB,
			Tier:         models.ExperimentTierOverride,
			Treatments:   nil,
			Segment:      models.ExperimentSegment{},
			Status:       models.ExperimentStatusInactive,
			EndTime:      time.Date(2022, 1, 1, 1, 1, 1, 0, time.UTC),
			StartTime:    time.Date(2022, 2, 2, 1, 1, 1, 0, time.UTC),
			UpdatedBy:    "test-updated-by",
		},
	}

	// Create test experiment history
	for _, hist := range expHistory {
		err := db.Create(&hist).Error
		if err != nil {
			return []*models.Experiment{}, []*models.ExperimentHistory{}, err
		}
	}

	// Set up expected experiment history data
	historyRecords := []*models.ExperimentHistory{
		&expHistory[0],
		&expHistory[1],
	}
	historyRecords[0].ID = models.ID(1)
	historyRecords[1].ID = models.ID(2)

	return expRecords, historyRecords, nil
}
