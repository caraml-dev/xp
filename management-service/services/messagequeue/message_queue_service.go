package messagequeue

import (
	"fmt"

	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/common/segmenters"
)

type MessageQueueService interface {
	PublishProjectSettingsMessage(updateType string, settings *_pubsub.ProjectSettings) error
	PublishExperimentMessage(updateType string, experiment *_pubsub.Experiment) error
	PublishProjectSegmenterMessage(updateType string, segmenter *segmenters.SegmenterConfiguration, projectId int64) error
}

func NewMessageQueueService(mqConfig common_mq_config.MessageQueueConfig) (MessageQueueService, error) {
	var mq MessageQueueService
	var err error
	switch mqConfig.Kind {
	case common_mq_config.NoopMQ:
		mq, err = NewNoopMQService()
	case common_mq_config.PubSubMQ:
		mq, err = NewPubSubMQService(*mqConfig.PubSubConfig)
	default:
		return nil, fmt.Errorf("invalid message queue kind (%s) was provided", mqConfig.Kind)
	}
	if err != nil {
		return nil, err
	}

	return mq, nil
}
