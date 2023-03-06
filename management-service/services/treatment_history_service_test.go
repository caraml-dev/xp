//go:build integration

package services_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	tu "github.com/caraml-dev/xp/management-service/internal/testutils"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/pagination"
	"github.com/caraml-dev/xp/management-service/services"
)

type TreatmentHistoryServiceTestSuite struct {
	suite.Suite
	services.TreatmentHistoryService
	CleanUpFunc func()

	Treatments       []*models.Treatment
	TreatmentHistory []*models.TreatmentHistory
}

func (s *TreatmentHistoryServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up TreatentHistoryServiceTestSuite")

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB(tu.MigrationsPath)
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	// Init treatment history service
	s.TreatmentHistoryService = services.NewTreatmentHistoryService(db)

	// Create test data
	s.Treatments, s.TreatmentHistory, err = createTestTreatmentHistory(db)

	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func (s *TreatmentHistoryServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up TreatmentHistoryServiceTestSuite")
	s.CleanUpFunc()
}

func (s *TreatmentHistoryServiceTestSuite) TestTreatmmentHistoryServiceGetIntegration() {
	// Successful get
	histResponse, err := s.TreatmentHistoryService.GetTreatmentHistory(1, 1)
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.TreatmentHistory[0], histResponse)
	// Invalid version
	_, err = s.TreatmentHistoryService.GetTreatmentHistory(1, 200)
	s.Suite.Require().EqualError(err, "record not found")
}

func (s *TreatmentHistoryServiceTestSuite) TestTreatmentHistoryServiceListCreateIntegration() {
	// Test list treatment history first, since the create could affect the results
	testListTreatmentHistory(s)
	testCreateTreatmentHistory(s)
	testDeleteTreatmentHistory(s)
}

func testListTreatmentHistory(s *TreatmentHistoryServiceTestSuite) {
	histResponse, paging, err := s.TreatmentHistoryService.ListTreatmentHistory(1, services.ListTreatmentHistoryParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), s.TreatmentHistory, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 1, Pages: 1, Total: 2}, paging)
	// Pagination
	var page, pageSize int32 = 2, 1
	histResponse, paging, err = s.TreatmentHistoryService.ListTreatmentHistory(1, services.ListTreatmentHistoryParams{
		pagination.PaginationOptions{
			Page:     &page,
			PageSize: &pageSize,
		},
	})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), []*models.TreatmentHistory{s.TreatmentHistory[1]}, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 2, Pages: 2, Total: 2}, paging)
	// No history
	histResponse, paging, err = s.TreatmentHistoryService.ListTreatmentHistory(2, services.ListTreatmentHistoryParams{})
	s.Suite.Require().NoError(err)
	tu.AssertEqualValues(s.Suite.T(), []*models.TreatmentHistory{}, histResponse)
	tu.AssertEqualValues(s.Suite.T(), &pagination.Paging{Page: 1, Pages: 0, Total: 0}, paging)
}

func testCreateTreatmentHistory(s *TreatmentHistoryServiceTestSuite) {
	treatment := s.Treatments[1]
	treatmentHist, err := s.TreatmentHistoryService.CreateTreatmentHistory(treatment)

	s.Suite.Require().NoError(err)

	// Set up expected record
	expected := &models.TreatmentHistory{
		ID:          models.ID(3),
		TreatmentID: treatment.ID,
		Model: models.Model{
			CreatedAt: treatment.UpdatedAt,
			UpdatedAt: treatmentHist.UpdatedAt, // Copy the updated_at value from the new record
		},
		Version:       1,
		Name:          treatment.Name,
		Configuration: treatment.Configuration,
		UpdatedBy:     treatment.UpdatedBy,
	}
	expectedJSON, _ := json.Marshal(expected)

	// JSON Marshal to compare, to bypass timezone precision issues
	treatmentHistJSON, _ := json.Marshal(treatmentHist)
	s.Suite.Assert().JSONEq(string(expectedJSON), string(treatmentHistJSON))

	// Get the newly created treatment history and verify
	treatmentHist, err = s.TreatmentHistoryService.GetTreatmentHistory(int64(s.Treatments[1].ID), 1)
	s.Suite.Require().NoError(err)
	// JSON Marshal to compare, to bypass timezone precision issues
	treatmentHistJSON, _ = json.Marshal(treatmentHist)
	s.Suite.Assert().JSONEq(string(expectedJSON), string(treatmentHistJSON))
}

func testDeleteTreatmentHistory(s *TreatmentHistoryServiceTestSuite) {
	// Delete TreatmentHistory
	deletedTreatmentId := int64(3)
	err := s.TreatmentHistoryService.DeleteTreatmentHistory(deletedTreatmentId)
	s.Suite.Require().NoError(err)
}

func TestTreatmentHistoryService(t *testing.T) {
	suite.Run(t, new(TreatmentHistoryServiceTestSuite))
}

func createTestTreatmentHistory(db *gorm.DB) ([]*models.Treatment, []*models.TreatmentHistory, error) {
	// Create treatments (method reused from treatment service test)
	_, treatmentRecords, err := createTestTreatments(db)
	if err != nil {
		return []*models.Treatment{}, []*models.TreatmentHistory{}, err
	}

	// Create treatment history records, associated to the first treatment
	treatmentName := "treatment-hist"
	updatedBy := "test-updated-by"
	// Records have descending updated_at value, so the order would be maintained
	// by list treatment history API
	treatmentHistory := []models.TreatmentHistory{
		{
			Model: models.Model{
				CreatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2018, 1, 1, 2, 3, 5, 0, time.UTC),
			},
			TreatmentID:   treatmentRecords[0].ID,
			Version:       int64(1),
			Name:          treatmentName,
			Configuration: models.TreatmentConfig{},
			UpdatedBy:     updatedBy,
		},
		{
			Model: models.Model{
				CreatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2018, 1, 1, 2, 3, 4, 0, time.UTC),
			},
			TreatmentID:   treatmentRecords[0].ID,
			Version:       int64(2),
			Name:          treatmentName,
			Configuration: models.TreatmentConfig{},
			UpdatedBy:     updatedBy,
		},
	}

	// Create test treatment history
	for _, hist := range treatmentHistory {
		err := db.Create(&hist).Error
		if err != nil {
			return []*models.Treatment{}, []*models.TreatmentHistory{}, err
		}
	}

	// Set up expected treatment history data
	historyRecords := []*models.TreatmentHistory{
		&treatmentHistory[0],
		&treatmentHistory[1],
	}
	historyRecords[0].ID = models.ID(1)
	historyRecords[1].ID = models.ID(2)

	return treatmentRecords, historyRecords, nil
}
