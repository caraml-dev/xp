package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gojek/xp/common/api/schema"
	_pubsub "github.com/gojek/xp/common/pubsub"
	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/treatment-service/models"
)

type TreatmentSelectionSuite struct {
	suite.Suite

	treatmentService TreatmentService
	dayStart         time.Time
	hourStart        time.Time
	hourEnd          time.Time
}

func newTestXPExperiment(
	projectId int64,
	experimentType _pubsub.Experiment_Type,
	treatments []*_pubsub.ExperimentTreatment,
	startTime time.Time,
	endTime time.Time,
) _pubsub.Experiment {

	name, _ := uuid.NewUUID()

	return _pubsub.Experiment{
		Id:         0,
		ProjectId:  projectId,
		Status:     _pubsub.Experiment_Active,
		Name:       name.String(),
		Segments:   make(map[string]*_segmenters.ListSegmenterValue),
		Type:       experimentType,
		Interval:   30,
		StartTime:  timestamppb.New(startTime),
		EndTime:    timestamppb.New(endTime),
		Treatments: treatments,
		UpdatedAt:  timestamppb.New(time.Time{}),
	}
}

func (suite *TreatmentSelectionSuite) SetupSuite() {
	localStorage := models.LocalStorage{}
	treatmentService, _ := NewTreatmentService(&localStorage)
	suite.treatmentService = treatmentService

	dayStart := time.Now().Truncate(24 * time.Hour)
	hourStart := time.Now().Truncate(time.Hour)
	hourEnd := time.Now().Truncate(time.Hour).Add(time.Hour)
	suite.dayStart = dayStart
	suite.hourStart = hourStart
	suite.hourEnd = hourEnd
}

func TestTreatmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TreatmentSelectionSuite))
}

func (suite *TreatmentSelectionSuite) TestNoExperiments() {
	resp, err := suite.treatmentService.GetTreatment(nil, nil)

	suite.Require().NoError(err)
	suite.Require().Equal(&_pubsub.ExperimentTreatment{}, resp)
}

func (suite *TreatmentSelectionSuite) TestNoRandomizationValue() {
	treatment := []*_pubsub.ExperimentTreatment{{
		Name:    "ab-exp1-treatment1",
		Traffic: 100,
		Config:  &structpb.Struct{},
	},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_A_B, treatment, suite.dayStart, suite.hourStart)
	_, err := suite.treatmentService.GetTreatment(&experiment, nil)

	suite.Require().Error(err, "randomization key's value is nil")
}

func (suite *TreatmentSelectionSuite) TestTreatmentConversion() {
	rawCfg := map[string]interface{}{
		"a": "b",
		"c": map[string]interface{}{
			"d": 2.5,
			"e": []interface{}{true, false},
		},
	}
	config, err := structpb.NewStruct(rawCfg)
	suite.Require().NoError(err)
	treatment := []*_pubsub.ExperimentTreatment{
		{
			Name:    "ab-exp1-treatment1",
			Traffic: 100,
			Config:  config,
		},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_A_B, treatment, suite.dayStart, suite.hourStart)
	randomizationValue := ""
	resp, _ := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)

	var experimentTreatment *models.ExperimentTreatment
	var traffic int32
	experimentTreatment = &models.ExperimentTreatment{
		Configuration: models.DecodeTreatmentConfig(resp.GetConfig()),
		Name:          resp.GetName(),
		Traffic:       &traffic,
	}

	convertedResp := schema.SelectedTreatmentData{
		Name:          experimentTreatment.Name,
		Traffic:       experimentTreatment.Traffic,
		Configuration: experimentTreatment.Configuration,
	}
	traffic = int32(100)
	expectedTreatment := schema.SelectedTreatmentData{
		Configuration: rawCfg,
		Name:          "ab-exp1-treatment1",
		Traffic:       &traffic,
	}

	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment, convertedResp)
}

