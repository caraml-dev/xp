package services

import (
	"fmt"
	"testing"

	"github.com/gojek/xp/common/api/schema"
	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/treatment-service/models"
	"github.com/stretchr/testify/suite"
)

type SegmenterServiceTestSuite struct {
	suite.Suite
	SegmenterService
}

func (s *SegmenterServiceTestSuite) SetupSuite() {
	segmenterConfig := map[string]interface{}{
		"s2_ids": map[string]interface{}{
			"mins2celllevel": 10,
			"maxs2celllevel": 14,
		},
	}
	localStorage := models.LocalStorage{
		ProjectSegmenters: map[models.ProjectId]map[string]schema.SegmenterType{1: {}},
	}
	var err error
	s.SegmenterService, err = NewSegmenterService(&localStorage, segmenterConfig)
	if err != nil {
		s.T().Fatalf("failed to start segmenter service: %s", err)
	}
}

func TestSegmenterService(t *testing.T) {
	suite.Run(t, new(SegmenterServiceTestSuite))
}

func (s *SegmenterServiceTestSuite) TestGetTransformation() {
	errStringFormat := "experiment variable (%s) was not provided for segmenter (%s)"
	segmenterName := "days_of_week"
	requiredVariableName := "day_of_week"
	timezone := "tz"
	tests := map[string]struct {
		projectId            int
		segmenterName        string
		requiredVariableName string
		providedVariables    map[string]interface{}
		experimentVariables  []string
		expectedValue        []*_segmenters.SegmenterValue
		errString            string
	}{
		"failure | missing experiment variable": {
			projectId:            1,
			segmenterName:        segmenterName,
			requiredVariableName: requiredVariableName,
			providedVariables: map[string]interface{}{
				timezone: "Asia/Singapore",
			},
			experimentVariables: []string{requiredVariableName},
			errString:           fmt.Sprintf(errStringFormat, requiredVariableName, segmenterName),
		},
		"failure | invalid proj": {
			projectId: 99,
			errString: "project segmenter not found for project id: 99",
		},
		"failure | invalid segmenter": {
			projectId:     1,
			segmenterName: "non-existence-segmenter",
			errString:     "Type mapping not found for Segmenter:non-existence-segmenter",
		},
		"success": {
			projectId:            1,
			segmenterName:        segmenterName,
			requiredVariableName: requiredVariableName,
			providedVariables: map[string]interface{}{
				"day_of_week": float64(1),
			},
			expectedValue:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(1)}}},
			experimentVariables: []string{requiredVariableName},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			got, err := s.SegmenterService.GetTransformation(
				models.ProjectId(data.projectId),
				data.segmenterName,
				data.providedVariables,
				data.experimentVariables)

			if data.errString == "" {
				s.Suite.Require().NoError(err)
				s.Suite.Assert().Equal(data.expectedValue, got)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}
