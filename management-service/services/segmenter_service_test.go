package services_test

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/caraml-dev/xp/common/api/schema"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	tu "github.com/caraml-dev/xp/management-service/internal/testutils"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/segmenters"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/mocks"
)

var segmenterScopeGlobal = schema.SegmenterScopeGlobal
var segmenterScopeProject = schema.SegmenterScopeProject
var segmenterStatusActive = schema.SegmenterStatusActive
var segmenterStatusInactive = schema.SegmenterStatusInactive

type SegmenterServiceTestSuite struct {
	suite.Suite
	services.SegmenterService
	CleanUpFunc func()

	Settings         models.Settings
	GlobalSegmenters map[string]segmenters.Segmenter
	CustomSegmenters []models.CustomSegmenter
}

func (s *SegmenterServiceTestSuite) SetupSuite() {
	segmenterConfig := map[string]interface{}{
		"s2_ids": map[string]interface{}{
			"mins2celllevel": 14,
			"maxs2celllevel": 15,
		},
	}

	// Create test DB, save the DB clean up function to be executed on tear down
	db, cleanup, err := tu.CreateTestDB()
	if err != nil {
		s.Suite.T().Fatalf("Could not create test DB: %v", err)
	}
	s.CleanUpFunc = cleanup

	settingsSvc := mocks.ProjectSettingsService{}
	// This will define which segmenters are active
	settingsSvc.On("GetDBRecord", models.ID(0)).Return(
		&models.Settings{
			ProjectID: models.ID(0),
			Config: &models.ExperimentationConfig{Segmenters: models.ProjectSegmenters{
				Names: []string{
					"hours_of_day",
					"days_of_week",
					"country",
					"area",
					"bool_segmenter",
					"s2_ids",
				},
			}},
		},
		nil,
	)
	settingsSvc.On("GetDBRecord", mock.Anything).Return(
		&models.Settings{
			ProjectID: models.ID(1),
			Config: &models.ExperimentationConfig{Segmenters: models.ProjectSegmenters{
				Names: []string{"s2_ids", "country", "test-custom-segmenter-1"}}}},
		nil,
	)

	pubSubSvc := &mocks.PubSubPublisherService{}
	pubSubSvc.On("PublishProjectSegmenterMessage", mock.Anything, mock.Anything, mock.Anything).Return(
		nil)
	allServices := &services.Services{
		ProjectSettingsService: &settingsSvc,
		PubSubPublisherService: pubSubSvc,
	}

	s.SegmenterService, err = services.NewSegmenterService(allServices, segmenterConfig, db)
	if err != nil {
		s.T().Fatalf("failed to start segmenter service: %s", err)
	}

	// Create test data
	s.Settings, s.CustomSegmenters, err = createTestCustomSegmenters(db)
	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
	s.GlobalSegmenters, err = createTestGlobalSegmenters(segmenterConfig)
	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}
}

func TestSegmenterService(t *testing.T) {
	suite.Run(t, new(SegmenterServiceTestSuite))
}