func (suite *TreatmentSelectionSuite) TestSingleAbExperiment() {
	treatment := []*_pubsub.ExperimentTreatment{
		{
			Name:    "ab-exp1-treatment1",
			Traffic: 100,
			Config:  &structpb.Struct{},
		},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_A_B, treatment, suite.dayStart, suite.hourStart)
	randomizationValue := ""
	resp, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)

	expectedTreatment := &_pubsub.ExperimentTreatment{
		Config:  &structpb.Struct{},
		Name:    "ab-exp1-treatment1",
		Traffic: uint32(100),
	}

	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment, resp)
}

func (suite *TreatmentSelectionSuite) TestMultipleAbExperiment() {
	treatment := []*_pubsub.ExperimentTreatment{
		{
			Name:    "ab-exp2-treatment1",
			Traffic: 30,
			Config:  &structpb.Struct{},
		},
		{
			Name:    "ab-exp2-treatment2",
			Traffic: 70,
			Config:  &structpb.Struct{},
		},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_A_B, treatment, suite.dayStart, suite.hourStart)
	// Should return different treatment based on randomization value
	randomizationValue := "1234567891"
	resp1, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)

	expectedTreatment1 := &_pubsub.ExperimentTreatment{
		Config:  &structpb.Struct{},
		Name:    "ab-exp2-treatment1",
		Traffic: 30,
	}
	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment1, resp1)

	randomizationValue = "12341"
	resp2, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)
	expectedTreatment2 := &_pubsub.ExperimentTreatment{
		Config:  &structpb.Struct{},
		Name:    "ab-exp2-treatment2",
		Traffic: 70,
	}

	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment2, resp2)
}

func (suite *TreatmentSelectionSuite) TestSingleCyclicalSbExperiment() {
	treatment := []*_pubsub.ExperimentTreatment{
		{
			Config: &structpb.Struct{},
			Name:   "sb-exp1-treatment1",
		},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_Switchback, treatment, suite.hourStart, suite.hourEnd)
	randomizationValue := "1234"
	resp, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)

	expectedTreatment := &_pubsub.ExperimentTreatment{
		Config: &structpb.Struct{},
		Name:   "sb-exp1-treatment1",
	}

	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment, resp)
}

func (suite *TreatmentSelectionSuite) TestMultiCyclicalSbExperiment() {
	treatment := []*_pubsub.ExperimentTreatment{
		{
			Config: &structpb.Struct{},
			Name:   "sb-exp2-treatment1",
		},
		{
			Config: &structpb.Struct{},
			Name:   "sb-exp2-treatment2",
		},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_Switchback, treatment, suite.hourStart, suite.hourEnd)

	randomizationValue := "1234"
	resp, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)

	expectedTreatment1 := &_pubsub.ExperimentTreatment{
		Config: &structpb.Struct{},
		Name:   "sb-exp2-treatment1",
	}
	expectedTreatment2 := &_pubsub.ExperimentTreatment{
		Config: &structpb.Struct{},
		Name:   "sb-exp2-treatment2",
	}

	suite.Require().NoError(err)
	// Different treatments based on 30min interval
	if time.Now().Minute() > 30 {
		suite.Require().Equal(expectedTreatment2, resp)
	} else {
		suite.Require().Equal(expectedTreatment1, resp)
	}
}

func (suite *TreatmentSelectionSuite) TestSingleRandomSbExperiment() {
	treatment := []*_pubsub.ExperimentTreatment{
		{
			Config:  &structpb.Struct{},
			Name:    "sb-exp3-treatment1",
			Traffic: 100,
		},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_Switchback, treatment, suite.hourStart, suite.hourEnd)

	randomizationValue := "1234"
	resp, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)

	expectedTreatment := &_pubsub.ExperimentTreatment{
		Config:  &structpb.Struct{},
		Name:    "sb-exp3-treatment1",
		Traffic: 100,
	}

	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment, resp)
}

