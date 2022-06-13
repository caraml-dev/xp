package controller

import (
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	_segmenters "github.com/gojek/turing-experiments/common/segmenters"
	"github.com/gojek/turing-experiments/management-service/appcontext"
	"github.com/gojek/turing-experiments/management-service/errors"
	"github.com/gojek/turing-experiments/management-service/models"
	"github.com/gojek/turing-experiments/management-service/services"
	"github.com/gojek/turing-experiments/management-service/services/mocks"
)

type SegmenterControllerTestSuite struct {
	suite.Suite
	ctrl                        *SegmenterController
	expectedErrorResponseFormat string
	expectedSegmentersResponse  string
}

func (s *SegmenterControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up SegmenterControllerTestSuite")

	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`
	s.expectedSegmentersResponse = `{
		"data": [{
			"constraints": null,
			"description": "Test Desc",
			"multi_valued": false,
			"name": "test-segmenter",
			"options": {},
			"treatment_request_fields": [
					["test-segmenter"]
				],
			"type": "STRING",
			"required": false
		}]
	}`

	segmentersDescription := "Test Desc"
	segmenter := []*_segmenters.SegmenterConfiguration{
		{
			Constraints: nil,
			MultiValued: false,
			Description: segmentersDescription,
			Name:        "test-segmenter",
			Options:     make(map[string]*_segmenters.SegmenterValue),
			TreatmentRequestFields: &_segmenters.ListExperimentVariables{
				Values: []*_segmenters.ExperimentVariables{
					{
						Value: []string{"test-segmenter"},
					},
				},
			},
			Type: _segmenters.SegmenterValueType_STRING,
		},
	}
	projectSettings := models.Settings{
		ProjectID: 2,
		Config: &models.ExperimentationConfig{
			Segmenters: models.ProjectSegmenters{
				Names: []string{"test-segmenter"},
				Variables: map[string][]string{
					"test-segmenter": {"exp-var-test"},
				},
			},
			RandomizationKey: "rand",
		},
	}

	segmenterSvc := &mocks.SegmenterService{}
	segmenterSvc.
		On("ListSegmenterNames").
		Return([]string{"test-segmenter"})
	segmenterSvc.
		On("GetSegmenterConfigurations", []string{"test-segmenter"}).
		Return(segmenter, nil)

	settingsSvc := &mocks.ProjectSettingsService{}
	settingsSvc.
		On("GetDBRecord", models.ID(2)).
		Return(&projectSettings, nil)

	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", int64(1)).Return(nil, nil)
	mlpSvc.On("GetProject", int64(2)).Return(nil, nil)
	mlpSvc.On(
		"GetProject", int64(3),
	).Return(nil, errors.Newf(errors.NotFound, "MLP Project info for id %d not found in the cache", int64(3)))

	// Create test controller
	s.ctrl = &SegmenterController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				MLPService:             mlpSvc,
				SegmenterService:       segmenterSvc,
				ProjectSettingsService: settingsSvc,
			},
		},
	}
}

func TestSegmenterController(t *testing.T) {
	suite.Run(t, new(SegmenterControllerTestSuite))
}

func (s *SegmenterControllerTestSuite) TestListSegmenters() {
	t := s.Suite.T()

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "success",
			expected: s.expectedSegmentersResponse,
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.ListSegmenters(w, nil)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			s.Suite.Require().NoError(err)
			s.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}

func (s *SegmenterControllerTestSuite) TestGetSegmenters() {
	t := s.Suite.T()

	tests := []struct {
		name      string
		projectID int64
		expected  string
	}{
		{
			name:      "success",
			projectID: 2,
			expected:  s.expectedSegmentersResponse,
		},
		{
			name:      "mlp project not found",
			projectID: 3,
			expected:  fmt.Sprintf(s.expectedErrorResponseFormat, 404, "\"MLP Project info for id 3 not found in the cache\""),
		},
	}

	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Test error response
			s.ctrl.GetSegmenters(w, nil, data.projectID)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			s.Suite.Require().NoError(err)
			s.Suite.Assert().JSONEq(data.expected, string(body))
		})
	}
}
