package appcontext

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/monitoring"
	"github.com/caraml-dev/xp/treatment-service/services"
)

type AppContext struct {
	ExperimentService services.ExperimentService
	MetricService     services.MetricService
	SchemaService     services.SchemaService
	TreatmentService  services.TreatmentService
	SegmenterService  services.SegmenterService

	AssignedTreatmentLogger *monitoring.AssignedTreatmentLogger
	ExperimentSubscriber    services.ExperimentSubscriber
}

func NewAppContext(cfg *config.Config) (*AppContext, error) {
	log.Println("Initializing local storage...")
	localStorage, err := models.NewLocalStorage(
		cfg.GetProjectIds(),
		cfg.ManagementService.URL,
		cfg.ManagementService.AuthorizationEnabled,
	)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing segmenter service...")
	segmenterSvc, err := services.NewSegmenterService(localStorage, cfg.SegmenterConfig)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing schema service...")
	schemaSvc, err := services.NewSchemaService(localStorage, segmenterSvc)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing experiment service...")
	experimentSvc, err := services.NewExperimentService(localStorage)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing treatment service...")
	treatmentSvc, err := services.NewTreatmentService(localStorage)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing metric service...")
	metricService, err := services.NewMetricService(cfg.MonitoringConfig, localStorage)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing assigned treatment logger...")
	loggerConfig := cfg.AssignedTreatmentLogger
	var logger *monitoring.AssignedTreatmentLogger

	switch loggerConfig.Kind {
	case config.KafkaLogger:
		logger, err = monitoring.NewKafkaAssignedTreatmentLogger(
			*loggerConfig.KafkaConfig,
			loggerConfig.QueueLength, time.Duration(loggerConfig.FlushIntervalSeconds)*time.Second)
	case config.BQLogger:
		logger, err = monitoring.NewBQAssignedTreatmentLogger(
			*loggerConfig.BQConfig,
			loggerConfig.QueueLength, time.Duration(loggerConfig.FlushIntervalSeconds)*time.Second)
	case config.NoopLogger:
		logger, err = monitoring.NewNoopAssignedTreatmentLogger()
	default:
		err = fmt.Errorf("unrecognized Treatment Logger Kind: %s", loggerConfig.Kind)
	}
	if err != nil {
		return nil, err
	}

	log.Println("Initializing pubsub subscriber...")
	pubsubConfig := services.PubsubSubscriberConfig{
		Project:         cfg.PubSub.Project,
		UpdateTopicName: cfg.PubSub.TopicName,
		ProjectIds:      cfg.GetProjectIds(),
	}
	pubsubInitContext, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.PubSub.PubSubTimeoutSeconds)*time.Second)
	defer cancel()
	experimentSubscriber, err := services.NewPubsubSubscriber(pubsubInitContext, localStorage, pubsubConfig)
	if err != nil {
		return nil, err
	}

	appContext := &AppContext{
		ExperimentService:       experimentSvc,
		MetricService:           metricService,
		SegmenterService:        segmenterSvc,
		SchemaService:           schemaSvc,
		TreatmentService:        treatmentSvc,
		AssignedTreatmentLogger: logger,
		ExperimentSubscriber:    experimentSubscriber,
	}

	return appContext, nil
}
