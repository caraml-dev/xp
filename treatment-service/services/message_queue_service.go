package services

import (
	"context"
	"fmt"

	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type MessageQueueService interface {
	SubscribeToManagementService(ctx context.Context) error
	DeleteSubscriptions(ctx context.Context) error
}

func NewMessageQueueService(
	ctx context.Context,
	storage *models.LocalStorage,
	mqConfig config.MessageQueueConfig,
	projectIds []uint32,
	googleApplicationCredentialsEnvVar string,
) (MessageQueueService, error) {
	var mq MessageQueueService
	var err error
	switch mqConfig.Kind {
	case config.NoopMQ:
		mq, err = NewNoopMQ()
	case config.PubSubMQ:
		pubsubConfig := PubsubSubscriberConfig{
			Project:         mqConfig.PubSubConfig.Project,
			UpdateTopicName: mqConfig.PubSubConfig.TopicName,
			ProjectIds:      projectIds,
		}
		mq, err = NewPubsubSubscriber(ctx, storage, pubsubConfig, googleApplicationCredentialsEnvVar)
	default:
		return nil, fmt.Errorf("invalid message queue config (%s) was provided", mqConfig.Kind)
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

func (k *NoopMQ) SubscribeToManagementService(ctx context.Context) error {
	return nil
}

func (k *NoopMQ) DeleteSubscriptions(ctx context.Context) error {
	return nil
}
