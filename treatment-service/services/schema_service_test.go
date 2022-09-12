package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/util"
)

type SchemaServiceTestSuite struct {
	suite.Suite

	schemaService SchemaService
}

type MockProjectSettingsStore struct {
	ProjectSettings map[models.ProjectId]*_pubsub.ProjectSettings
}

func (i *MockProjectSettingsStore) FindProjectSettingsWithId(projectId models.ProjectId) *_pubsub.ProjectSettings {
	return i.ProjectSettings[projectId]
}

func newMockProjectSettingsStore() models.ProjectSettingsStorage {
	return &MockProjectSettingsStore{
		ProjectSettings: map[models.ProjectId]*_pubsub.ProjectSettings{
			models.NewProjectId(1): {
				ProjectId:            1,
				Username:             "User1",
				EnableS2IdClustering: false,
				Segmenters: &_pubsub.Segmenters{
					Names: []string{
						"s2_ids", "days_of_week", "hours_of_day",
					},
					Variables: map[string]*_pubsub.ExperimentVariables{
						"s2_ids":       {Value: []string{"latitude", "longitude"}},
						"days_of_week": {Value: []string{"tz"}},
						"hours_of_day": {Value: []string{"tz"}},
					},
				},
				RandomizationKey: "order-id",
			},
			models.NewProjectId(2): {
				ProjectId:            2,
				Username:             "User2",
				EnableS2IdClustering: false,
				Segmenters: &_pubsub.Segmenters{
					Names: []string{
						"days_of_week", "hours_of_day",
					},
					Variables: map[string]*_pubsub.ExperimentVariables{
						"days_of_week": {Value: []string{"tz"}},
						"hours_of_day": {Value: []string{"tz"}},
					},
				},
				RandomizationKey: "merchant-id",
			},
			models.NewProjectId(3): {
				ProjectId:            3,
				Username:             "User3",
				EnableS2IdClustering: false,
				Segmenters: &_pubsub.Segmenters{
					Names: []string{"days_of_week"},
					Variables: map[string]*_pubsub.ExperimentVariables{
						"days_of_week": {Value: []string{"tz"}},
					},
				},
				RandomizationKey: "driver-id",
			},
			models.NewProjectId(4): {
				ProjectId:            4,
				Username:             "User4",
				EnableS2IdClustering: false,
				Segmenters: &_pubsub.Segmenters{
					Names: []string{"days_of_week"},
					Variables: map[string]*_pubsub.ExperimentVariables{
						"days_of_week": {Value: []string{"tz"}},
					},
				},
				RandomizationKey: "driver-id",
			},
		},
	}
}

func (suite *SchemaServiceTestSuite) SetupSuite() {
	cfg := map[string]interface{}{"s2_ids": map[string]interface{}{"mins2celllevel": 10, "maxs2celllevel": 14}}
	localStorage := models.LocalStorage{}
	segmenterSvc, err := NewSegmenterService(&localStorage, cfg)
	if err != nil {
		suite.FailNow("error creating segmenter service")
	}
	schemaService, err := NewSchemaService(newMockProjectSettingsStore(), segmenterSvc)
	if err != nil {
		suite.FailNow("error creating schema service")
	}
	suite.schemaService = schemaService
}

func TestSchemaServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SchemaServiceTestSuite))
}

func (suite *SchemaServiceTestSuite) TestGetRandomizationKeyValue() {
	filterParams := map[string]interface{}{
		"order-id": "1234",
	}
	expected := "1234"
	actual, err := suite.schemaService.GetRandomizationKeyValue(1, filterParams)
	suite.Require().Nil(err)
	suite.Require().Equal(&expected, actual)

	filterParams = map[string]interface{}{
		"merchant-id": "merchant-1234",
	}
	expected = "merchant-1234"
	actual, err = suite.schemaService.GetRandomizationKeyValue(2, filterParams)
	suite.Require().Nil(err)
	suite.Require().Equal(&expected, actual)

	filterParams = map[string]interface{}{}
	_, err = suite.schemaService.GetRandomizationKeyValue(3, filterParams)
	suite.Require().Nil(err)
}

func (suite *SchemaServiceTestSuite) TestGetRequestFilter() {
	timezone := "Asia/Singapore"
	longitude := 103.8998991137485
	latitude := 1.2537040223936706
	s2idL10, _ := util.GetS2ID(latitude, longitude, 14)
	s2IdSegmenterValues := []*_segmenters.SegmenterValue{}
	for i := 14; i >= 10; i-- {
		s2IdAtLevel := int64(s2idL10.Parent(i))
		segmenterValue := &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: s2IdAtLevel}}
		s2IdSegmenterValues = append(s2IdSegmenterValues, segmenterValue)
	}

	loc, _ := time.LoadLocation(timezone)
	dayOfWeek := util.RetrieveDayOfWeek(*loc)
	hourOfDay := util.RetrieveHourOfDay(*loc)
	filterParams := map[string]interface{}{
		"longitude": longitude,
		"latitude":  latitude,
		"order-id":  "1234",
		"tz":        timezone,
	}
	expected := map[string][]*_segmenters.SegmenterValue{
		"days_of_week": {{Value: &_segmenters.SegmenterValue_Integer{Integer: dayOfWeek}}},
		"hours_of_day": {{Value: &_segmenters.SegmenterValue_Integer{Integer: hourOfDay}}},
		"s2_ids":       s2IdSegmenterValues,
	}

	actual, err := suite.schemaService.GetRequestFilter(1, filterParams)
	suite.Require().Nil(err)
	suite.Require().Equal(expected, actual)

	filterParams = map[string]interface{}{}
	_, err = suite.schemaService.GetRequestFilter(5, filterParams)
	suite.Require().Errorf(err, "unable to find project id 5")
}