func (suite *TreatmentSelectionSuite) TestMultiRandomSbExperiment() {
	treatment := []*_pubsub.ExperimentTreatment{
		{
			Config:  &structpb.Struct{},
			Name:    "sb-exp4-treatment1",
			Traffic: 80,
		},
		{
			Config:  &structpb.Struct{},
			Name:    "sb-exp4-treatment2",
			Traffic: 20,
		},
	}
	experiment := newTestXPExperiment(1, _pubsub.Experiment_Switchback, treatment, suite.hourStart, suite.hourEnd)

	expectedTreatment1 := &_pubsub.ExperimentTreatment{
		Config:  &structpb.Struct{},
		Name:    "sb-exp4-treatment1",
		Traffic: 80,
	}

	randomizationValue := "1234"
	resp1, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment1, resp1)

	randomizationValue = "12341"
	resp2, err := suite.treatmentService.GetTreatment(&experiment, &randomizationValue)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedTreatment1, resp2)
}

func (suite *TreatmentSelectionSuite) TestWeightedChoice() {

	// 20-80 split
	traffic10 := uint32(10)
	traffic20 := uint32(20)
	traffic30 := uint32(30)
	traffic60 := uint32(60)
	traffic80 := uint32(80)
	cumulativeTraffic := []uint32{20, 100}
	treatments := []*_pubsub.ExperimentTreatment{
		{
			Config:  nil,
			Name:    "exp1-treatment20",
			Traffic: traffic20,
		},
		{
			Config:  nil,
			Name:    "exp1-treatment80",
			Traffic: traffic80,
		},
	}
	randomNum := uint32(18)
	expectedTreatment := &_pubsub.ExperimentTreatment{
		Config:  nil,
		Name:    "exp1-treatment20",
		Traffic: 20,
	}
	actualTreatment, err := getWeightedChoiceTreatment(randomNum, cumulativeTraffic, treatments)
	suite.Require().Nil(err)
	suite.Require().Equal(expectedTreatment, actualTreatment)

	randomNum = uint32(58)
	expectedTreatment = &_pubsub.ExperimentTreatment{
		Config:  nil,
		Name:    "exp1-treatment80",
		Traffic: 80,
	}
	actualTreatment, err = getWeightedChoiceTreatment(randomNum, cumulativeTraffic, treatments)
	suite.Require().Nil(err)
	suite.Require().Equal(expectedTreatment, actualTreatment)

	// 60-10-30 split
	cumulativeTraffic = []uint32{60, 70, 100}
	treatments = []*_pubsub.ExperimentTreatment{
		{
			Config:  nil,
			Name:    "exp2-treatment60",
			Traffic: traffic60,
		},
		{
			Config:  nil,
			Name:    "exp2-treatment10",
			Traffic: traffic10,
		},
		{
			Config:  nil,
			Name:    "exp2-treatment30",
			Traffic: traffic30,
		},
	}

	randomNum = uint32(58)
	expectedTreatment = &_pubsub.ExperimentTreatment{
		Config:  nil,
		Name:    "exp2-treatment60",
		Traffic: 60,
	}
	actualTreatment, err = getWeightedChoiceTreatment(randomNum, cumulativeTraffic, treatments)
	suite.Require().Nil(err)
	suite.Require().Equal(expectedTreatment, actualTreatment)

	randomNum = uint32(68)
	expectedTreatment = &_pubsub.ExperimentTreatment{
		Config:  nil,
		Name:    "exp2-treatment10",
		Traffic: 10,
	}
	actualTreatment, err = getWeightedChoiceTreatment(randomNum, cumulativeTraffic, treatments)
	suite.Require().Nil(err)
	suite.Require().Equal(expectedTreatment, actualTreatment)

	randomNum = uint32(99)
	expectedTreatment = &_pubsub.ExperimentTreatment{
		Config:  nil,
		Name:    "exp2-treatment30",
		Traffic: 30,
	}
	actualTreatment, err = getWeightedChoiceTreatment(randomNum, cumulativeTraffic, treatments)
	suite.Require().Nil(err)
	suite.Require().Equal(expectedTreatment, actualTreatment)
}
