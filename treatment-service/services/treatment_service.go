package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	_pubsub "github.com/gojek/xp/common/pubsub"
	"github.com/gojek/xp/treatment-service/models"
	"github.com/gojek/xp/treatment-service/util"
)

type TreatmentService interface {
	// GetTreatment returns treatment based on provided experiment
	GetTreatment(experiment *_pubsub.Experiment, randomizationValue *string) (*_pubsub.ExperimentTreatment, error)
}

type treatmentService struct {
	localStorage *models.LocalStorage
}

func NewTreatmentService(localStorage *models.LocalStorage) (TreatmentService, error) {
	svc := &treatmentService{
		localStorage: localStorage,
	}

	return svc, nil
}

func (ts *treatmentService) GetTreatment(experiment *_pubsub.Experiment, randomizationValue *string) (*_pubsub.ExperimentTreatment, error) {
	if experiment == nil {
		// No experiments found
		return &_pubsub.ExperimentTreatment{}, nil
	}

	var treatment *_pubsub.ExperimentTreatment
	var err error
	if experiment.Type == _pubsub.Experiment_A_B {
		if randomizationValue == nil {
			return &_pubsub.ExperimentTreatment{}, errors.New("randomization key's value is nil")
		}
		treatment, err = getAbExperimentTreatment(experiment.Id, experiment.GetTreatments(), *randomizationValue)
	} else if experiment.Type == _pubsub.Experiment_Switchback {
		// TODO: Take into consideration when S2ID Clustering project settings option is switched on
		treatment, err = getSwitchbackExperimentTreatment(experiment.StartTime, experiment.Interval, experiment.Id, experiment.GetTreatments(), "")
	}
	if err != nil {
		return &_pubsub.ExperimentTreatment{}, err
	}

	return treatment, nil
}

func getSwitchbackExperimentTreatment(
	startTime *timestamppb.Timestamp,
	interval int32,
	experimentId int64,
	treatments []*_pubsub.ExperimentTreatment,
	randomizationValue string,
) (*_pubsub.ExperimentTreatment, error) {
	timeDifference := time.Since(startTime.AsTime()).Minutes()

	isCyclical := true
	if treatments[0].Traffic != 0 {
		isCyclical = false
	}

	var err error
	var treatmentIntervalIndex int
	// Cyclical Switchback Experiment; Traffic is not specified
	if isCyclical {
		treatmentIntervalIndex = int(math.Floor(
			math.Mod(timeDifference/float64(interval), float64(len(treatments))),
		))
		return treatments[treatmentIntervalIndex], nil
	}

	// Random Switchback Experiment; Traffic is specified
	treatmentIntervalIndex = int(math.Floor(timeDifference / float64(interval)))
	seed := getSwitchbackSeed(experimentId, randomizationValue, treatmentIntervalIndex)
	selectedTreatment, err := weightedChoice(treatments, seed)
	if err != nil {
		return &_pubsub.ExperimentTreatment{}, err
	}

	return selectedTreatment, nil
}

func getAbExperimentTreatment(
	experimentId int64,
	treatments []*_pubsub.ExperimentTreatment,
	randomizationValue string,
) (*_pubsub.ExperimentTreatment, error) {
	seed := getAbSeed(experimentId, randomizationValue)
	selectedTreatment, err := weightedChoice(treatments, seed)
	if err != nil {
		return &_pubsub.ExperimentTreatment{}, err
	}

	return selectedTreatment, nil
}

func weightedChoice(treatments []*_pubsub.ExperimentTreatment, seed string) (*_pubsub.ExperimentTreatment, error) {
	cumulativeTraffic := make([]uint32, len(treatments))
	total := uint32(0)
	for i, treatment := range treatments {
		total += treatment.Traffic
		cumulativeTraffic[i] = total
	}

	// Formulate Uniform distribution and get random number
	randomNum := getRandomNumber(seed, total)

	treatment, err := getWeightedChoiceTreatment(randomNum, cumulativeTraffic, treatments)
	if err != nil {
		return &_pubsub.ExperimentTreatment{}, err
	}

	return treatment, nil
}

func getWeightedChoiceTreatment(
	randomNum uint32,
	cumulativeTraffic []uint32,
	treatments []*_pubsub.ExperimentTreatment,
) (*_pubsub.ExperimentTreatment, error) {
	for i, threshold := range cumulativeTraffic {
		if randomNum < threshold {
			return treatments[i], nil
		}
	}

	return &_pubsub.ExperimentTreatment{}, errors.New("no suitable weighted choice found")
}

func getRandomNumber(seed string, maxNum uint32) uint32 {
	hashedSeed := util.Hash(seed)
	idx := hashedSeed % maxNum

	return idx
}

func getAbSeed(experimentID int64, randomizationUnit string) string {
	return fmt.Sprintf("%s-%d", randomizationUnit, experimentID)
}

func getSwitchbackSeed(experimentID int64, randomizationUnit string, treatmentIntervalIndex int) string {
	return fmt.Sprintf("%s-%d-%d", randomizationUnit, treatmentIntervalIndex, experimentID)
}