func (s *SegmenterServiceTestSuite) TestGetSegmenterConfigurations() {
	daysOfWeekSegmenter := &[]_segmenters.SegmenterConfiguration{
		{
			Name: "days_of_week",
			Type: 2,
			Options: map[string]*_segmenters.SegmenterValue{
				"Monday":    {Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
				"Tuesday":   {Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
				"Wednesday": {Value: &_segmenters.SegmenterValue_Integer{Integer: 3}},
				"Thursday":  {Value: &_segmenters.SegmenterValue_Integer{Integer: 4}},
				"Friday":    {Value: &_segmenters.SegmenterValue_Integer{Integer: 5}},
				"Saturday":  {Value: &_segmenters.SegmenterValue_Integer{Integer: 6}},
				"Sunday":    {Value: &_segmenters.SegmenterValue_Integer{Integer: 7}},
			},
			MultiValued: true,
			TreatmentRequestFields: &_segmenters.ListExperimentVariables{
				Values: []*_segmenters.ExperimentVariables{
					{
						Value: []string{"tz"},
					},
				},
			},
			Constraints: nil,
		},
	}
	s2IdSegmenter := &[]_segmenters.SegmenterConfiguration{
		{
			Name:        "s2_ids",
			Type:        2,
			Options:     map[string]*_segmenters.SegmenterValue{},
			MultiValued: true,
			TreatmentRequestFields: &_segmenters.ListExperimentVariables{
				Values: []*_segmenters.ExperimentVariables{
					{
						Value: []string{"s2id"},
					},
					{
						Value: []string{"latitude", "longitude"},
					},
				},
			},
			Constraints: nil,
		},
	}
	boolSegmenter := &[]_segmenters.SegmenterConfiguration{
		{
			Name: "bool_segmenter",
			Type: 3,
			Options: map[string]*_segmenters.SegmenterValue{
				"Yes": {Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
				"No":  {Value: &_segmenters.SegmenterValue_Bool{Bool: false}},
			},
			MultiValued: true,
			TreatmentRequestFields: &_segmenters.ListExperimentVariables{
				Values: []*_segmenters.ExperimentVariables{
					{
						Value: []string{"bool_segmenter"},
					},
				},
			},
			Constraints: nil,
		},
	}

	tests := map[string]struct {
		segmenterNames []string
		expectedFields *[]_segmenters.SegmenterConfiguration
		errString      string
	}{
		"success | segmenter names": {
			segmenterNames: []string{"s2_ids"},
			expectedFields: s2IdSegmenter,
		},
		"success | time fields": {
			segmenterNames: []string{"days_of_week"},
			expectedFields: daysOfWeekSegmenter,
		},
		"success | boolean values": {
			segmenterNames: []string{"bool_segmenter"},
			expectedFields: boolSegmenter,
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			got, err := s.SegmenterService.GetSegmenterConfigurations(int64(0), data.segmenterNames)
			s.Suite.Require().NoError(err)
			s.Suite.Assert().True(true, reflect.DeepEqual(data.expectedFields, got))
		})
	}
}

func (s *SegmenterServiceTestSuite) TestGetFormattedSegmenters() {
	s2IdFloat := []interface{}{float64(3592210809859604480)}
	daysOfWeekFloat := []interface{}{float64(1), float64(2)}
	countryString := []interface{}{"ID", "SG"}
	boolValues := []interface{}{false, true}

	s2IdResp := []interface{}{int64(3592210809859604480)}
	daysOfWeekResp := []interface{}{int64(1), int64(2)}
	boolValuesResp := []interface{}{false, true}
	countryWithQuotesResp := []interface{}{"\"ID\"", "\"SG\""}

	tests := map[string]struct {
		segmenters models.ExperimentSegmentRaw
		expected   map[string]*[]interface{}
		errString  string
	}{
		"success | s2_ids": {
			segmenters: models.ExperimentSegmentRaw{
				"s2_ids": s2IdFloat,
			},
			expected: map[string]*[]interface{}{
				"s2_ids": &s2IdResp,
			},
		},
		"success | time": {
			segmenters: models.ExperimentSegmentRaw{
				"days_of_week": daysOfWeekFloat,
			},
			expected: map[string]*[]interface{}{
				"days_of_week": &daysOfWeekResp,
			},
		},
		"success | multiple field types": {
			segmenters: models.ExperimentSegmentRaw{
				"country":        countryString,
				"days_of_week":   daysOfWeekFloat,
				"bool_segmenter": boolValues,
			},
			expected: map[string]*[]interface{}{
				"country":        &countryWithQuotesResp,
				"days_of_week":   &daysOfWeekResp,
				"bool_segmenter": &boolValuesResp,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			got, err := s.SegmenterService.GetFormattedSegmenters(int64(0), data.segmenters)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
				s.Suite.Assert().Equal(data.expected, got)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *SegmenterServiceTestSuite) TestValidateExperimentSegment() {
	countryId := []interface{}{"ID"}
	countries := []interface{}{"ID", "SG"}
	areasId := []interface{}{float64(1), float64(2)}
	areasSg := []interface{}{float64(3)}
	s2IdsValid := []interface{}{float64(3592184395810734080)}

	tests := map[string]struct {
		userSegmenters []string
		expSegment     models.ExperimentSegmentRaw
		errString      string
	}{
		"success | absent segmenters": {
			userSegmenters: []string{"country", "area"},
			expSegment: models.ExperimentSegmentRaw{
				"country": countryId,
			},
		},
		"failure | single value check failed": {
			userSegmenters: []string{"area", "country"},
			expSegment: models.ExperimentSegmentRaw{
				"area":    areasId,
				"country": countries,
			},
			errString: "Segmenter country is configured as single-valued but has multiple input values",
		},
		"failure | constraint failed": {
			userSegmenters: []string{"country", "area"},
			expSegment: models.ExperimentSegmentRaw{
				"country": countryId,
				"area":    areasSg,
			},
			errString: "Values for segmenter area do not satisfy the constraint",
		},
		"success": {
			userSegmenters: []string{"country", "area", "s2_ids"},
			expSegment: models.ExperimentSegmentRaw{
				"country": countryId,
				"area":    areasId,
				"s2_ids":  s2IdsValid,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidateExperimentSegment(int64(0), data.userSegmenters, data.expSegment)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *SegmenterServiceTestSuite) TestValidateSegmentOrthogonality() {
	s2IdRaw := []interface{}{float64(3592210809859604480), float64(3592210814154571776)}
	daysOfWeekRaw := []interface{}{float64(1)}
	testCountriesRaw1 := []interface{}{"SG", "ID"}
	invalidTestAreasRaw1 := []interface{}{"invalid"}

	testS2Id1 := []string{"3592210809859604480"}
	testS2Id2 := []string{"3592210814154571776"}
	testDaysOfWeek1 := []string{"1"}
	testDaysOfWeek2 := []string{"2"}
	testDaysOfWeek3 := []string{"3"}
	testDaysOfWeek4 := []string{"1", "4"}
	testCountries1 := []string{"SG", "ID"}
	testCountries2 := []string{"SG"}
	testAreas1 := []string{"2"}
	testAreas2 := []string{"1", "4"}
	tests := map[string]struct {
		userSegmenters []string
		expSegment     models.ExperimentSegmentRaw
		allExps        []models.Experiment
		errString      string
	}{
		"failure | invalid experiment segment values": {
			userSegmenters: []string{"country", "area"},
			expSegment: models.ExperimentSegmentRaw{
				"country": testCountriesRaw1,
				"area":    invalidTestAreasRaw1,
			},
			allExps: []models.Experiment{
				{
					Segment: models.ExperimentSegment{
						"country": testCountries1,
						"area":    testAreas1,
					},
				},
				{
					Segment: models.ExperimentSegment{
						"country": testCountries2,
						"area":    testAreas2,
					},
				},
			},
			errString: "received wrong type of segmenter value; area expects type integer",
		},
		"failure | overlap": {
			userSegmenters: []string{"s2_ids", "days_of_week"},
			expSegment: models.ExperimentSegmentRaw{
				"s2_ids":       s2IdRaw,
				"days_of_week": daysOfWeekRaw,
			},
			allExps: []models.Experiment{
				{
					Segment: models.ExperimentSegment{
						"s2_ids":       testS2Id1,
						"days_of_week": testDaysOfWeek2,
					},
				},
				{
					Segment: models.ExperimentSegment{
						"s2_ids":       testS2Id2,
						"days_of_week": testDaysOfWeek4,
					},
				},
			},
			errString: "Segment Orthogonality check failed against experiment ID 0",
		},
		"failure | both segmenters optional": {
			userSegmenters: []string{"s2_ids", "days_of_week"},
			expSegment: models.ExperimentSegmentRaw{
				"s2_ids": s2IdRaw,
			},
			allExps: []models.Experiment{
				{
					Segment: models.ExperimentSegment{
						"s2_ids":       testS2Id1,
						"days_of_week": testDaysOfWeek1,
					},
				},
				{
					Segment: models.ExperimentSegment{
						"s2_ids": testS2Id1,
					},
				},
			},
			errString: "Segment Orthogonality check failed against experiment ID 0",
		},
		"success | existing segmenter optional": {
			userSegmenters: []string{"s2_ids", "days_of_week"},
			expSegment: models.ExperimentSegmentRaw{
				"s2_ids":       s2IdRaw,
				"days_of_week": daysOfWeekRaw,
			},
			allExps: []models.Experiment{
				{
					Segment: models.ExperimentSegment{
						"s2_ids": testS2Id1,
					},
				},
			},
		},
		"success | new segmenter optional": {
			userSegmenters: []string{"s2_ids", "days_of_week"},
			expSegment: models.ExperimentSegmentRaw{
				"s2_ids": s2IdRaw,
			},
			allExps: []models.Experiment{
				{
					Segment: models.ExperimentSegment{
						"s2_ids":       testS2Id1,
						"days_of_week": testDaysOfWeek1,
					},
				},
			},
		},
		"success | no overlap": {
			userSegmenters: []string{"s2_ids", "days_of_week"},
			expSegment: models.ExperimentSegmentRaw{
				"s2_ids":       s2IdRaw,
				"days_of_week": daysOfWeekRaw,
			},
			allExps: []models.Experiment{
				{
					Segment: models.ExperimentSegment{
						"s2_ids":       testS2Id1,
						"days_of_week": testDaysOfWeek2,
					},
				},
				{
					Segment: models.ExperimentSegment{
						"s2_ids":       testS2Id2,
						"days_of_week": testDaysOfWeek3,
					},
				},
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidateSegmentOrthogonality(int64(0), data.userSegmenters, data.expSegment, data.allExps)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *SegmenterServiceTestSuite) TestValidateRequiredSegmenters() {
	tests := map[string]struct {
		projectId          int64
		providedSegmenters []string
		errString          string
	}{
		"failure | error retrieving segmenter types for given project id": {
			projectId: int64(-99999999999999999),
			errString: "pq: value \"-99999999999999999\" is out of range for type integer",
		},
		"failure | required custom segmenter not chosen": {
			projectId:          int64(1),
			providedSegmenters: []string{"country", "area"},
			errString:          "segmenter test-custom-segmenter-1 is a required segmenter that must be chosen",
		},
		"success": {
			projectId:          int64(0),
			providedSegmenters: []string{"country", "area"},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidateRequiredSegmenters(data.projectId, data.providedSegmenters)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *SegmenterServiceTestSuite) TestValidateDependentSegmenters() {
	tests := map[string]struct {
		providedSegmenters []string
		errString          string
	}{
		"failure | missing dependent segmenter": {
			providedSegmenters: []string{"area"},
			errString:          "segmenter area requires country to also be chosen",
		},
		"success": {
			providedSegmenters: []string{"country", "area"},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidatePrereqSegmenters(int64(0), data.providedSegmenters)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *SegmenterServiceTestSuite) TestValidateExperimentVariables() {
	tests := map[string]struct {
		projectSegmenters models.ProjectSegmenters
		errString         string
	}{
		"success | no segmenters": {
			projectSegmenters: models.ProjectSegmenters{},
		},
		"success | with segmenter": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"area"},
				Variables: map[string][]string{
					"area": {"area"},
				},
			},
		},
		"success | with segmenter indirect variable mapping": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"area"},
				Variables: map[string][]string{
					"area": {"latitude", "longitude"},
				},
			},
		},
		"failure | missing dependent segmenter": {
			projectSegmenters: models.ProjectSegmenters{
				Names:     []string{"area"},
				Variables: map[string][]string{},
			},
			errString: "len of project segmenters does not match mapping of experiment variables",
		},
		"failure | invalid experiment variables mapping": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"area"},
				Variables: map[string][]string{
					"country": {"latitude", "longitude"},
				},
			},
			errString: "project segmenters does not match mapping of experiment variables",
		},
		"failure | invalid segmenter": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"abc"},
				Variables: map[string][]string{
					"abc": {"a"},
				},
			},
			errString: "unknown segmenter: abc",
		},
		"failure | invalid variables": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"area"},
				Variables: map[string][]string{
					"area": {"latitude", "longitude", "unknown"},
				},
			},
			errString: "segmenter (area) does not have valid experiment variable(s) provided",
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidateExperimentVariables(int64(0), data.projectSegmenters)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

// Requires s2id, days_of_week, test-custom-segmenter-1,
// test-custom-segmenter-2 segmenter to be registered
// "s2_id", "test-custom-segmenter-1" set as active segmenters
func (s *SegmenterServiceTestSuite) TestGetSegmenter() {
	activeGlobalSegmenterConfig, err := s.GlobalSegmenters["s2_ids"].GetConfiguration()
	s.Suite.Assert().NoError(err)
	activeGlobalSegmenter, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(activeGlobalSegmenterConfig)
	s.Suite.Assert().NoError(err)
	activeGlobalSegmenter.Status = &segmenterStatusActive
	activeGlobalSegmenter.Scope = &segmenterScopeGlobal

	inactiveGlobalSegmenterConfig, err := s.GlobalSegmenters["days_of_week"].GetConfiguration()
	s.Suite.Assert().NoError(err)
	inactiveGlobalSegmenter, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(inactiveGlobalSegmenterConfig)
	s.Suite.Assert().NoError(err)
	inactiveGlobalSegmenter.Status = &segmenterStatusInactive
	inactiveGlobalSegmenter.Scope = &segmenterScopeGlobal

	activeCustomSegmenter := s.CustomSegmenters[0].ToApiSchema()
	activeCustomSegmenter.Status = &segmenterStatusActive
	activeCustomSegmenter.Scope = &segmenterScopeProject

	inactiveCustomSegmenter1 := s.CustomSegmenters[1].ToApiSchema()
	inactiveCustomSegmenter1.Status = &segmenterStatusInactive
	inactiveCustomSegmenter1.Scope = &segmenterScopeProject

	inactiveCustomSegmenter2 := s.CustomSegmenters[2].ToApiSchema()
	inactiveCustomSegmenter2.Status = &segmenterStatusInactive
	inactiveCustomSegmenter2.Scope = &segmenterScopeProject

	inactiveCustomSegmenter3 := s.CustomSegmenters[3].ToApiSchema()
	inactiveCustomSegmenter3.Status = &segmenterStatusInactive
	inactiveCustomSegmenter3.Scope = &segmenterScopeProject

	tests := map[string]struct {
		segmenterName string
		expected      *schema.Segmenter
		errString     string
	}{
		"Success | Global Segmenter Active": {
			segmenterName: "s2_ids",
			expected:      activeGlobalSegmenter,
			errString:     "",
		},
		"Success | Global Segmenter Inactive": {
			segmenterName: "days_of_week",
			expected:      inactiveGlobalSegmenter,
			errString:     "",
		},
		"Success | Project Segmenter 0 Active": {
			segmenterName: "test-custom-segmenter-1",
			expected:      &activeCustomSegmenter,
			errString:     "",
		},
		"Success | Project Segmenter 1 Inactive": {
			segmenterName: "test-custom-segmenter-2",
			expected:      &inactiveCustomSegmenter1,
			errString:     "",
		},
		"Success | Project Segmenter 2 Inactive": {
			segmenterName: "test-custom-segmenter-3",
			expected:      &inactiveCustomSegmenter2,
			errString:     "",
		},
		"Success | Project Segmenter 3 Inactive": {
			segmenterName: "test-custom-segmenter-4",
			expected:      &inactiveCustomSegmenter3,
			errString:     "",
		},
		"Fail | Unknown Segmenter": {
			segmenterName: "randomsegmenter",
			errString:     "unknown segmenter: randomsegmenter",
		},
	}
	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			got, err := s.SegmenterService.GetSegmenter(1, data.segmenterName)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
				s.Suite.Assert().NoError(err)
				tu.AssertEqualValues(s.Suite.T(), data.expected, got)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

// Requires s2id, test-custom-segmenter-2 to be registered
// "s2_id", "test-custom-segmenter-1" set as active segmenters
func (s *SegmenterServiceTestSuite) TestListSegmenters() {

	s2idConfig, err := s.GlobalSegmenters["s2_ids"].GetConfiguration()
	s.Suite.Assert().NoError(err)
	s2idSegmenter, err := segmenters.ProtobufSegmenterConfigToOpenAPISegmenterConfig(s2idConfig)
	s2idSegmenter.Status = &segmenterStatusActive
	s2idSegmenter.Scope = &segmenterScopeGlobal

	testCustomSegmenter2 := s.CustomSegmenters[1].ToApiSchema()
	testCustomSegmenter2.Status = &segmenterStatusInactive
	testCustomSegmenter2.Scope = &segmenterScopeProject

	s.Suite.Assert().NoError(err)

	segmenterScopeGlobal := services.SegmenterScopeGlobal
	segmenterScopeProject := services.SegmenterScopeProject
	segmenterStatusInactive := services.SegmenterStatusInactive
	segmenterStatusActive := services.SegmenterStatusActive

	searchString := "2"

	tests := map[string]struct {
		param          services.ListSegmentersParams
		expectedSubset []*schema.Segmenter
		expectedLength int
		errString      string
	}{
		"Success | No Param": {
			expectedSubset: []*schema.Segmenter{s2idSegmenter, &testCustomSegmenter2},
			expectedLength: len(s.GlobalSegmenters) + len(s.CustomSegmenters),
			errString:      "",
		},
		"Success | Global Segmenter": {
			expectedSubset: []*schema.Segmenter{s2idSegmenter},
			expectedLength: len(s.GlobalSegmenters),
			errString:      "",
			param: services.ListSegmentersParams{
				Scope: &segmenterScopeGlobal,
			},
		},
		"Success | Project Segmenter": {
			expectedSubset: []*schema.Segmenter{&testCustomSegmenter2},
			expectedLength: len(s.CustomSegmenters),
			errString:      "",
			param: services.ListSegmentersParams{
				Scope: &segmenterScopeProject,
			},
		},
		"Success | Active Segmenter": {
			expectedSubset: []*schema.Segmenter{s2idSegmenter},
			// base on suite setup mock response from project settings
			expectedLength: 3,
			errString:      "",
			param: services.ListSegmentersParams{
				Status: &segmenterStatusActive,
			},
		},
		"Success | Inactive Segmenter": {
			expectedSubset: []*schema.Segmenter{&testCustomSegmenter2},
			expectedLength: len(s.GlobalSegmenters) + len(s.CustomSegmenters) - 3,
			errString:      "",
			param: services.ListSegmentersParams{
				Status: &segmenterStatusInactive,
			},
		},
		"Success | Active Global Segmenter": {
			expectedSubset: []*schema.Segmenter{s2idSegmenter},
			// base on suite setup mock response from project settings
			expectedLength: 2,
			errString:      "",
			param: services.ListSegmentersParams{
				Scope:  &segmenterScopeGlobal,
				Status: &segmenterStatusActive,
			},
		},
		"Success | Inactive Project Segmenter": {
			expectedSubset: []*schema.Segmenter{&testCustomSegmenter2},
			// base on suite setup mock response from project settings
			expectedLength: len(s.CustomSegmenters) - 1,
			errString:      "",
			param: services.ListSegmentersParams{
				Scope:  &segmenterScopeProject,
				Status: &segmenterStatusInactive,
			},
		},
		"Success | Search": {
			expectedSubset: []*schema.Segmenter{&testCustomSegmenter2, s2idSegmenter},
			expectedLength: 3,
			errString:      "",
			param: services.ListSegmentersParams{
				Search: &searchString,
			},
		},
	}
	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			got, err := s.SegmenterService.ListSegmenters(1, data.param)
			s.Suite.Require().Equal(data.expectedLength, len(got))
			if data.errString == "" {
				s.Suite.Require().NoError(err)
				for _, segmenter := range data.expectedSubset {
					found := false
					for _, fetchedSegmenter := range got {
						if segmenter.Name == fetchedSegmenter.Name {
							found = true
							tu.AssertEqualValues(s.Suite.T(), segmenter, fetchedSegmenter)
						}
					}
					s.Suite.Require().True(found)
				}
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

// Requires test-create-segmenter-1 to NOT be registered
func (s *SegmenterServiceTestSuite) TestCreateSegmenter() {
	projectId := 1
	segmenterName := "test-create-segmenter-1"
	segmenterType := _segmenters.SegmenterValueType_BOOL.String()
	description := "Just a test segmenter"
	var constraints *models.Constraints

	// Failure flow
	invalidSegmenterOptions := models.Options{
		"InvalidYes": "true", // validation checks occur on these 2 keys in no particular order
		"InvalidNo":  "true", // the value for this key is the same as above to prevent non-deterministic test behaviour
	}
	invalidTestRequest := services.CreateCustomSegmenterRequestBody{
		Name:        segmenterName,
		Type:        segmenterType,
		Options:     &invalidSegmenterOptions,
		MultiValued: false,
		Constraints: nil,
		Required:    false,
		Description: &description,
	}
	segmenterResponse, err := s.SegmenterService.CreateCustomSegmenter(int64(projectId), invalidTestRequest)
	s.Suite.Assert().Nil(segmenterResponse)
	s.Suite.Assert().EqualError(err, "received wrong type of segmenter value; true expects type bool")

	// Success flow
	validSegmenterOptions := models.Options{
		"Yes": true,
		"No":  false,
	}

	validTestRequest := services.CreateCustomSegmenterRequestBody{
		Name:        segmenterName,
		Type:        segmenterType,
		Options:     &validSegmenterOptions,
		MultiValued: false,
		Constraints: nil,
		Required:    false,
		Description: &description,
	}

	segmenterResponse, err = s.SegmenterService.CreateCustomSegmenter(int64(projectId), validTestRequest)
	s.Suite.Require().NoError(err)
	s.Suite.Assert().Equal(segmenterName, segmenterResponse.Name)
	s.Suite.Assert().Equal(models.ID(projectId), segmenterResponse.ProjectID)
	s.Suite.Assert().Equal(segmenterType, string(segmenterResponse.Type))
	s.Suite.Assert().Equal(validSegmenterOptions, *segmenterResponse.Options)
	s.Suite.Assert().Equal(false, segmenterResponse.MultiValued)
	s.Suite.Assert().Equal(constraints, segmenterResponse.Constraints)
	s.Suite.Assert().Equal(false, segmenterResponse.Required)
	s.Suite.Assert().Equal(&description, segmenterResponse.Description)

	// create again and expect error
	_, err = s.SegmenterService.CreateCustomSegmenter(int64(projectId), validTestRequest)
	s.Suite.Assert().EqualError(err, "a segmenter with the name test-create-segmenter-1 already exists")

}

// Requires test-custom-segmenter-for-update to be registered
func (s *SegmenterServiceTestSuite) TestUpdateSegmenter() {
	projectId := 1
	segmenterName := "test-custom-segmenter-for-update"
	description := "Just a test segmenter"
	var constraints *models.Constraints
	// Failure flow
	invalidSegmenterOptions := models.Options{
		"InvalidYes": 1, // validation checks occur on these 2 keys in no particular order
		"InvalidNo":  1, // the value for this key is the same as above to prevent non-deterministic test behaviour
	}
	invalidTestRequest := services.UpdateCustomSegmenterRequestBody{
		Options:     &invalidSegmenterOptions,
		MultiValued: false,
		Constraints: nil,
		Required:    false,
		Description: &description,
	}
	segmenterResponse, err := s.SegmenterService.UpdateCustomSegmenter(int64(projectId), segmenterName, invalidTestRequest)
	s.Suite.Assert().Nil(segmenterResponse)
	s.Suite.Assert().EqualError(err, "received wrong type of segmenter value; %!s(int=1) expects type real")

	// Success flow
	validSegmenterOptions := models.Options{
		"NewYes": 1.0,
		"NewNo":  0.0,
	}
	validTestRequest := services.UpdateCustomSegmenterRequestBody{
		Options:     &validSegmenterOptions,
		MultiValued: false,
		Constraints: nil,
		Required:    false,
		Description: &description,
	}

	segmenterResponse, err = s.SegmenterService.UpdateCustomSegmenter(int64(projectId), segmenterName, validTestRequest)
	s.Suite.Require().NoError(err)
	s.Suite.Assert().Equal(segmenterName, segmenterResponse.Name)
	s.Suite.Assert().Equal(models.ID(projectId), segmenterResponse.ProjectID)
	s.Suite.Assert().Equal(validSegmenterOptions, *segmenterResponse.Options)
	s.Suite.Assert().Equal(false, segmenterResponse.MultiValued)
	s.Suite.Assert().Equal(constraints, segmenterResponse.Constraints)
	s.Suite.Assert().Equal(false, segmenterResponse.Required)
	s.Suite.Assert().Equal(&description, segmenterResponse.Description)

	// update non existing segmenter and expect error
	_, err = s.SegmenterService.UpdateCustomSegmenter(int64(projectId), "non-existence!", validTestRequest)
	s.Suite.Assert().EqualError(err, "unknown segmenter: non-existence!")
}

// Requires test-custom-segmenter-for-delete to be registered
// "test-custom-segmenter-1" set as active segmenters
func (s *SegmenterServiceTestSuite) TestDeleteSegmenter() {
	// Success flow
	err := s.SegmenterService.DeleteCustomSegmenter(int64(1), "test-custom-segmenter-for-delete")
	s.Suite.Require().NoError(err)

	// Error Active Segmenter
	err = s.SegmenterService.DeleteCustomSegmenter(int64(1), "test-custom-segmenter-1")
	s.Suite.Assert().EqualError(err,
		"custom segmenter: test-custom-segmenter-1 is currently in use in the project settings and cannot be deleted")

	// Error Non existence Segmenter
	err = s.SegmenterService.DeleteCustomSegmenter(int64(1), "unknown")
	s.Suite.Assert().EqualError(err, "unknown segmenter: unknown")
}

func createTestGlobalSegmenters(segmenterConfig map[string]interface{}) (map[string]segmenters.Segmenter, error) {
	globalSegmenters := make(map[string]segmenters.Segmenter)
	for name := range segmenters.Segmenters {
		if _, ok := segmenterConfig[name]; ok {
			configJSON, err := json.Marshal(segmenterConfig[name])
			if err != nil {
				return nil, err
			}

			m, err := segmenters.Get(name, configJSON)
			if err != nil {
				return nil, err
			}
			globalSegmenters[name] = m
			continue
		}
		m, err := segmenters.Get(name, nil)
		if err != nil {
			return nil, err
		}
		globalSegmenters[name] = m
	}
	return globalSegmenters, nil
}

func createTestCustomSegmenters(db *gorm.DB) (models.Settings, []models.CustomSegmenter, error) {
	// Create test project settings (with project_id=1)
	var settings models.Settings
	err := db.Create(&models.Settings{
		ProjectID: models.ID(1),
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"seg-1"},
				Variables: map[string][]string{
					"seg-1": {"exp-var-1"},
				},
			},
		},
	}).Error
	if err != nil {
		return settings, []models.CustomSegmenter{}, err
	}
	// Query the created settings
	query := db.Where("project_id = 1").First(&settings)
	if err := query.Error; err != nil {
		return settings, []models.CustomSegmenter{}, err
	}

	testDescription1 := "test-custom-segmenter: string"
	testDescription2 := "test-custom-segmenter: bool"
	testDescription3 := "test-custom-segmenter: real"
	testDescription4 := "test-custom-segmenter: integer"

	testSegmenterTypesMap := make(map[string]schema.SegmenterType)
	testSegmenterTypesMap["prerequisite_1"] = schema.SegmenterTypeString
	testSegmenterTypesMap["prerequisite_2"] = schema.SegmenterTypeBool

	// Define test custom segmenters
	customSegmenters := []models.CustomSegmenter{
		{
			ProjectID:   models.ID(1),
			Name:        "test-custom-segmenter-1",
			Type:        models.SegmenterValueTypeString,
			Description: &testDescription1,
			Required:    true,
			MultiValued: true,
			Options: &models.Options{
				"option_1": "string_1",
				"option_2": "string_2",
			},
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
		{
			ProjectID:   models.ID(1),
			Name:        "test-custom-segmenter-2",
			Type:        models.SegmenterValueTypeBool,
			Description: &testDescription2,
			Required:    false,
			MultiValued: true,
			Options: &models.Options{
				"option_1": "true",
				"option_2": "false",
			},
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
		{
			ProjectID:   models.ID(1),
			Name:        "test-custom-segmenter-3",
			Type:        models.SegmenterValueTypeReal,
			Description: &testDescription3,
			Required:    false,
			MultiValued: true,
			Options: &models.Options{
				"option_1": "0.0",
				"option_2": "1.1",
			},
			Constraints: &models.Constraints{
				models.Constraint{
					PreRequisites: []models.PreRequisite{
						{
							SegmenterName:   "prerequisite_1",
							SegmenterValues: []interface{}{"abc", "def"},
						},
					},
					AllowedValues: []interface{}{"2.1", "2.2"},
					Options: &models.Options{
						"constraint_option_1": "3.1",
						"constraint_option_2": "3.2",
					},
				},
			},
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
		{
			ProjectID:   models.ID(1),
			Name:        "test-custom-segmenter-4",
			Type:        models.SegmenterValueTypeInteger,
			Description: &testDescription4,
			Required:    false,
			MultiValued: true,
			Options: &models.Options{
				"option_1": "1",
				"option_2": "2",
			},
			Constraints: &models.Constraints{
				models.Constraint{
					PreRequisites: []models.PreRequisite{
						{
							SegmenterName:   "prerequisite_1",
							SegmenterValues: []interface{}{"true", "false"},
						},
					},
					AllowedValues: []interface{}{"21", "22"},
					Options: &models.Options{
						"constraint_option_1": "31",
						"constraint_option_2": "32",
					},
				},
			},
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
		{
			ProjectID:   models.ID(1),
			Name:        "test-custom-segmenter-for-update",
			Type:        models.SegmenterValueTypeReal,
			Description: &testDescription3,
			Required:    false,
			MultiValued: true,
			Options: &models.Options{
				"option_1": "0.0",
				"option_2": "1.1",
			},
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
		{
			ProjectID:   models.ID(1),
			Name:        "test-custom-segmenter-for-delete",
			Type:        models.SegmenterValueTypeInteger,
			Description: &testDescription4,
			Required:    false,
			MultiValued: true,
			Options: &models.Options{
				"option_1": "1",
				"option_2": "2",
			},
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
		{
			ProjectID:   models.ID(1),
			Name:        "prerequisite_1",
			Type:        models.SegmenterValueTypeString,
			Required:    false,
			MultiValued: false,
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
		{
			ProjectID:   models.ID(1),
			Name:        "prerequisite_2",
			Type:        models.SegmenterValueTypeBool,
			Required:    false,
			MultiValued: false,
			Model: models.Model{
				CreatedAt: time.Date(2021, 1, 1, 2, 3, 4, 0, time.UTC),
				UpdatedAt: time.Date(2021, 1, 2, 3, 3, 3, 0, time.UTC),
			},
		},
	}

	// Create test segmenters
	for i, customSegmenter := range customSegmenters {
		err := db.Create(&customSegmenter).Error
		if err != nil {
			return settings, []models.CustomSegmenter{}, err
		}
		// Return these custom segmenters with the correct segmenter types as expected test responses
		err = customSegmenter.FromStorageSchema(testSegmenterTypesMap)
		if err != nil {
			return settings, []models.CustomSegmenter{}, err
		}
		customSegmenters[i] = customSegmenter
	}

	return settings, customSegmenters, nil
}

func NewCountrySegmenter(_ json.RawMessage) (segmenters.Segmenter, error) {
	segmenterName := "country"
	countryConfig := &_segmenters.SegmenterConfiguration{
		Name: segmenterName,
		Type: _segmenters.SegmenterValueType_STRING,
		Options: map[string]*_segmenters.SegmenterValue{
			"indonesia": {Value: &_segmenters.SegmenterValue_String_{String_: "ID"}},
			"singapore": {Value: &_segmenters.SegmenterValue_String_{String_: "SG"}},
		},
		TreatmentRequestFields: &_segmenters.ListExperimentVariables{
			Values: []*_segmenters.ExperimentVariables{
				{
					Value: []string{segmenterName},
				},
				{
					Value: []string{"area"},
				},
			},
		},
		MultiValued: false,
		Required:    false,
	}

	return segmenters.NewBaseSegmenter(countryConfig), nil
}

func NewAreaSegmenter(configData json.RawMessage) (segmenters.Segmenter, error) {
	segmenterName := "area"
	areaConfig := _segmenters.SegmenterConfiguration{
		Name:        segmenterName,
		Type:        _segmenters.SegmenterValueType_INTEGER,
		MultiValued: true,
		TreatmentRequestFields: &_segmenters.ListExperimentVariables{
			Values: []*_segmenters.ExperimentVariables{
				{
					Value: []string{segmenterName},
				},
				{
					Value: []string{"latitude", "longitude"},
				},
			},
		},
		Required: false,
		Constraints: []*_segmenters.Constraint{
			{
				AllowedValues: &_segmenters.ListSegmenterValue{
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 3}},
					},
				},
				Options: map[string]*_segmenters.SegmenterValue{
					"area3": {Value: &_segmenters.SegmenterValue_Integer{Integer: 3}},
				},
				PreRequisites: []*_segmenters.PreRequisite{
					{
						SegmenterName: "country",
						SegmenterValues: &_segmenters.ListSegmenterValue{
							Values: []*_segmenters.SegmenterValue{
								{Value: &_segmenters.SegmenterValue_String_{String_: "SG"}},
							},
						},
					},
				},
			},
			{
				AllowedValues: &_segmenters.ListSegmenterValue{
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
					},
				},
				Options: map[string]*_segmenters.SegmenterValue{
					"area1": {Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
					"area2": {Value: &_segmenters.SegmenterValue_Integer{Integer: 2}},
				},
				PreRequisites: []*_segmenters.PreRequisite{
					{
						SegmenterName: "country",
						SegmenterValues: &_segmenters.ListSegmenterValue{
							Values: []*_segmenters.SegmenterValue{
								{Value: &_segmenters.SegmenterValue_String_{String_: "ID"}},
							},
						},
					},
				},
			},
		},
	}

	return segmenters.NewBaseSegmenter(&areaConfig), nil
}

func NewBoolSegmenter(_ json.RawMessage) (segmenters.Segmenter, error) {
	segmenterName := "bool_segmenter"
	boolConfig := &_segmenters.SegmenterConfiguration{
		Name: segmenterName,
		Type: _segmenters.SegmenterValueType_BOOL,
		Options: map[string]*_segmenters.SegmenterValue{
			"Yes": {Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
			"No":  {Value: &_segmenters.SegmenterValue_Bool{Bool: false}},
		},
		TreatmentRequestFields: &_segmenters.ListExperimentVariables{
			Values: []*_segmenters.ExperimentVariables{
				{
					Value: []string{segmenterName},
				},
			},
		},
		MultiValued: true,
		Required:    false,
	}

	return segmenters.NewBaseSegmenter(boolConfig), nil
}

func init() {
	err := segmenters.Register("country", NewCountrySegmenter)
	if err != nil {
		log.Fatal(err)
	}
	err = segmenters.Register("area", NewAreaSegmenter)
	if err != nil {
		log.Fatal(err)
	}
	err = segmenters.Register("bool_segmenter", NewBoolSegmenter)
	if err != nil {
		log.Fatal(err)
	}
}
