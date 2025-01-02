package appcontext

import (
	"context"
	"fmt"
	"log"
	"time"

	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/monitoring"
	"github.com/caraml-dev/xp/treatment-service/services"
	"github.com/caraml-dev/xp/treatment-service/services/messagequeue"
)

type AppContext struct {
	ExperimentService   services.ExperimentService
	MessageQueueService messagequeue.MessageQueueService
	MetricService       services.MetricService
	SchemaService       services.SchemaService
	TreatmentService    services.TreatmentService
	SegmenterService    services.SegmenterService

	AssignedTreatmentLogger *monitoring.AssignedTreatmentLogger
	LocalStorage            *models.LocalStorage
}

func NewAppContext(cfg *config.Config) (*AppContext, error) {
	log.Println("Initializing local storage...")
	localStorage, err := models.NewLocalStorage(
		cfg.GetProjectIds(),
		cfg.ManagementService.URL,
		cfg.ManagementService.AuthorizationEnabled,
		cfg.DeploymentConfig.GoogleApplicationCredentialsEnvVar,
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
			loggerConfig.QueueLength,
			time.Duration(loggerConfig.FlushIntervalSeconds)*time.Second,
			cfg.DeploymentConfig.GoogleApplicationCredentialsEnvVar,
		)
	case config.NoopLogger:
		logger, err = monitoring.NewNoopAssignedTreatmentLogger()
	default:
		err = fmt.Errorf("unrecognized Treatment Logger Kind: %s", loggerConfig.Kind)
	}
	if err != nil {
		return nil, err
	}

	log.Println("Initializing message queue subscriber...")
	var messageQueueService messagequeue.MessageQueueService
	switch cfg.MessageQueueConfig.Kind {
	case common_mq_config.NoopMQ:
		messageQueueService, err = messagequeue.NewMessageQueueService(
			context.Background(),
			localStorage,
			cfg.MessageQueueConfig,
			cfg.GetProjectIds(),
			cfg.DeploymentConfig.GoogleApplicationCredentialsEnvVar,
		)
	case common_mq_config.PubSubMQ:
		pubsubInitContext, cancel := context.WithTimeout(
			context.Background(), time.Duration(cfg.MessageQueueConfig.PubSubConfig.PubSubTimeoutSeconds)*time.Second)
		defer cancel()
		messageQueueService, err = messagequeue.NewMessageQueueService(
			pubsubInitContext,
			localStorage,
			cfg.MessageQueueConfig,
			cfg.GetProjectIds(),
			cfg.DeploymentConfig.GoogleApplicationCredentialsEnvVar,
		)
	default:
		err = fmt.Errorf("unrecognized Message Queue Kind: %s", loggerConfig.Kind)
	}
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
		MessageQueueService:     messageQueueService,
		LocalStorage:            localStorage,
	}

	return appContext, nil
}
