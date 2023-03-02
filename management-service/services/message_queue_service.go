package services

import (
	"fmt"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/management-service/config"
)

type MessageQueueService interface {
	PublishProjectSettingsMessage(updateType string, settings *_pubsub.ProjectSettings) error
	PublishExperimentMessage(updateType string, experiment *_pubsub.Experiment) error
	PublishProjectSegmenterMessage(updateType string, segmenter *segmenters.SegmenterConfiguration, projectId int64) error
}

func NewMessageQueueService(mqConfig *config.MessageQueueConfig) (MessageQueueService, error) {
	var mq MessageQueueService
	var err error
	switch mqConfig.Kind {
	case config.NoopMQ:
		mq, err = NewNoopMQ()
	case config.PubSubMQ:
		mq, err = NewPubSubPublisherService(mqConfig.PubSubConfig)
	default:
		return nil, fmt.Errorf("invalid message queue kind (%s) was provided", mqConfig.Kind)
	}
	if err != nil {
		return nil, err
	}

	return mq, nil
}

// NoopMQ is the struct for no operation to event updates
type NoopMQ struct{}

// NewNoopMQ initializes a NoopMQ struct
func NewNoopMQ() (*NoopMQ, error) {
	return &NoopMQ{}, nil
}

func (k *NoopMQ) PublishProjectSettingsMessage(updateType string, settings *_pubsub.ProjectSettings) error {
	return nil
}

func (k *NoopMQ) PublishExperimentMessage(updateType string, experiment *_pubsub.Experiment) error {
	return nil
}

func (k *NoopMQ) PublishProjectSegmenterMessage(updateType string, segmenter *segmenters.SegmenterConfiguration, projectId int64) error {
	return nil
}
