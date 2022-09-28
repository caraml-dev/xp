package monitoring

import (
	"log"
	"time"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

// ErrorResponseLog contains the status code and error string to be included in outcome logging
type ErrorResponseLog struct {
	Code  int
	Error string
}

type TreatmentMetadata struct {
	ExperimentType     string `json:"experiment_type"`
	ExperimentVersion  int64  `json:"experiment_version"`
	SwitchbackWindowId *int64 `json:"switchback_window_id"`
}

type AssignedTreatmentLog struct {
	ProjectID         models.ProjectId
	RequestID         string
	Experiment        *_pubsub.Experiment
	Treatment         *_pubsub.ExperimentTreatment
	TreatmentMetadata *TreatmentMetadata
	Request           *Request
	Segmenters        []models.SegmentFilter
	Error             *ErrorResponseLog
}

type AssignedTreatmentPublisher interface {
	Publish(log []*AssignedTreatmentLog) error
}

type AssignedTreatmentLogger struct {
	queue     chan *AssignedTreatmentLog
	publisher AssignedTreatmentPublisher

	flushInterval time.Duration
}

func (l *AssignedTreatmentLogger) Append(log *AssignedTreatmentLog) error {
	l.queue <- log
	return nil
}

func (l *AssignedTreatmentLogger) worker() {
	for range time.Tick(l.flushInterval) {
		logs := make([]*AssignedTreatmentLog, 0)

	collection:
		for {
			select {
			case log := <-l.queue:
				logs = append(logs, log)
			default:
				break collection
			}
		}

		if len(logs) > 0 {
			err := l.publisher.Publish(logs)
			if err != nil {
				log.Println("Failed to publish log:", err)
			}
		}
	}
}

func NewNoopAssignedTreatmentLogger() (*AssignedTreatmentLogger, error) {
	return nil, nil
}

func NewBQAssignedTreatmentLogger(
	config config.BigqueryConfig,
	queueLength int,
	flushInterval time.Duration,
) (*AssignedTreatmentLogger, error) {

	c := make(chan *AssignedTreatmentLog, queueLength)
	publisher, err := NewBQLogPublisher(config.Project, config.Dataset, config.Table)
	if err != nil {
		return nil, err
	}
	logger := &AssignedTreatmentLogger{
		queue:         c,
		publisher:     publisher,
		flushInterval: flushInterval,
	}

	go logger.worker()

	return logger, nil
}

func NewKafkaAssignedTreatmentLogger(
	config config.KafkaConfig,
	queueLength int,
	flushInterval time.Duration,
) (*AssignedTreatmentLogger, error) {

	c := make(chan *AssignedTreatmentLog, queueLength)
	publisher, err := NewKafkaLogPublisher(
		config.Brokers, config.Topic, config.MaxMessageBytes, config.CompressionType, config.ConnectTimeoutMS,
	)
	if err != nil {
		return nil, err
	}
	logger := &AssignedTreatmentLogger{
		queue:         c,
		publisher:     publisher,
		flushInterval: flushInterval,
	}

	go logger.worker()

	return logger, nil
}
