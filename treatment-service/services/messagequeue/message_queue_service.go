package messagequeue

import (
	"context"
	"fmt"

	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type MessageQueueService interface {
	SubscribeToManagementService(ctx context.Context) error
	DeleteSubscriptions(ctx context.Context) error
}

func NewMessageQueueService(
	ctx context.Context,
	storage *models.LocalStorage,
	mqConfig common_mq_config.MessageQueueConfig,
	projectIds []uint32,
	googleApplicationCredentialsEnvVar string,
) (MessageQueueService, error) {
	var mq MessageQueueService
	var err error
	switch mqConfig.Kind {
	case common_mq_config.NoopMQ:
		mq, err = NewNoopMQService()
	case common_mq_config.PubSubMQ:
		pubsubConfig := PubsubSubscriberConfig{
			Project:         mqConfig.PubSubConfig.Project,
			UpdateTopicName: mqConfig.PubSubConfig.TopicName,
			ProjectIds:      projectIds,
		}
		mq, err = NewPubsubMQService(ctx, storage, pubsubConfig, googleApplicationCredentialsEnvVar)
	default:
		return nil, fmt.Errorf("invalid message queue config (%s) was provided", mqConfig.Kind)
	}
	if err != nil {
		return nil, err
	}

	return mq, nil
}
