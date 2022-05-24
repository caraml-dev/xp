package services_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
)

type SegmenterServiceTestSuite struct {
	suite.Suite
	services.SegmenterService
}

func (s *SegmenterServiceTestSuite) SetupSuite() {
	segmenterConfig := map[string]interface{}{
		"s2_ids": map[string]interface{}{
			"mins2celllevel": 14,
			"maxs2celllevel": 15,
		},
	}
	var err error
	s.SegmenterService, err = services.NewSegmenterService(segmenterConfig)
	if err != nil {
		s.T().Fatalf("failed to start segmenter service: %s", err)
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
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			got, err := s.SegmenterService.GetSegmenterConfigurations(data.segmenterNames)
			s.Suite.Require().NoError(err)
			s.Suite.Assert().True(true, reflect.DeepEqual(data.expectedFields, got))
		})
	}
}

func (s *SegmenterServiceTestSuite) TestGetFormattedSegmenters() {
	s2IdFloat := []interface{}{float64(3592210809859604480)}
	daysOfWeekFloat := []interface{}{float64(1), float64(2)}

	s2IdResp := []interface{}{int64(3592210809859604480)}
	daysOfWeekResp := []interface{}{int64(1), int64(2)}

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
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			got, err := s.SegmenterService.GetFormattedSegmenters(data.segmenters)
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
	s2IdFloat := []interface{}{float64(3592210809859604480)}
	daysOfWeekFloat := []interface{}{float64(1), float64(2)}

	tests := map[string]struct {
		userSegmenters []string
		expSegment     models.ExperimentSegmentRaw
		errString      string
	}{
		"success | absent segmenters": {
			userSegmenters: []string{"s2_ids", "days_of_week"},
			expSegment: models.ExperimentSegmentRaw{
				"s2_ids": s2IdFloat,
			},
		},
		"success": {
			userSegmenters: []string{"s2_ids", "days_of_week"},
			expSegment: models.ExperimentSegmentRaw{
				"s2_ids":       s2IdFloat,
				"days_of_week": daysOfWeekFloat,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidateExperimentSegment(data.userSegmenters, data.expSegment)
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

	testS2Id1 := []string{"3592210809859604480"}
	testS2Id2 := []string{"3592210814154571776"}
	testDaysOfWeek1 := []string{"1"}
	testDaysOfWeek2 := []string{"2"}
	testDaysOfWeek3 := []string{"3"}
	testDaysOfWeek4 := []string{"1", "4"}
	tests := map[string]struct {
		userSegmenters []string
		expSegment     models.ExperimentSegmentRaw
		allExps        []models.Experiment
		errString      string
	}{
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
			err := s.SegmenterService.ValidateSegmentOrthogonality(data.userSegmenters, data.expSegment, data.allExps)
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
		providedSegmenters []string
		errString          string
	}{
		"success": {
			providedSegmenters: []string{"s2_ids", "days_of_week"},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidateRequiredSegmenters(data.providedSegmenters)
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
				Names: []string{"s2_ids"},
				Variables: map[string][]string{
					"s2_ids": {"s2id"},
				},
			},
		},
		"success | with segmenter indirect variable mapping": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"s2_ids"},
				Variables: map[string][]string{
					"s2_ids": {"latitude", "longitude"},
				},
			},
		},
		"failure | missing dependent segmenter": {
			projectSegmenters: models.ProjectSegmenters{
				Names:     []string{"s2_ids"},
				Variables: map[string][]string{},
			},
			errString: "len of project segmenters does not match mapping of experiment variables",
		},
		"failure | invalid experiment variables mapping": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"s2_ids"},
				Variables: map[string][]string{
					"days_of_week": {"latitude", "longitude"},
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
			errString: "Unknown segmenter abc",
		},
		"failure | invalid variables": {
			projectSegmenters: models.ProjectSegmenters{
				Names: []string{"s2_ids"},
				Variables: map[string][]string{
					"s2_ids": {"latitude", "longitude", "unknown"},
				},
			},
			errString: "segmenter (s2_ids) does not have valid experiment variable(s) provided",
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.SegmenterService.ValidateExperimentVariables(data.projectSegmenters)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}
