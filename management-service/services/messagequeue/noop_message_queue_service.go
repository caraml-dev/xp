package messagequeue

import (
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/common/segmenters"
)

// noopMQ is the struct for no operation to event updates
type noopMQ struct{}

// NewNoopMQ initializes a noopMQ struct
func NewNoopMQService() (MessageQueueService, error) {
	return &noopMQ{}, nil
}

func (k *noopMQ) PublishProjectSettingsMessage(updateType string, settings *_pubsub.ProjectSettings) error {
	return nil
}

func (k *noopMQ) PublishExperimentMessage(updateType string, experiment *_pubsub.Experiment) error {
	return nil
}

func (k *noopMQ) PublishProjectSegmenterMessage(updateType string, segmenter *segmenters.SegmenterConfiguration, projectId int64) error {
	return nil
}
