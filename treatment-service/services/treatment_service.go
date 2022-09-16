package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/util"
)

type TreatmentService interface {
	// GetTreatment returns treatment based on provided experiment. If the experiment's type is Switchback,
	// the window Id is also returned.
	GetTreatment(experiment *_pubsub.Experiment, randomizationValue *string) (*_pubsub.ExperimentTreatment, *int64, error)
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

func (ts *treatmentService) GetTreatment(experiment *_pubsub.Experiment, randomizationValue *string) (*_pubsub.ExperimentTreatment, *int64, error) {
	if experiment == nil {
		// No experiments found
		return &_pubsub.ExperimentTreatment{}, nil, nil
	}

	var switchbackWindowId *int64
	var treatment *_pubsub.ExperimentTreatment
	var err error
	if experiment.Type == _pubsub.Experiment_A_B {
		if randomizationValue == nil {
			return &_pubsub.ExperimentTreatment{}, nil, errors.New("randomization key's value is nil")
		}
		treatment, err = getAbExperimentTreatment(experiment.Id, experiment.GetTreatments(), *randomizationValue)
	} else if experiment.Type == _pubsub.Experiment_Switchback {
		// TODO: Take into consideration when S2ID Clustering project settings option is switched on
		var windowId int64
		treatment, windowId, err = getSwitchbackExperimentTreatment(
			experiment.StartTime,
			experiment.Interval,
			experiment.Id,
			experiment.GetTreatments(),
			"",
		)
		switchbackWindowId = &windowId
	}
	if err != nil {
		return &_pubsub.ExperimentTreatment{}, nil, err
	}

	return treatment, switchbackWindowId, nil
}

func getSwitchbackExperimentTreatment(
	startTime *timestamppb.Timestamp,
	interval int32,
	experimentId int64,
	treatments []*_pubsub.ExperimentTreatment,
	randomizationValue string,
) (*_pubsub.ExperimentTreatment, int64, error) {
	timeDifference := time.Since(startTime.AsTime()).Minutes()
	treatmentIntervalIndex := int64(math.Floor(timeDifference / float64(interval)))

	isCyclical := true
	if treatments[0].Traffic != 0 {
		isCyclical = false
	}

	var err error
	// Cyclical Switchback Experiment; Traffic is not specified
	if isCyclical {
		cyclicalIndex := int(math.Floor(
			math.Mod(timeDifference/float64(interval), float64(len(treatments))),
		))
		return treatments[cyclicalIndex], treatmentIntervalIndex, nil
	}

	// Random Switchback Experiment; Traffic is specified
	seed := getSwitchbackSeed(experimentId, randomizationValue, treatmentIntervalIndex)
	selectedTreatment, err := weightedChoice(treatments, seed)
	if err != nil {
		return &_pubsub.ExperimentTreatment{}, treatmentIntervalIndex, err
	}

	return selectedTreatment, treatmentIntervalIndex, nil
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

func getSwitchbackSeed(experimentID int64, randomizationUnit string, treatmentIntervalIndex int64) string {
	return fmt.Sprintf("%s-%d-%d", randomizationUnit, treatmentIntervalIndex, experimentID)
}
