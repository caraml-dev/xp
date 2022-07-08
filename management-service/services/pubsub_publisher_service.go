package services

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"

	_pubsub "github.com/gojek/xp/common/pubsub"
	"github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/management-service/config"
)

type PubSubPublisherService interface {
	PublishProjectSettingsMessage(updateType string, settings *_pubsub.ProjectSettings) error
	PublishExperimentMessage(updateType string, experiment *_pubsub.Experiment) error
	PublishProjectSegmenterMessage(updateType string, segmenter *segmenters.SegmenterConfiguration, projectId int64) error
}

type pubSubPublisherService struct {
	context context.Context
	config  config.PubSubConfig
	topic   *pubsub.Topic
}

func NewPubSubPublisherService(config *config.PubSubConfig) (PubSubPublisherService, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.Project)
	if err != nil {
		return nil, err
	}

	topicIsPresent := false
	topicIterator := client.Topics(ctx)
	for {
		topic, err := topicIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if topic.ID() == config.TopicName {
			topicIsPresent = true
		}
	}
	if !topicIsPresent {
		_, err = client.CreateTopic(ctx, config.TopicName)
		if err != nil {
			return nil, err
		}
	}

	topic := client.Topic(config.TopicName)
	pubSubPublisher := pubSubPublisherService{
		context: ctx,
		config:  *config,
		topic:   topic,
	}

	return &pubSubPublisher, nil
}

func serializeCreateExperiment(experiment *_pubsub.Experiment) ([]byte, error) {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ExperimentCreated{
			ExperimentCreated: &_pubsub.ExperimentCreated{
				Experiment: experiment,
			},
		},
	}
	return proto.Marshal(&updateClientState)
}

func serializeUpdateExperiment(experiment *_pubsub.Experiment) ([]byte, error) {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ExperimentUpdated{
			ExperimentUpdated: &_pubsub.ExperimentUpdated{
				Experiment: experiment,
			},
		},
	}
	return proto.Marshal(&updateClientState)
}

func serializeCreateSettings(settings *_pubsub.ProjectSettings) ([]byte, error) {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ProjectSettingsCreated{
			ProjectSettingsCreated: &_pubsub.ProjectSettingsCreated{
				ProjectSettings: settings,
			},
		},
	}
	return proto.Marshal(&updateClientState)
}

func serializeUpdateSettings(settings *_pubsub.ProjectSettings) ([]byte, error) {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ProjectSettingsUpdated{
			ProjectSettingsUpdated: &_pubsub.ProjectSettingsUpdated{
				ProjectSettings: settings,
			},
		},
	}
	return proto.Marshal(&updateClientState)
}

func (p *pubSubPublisherService) PublishProjectSettingsMessage(updateType string, settings *_pubsub.ProjectSettings) error {
	var payload []byte
	var err error

	switch updateType {
	case "create":
		payload, err = serializeCreateSettings(settings)
	case "update":
		payload, err = serializeUpdateSettings(settings)
	}

	if err != nil {
		return err
	}
	message := pubsub.Message{
		Data: payload,
	}

	_, err = p.topic.Publish(p.context, &message).Get(p.context)
	if err != nil {
		return err
	}

	return nil
}

func (p *pubSubPublisherService) PublishExperimentMessage(updateType string, experiment *_pubsub.Experiment) error {
	var payload []byte
	var err error

	switch updateType {
	case "create":
		payload, err = serializeCreateExperiment(experiment)
	case "update":
		payload, err = serializeUpdateExperiment(experiment)
	}

	if err != nil {
		return err
	}
	message := pubsub.Message{
		Data: payload,
	}

	_, err = p.topic.Publish(p.context, &message).Get(p.context)
	if err != nil {
		return err
	}

	return nil
}

func (p *pubSubPublisherService) PublishProjectSegmenterMessage(
	updateType string,
	segmenter *segmenters.SegmenterConfiguration,
	projectId int64,
) error {

	var payload []byte
	var err error

	switch updateType {
	case "create":
		payload, err = serializeCreateSegmenter(segmenter, projectId)
	case "update":
		payload, err = serializeUpdateSegmenter(segmenter, projectId)
	case "delete":
		payload, err = serializeDeleteSegmenter(segmenter, projectId)
	}

	if err != nil {
		return err
	}
	message := pubsub.Message{
		Data: payload,
	}

	_, err = p.topic.Publish(p.context, &message).Get(p.context)
	if err != nil {
		return err
	}

	return nil
}

func serializeCreateSegmenter(segmenter *segmenters.SegmenterConfiguration, projectId int64) ([]byte, error) {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ProjectSegmenterCreated{
			ProjectSegmenterCreated: &segmenters.ProjectSegmenterCreated{
				ProjectId:        projectId,
				ProjectSegmenter: segmenter,
			},
		},
	}
	return proto.Marshal(&updateClientState)
}

func serializeUpdateSegmenter(segmenter *segmenters.SegmenterConfiguration, projectId int64) ([]byte, error) {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ProjectSegmenterUpdated{
			ProjectSegmenterUpdated: &segmenters.ProjectSegmenterUpdated{
				ProjectId:        projectId,
				ProjectSegmenter: segmenter,
			},
		},
	}
	return proto.Marshal(&updateClientState)
}

func serializeDeleteSegmenter(segmenter *segmenters.SegmenterConfiguration, projectId int64) ([]byte, error) {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ProjectSegmenterDeleted{
			ProjectSegmenterDeleted: &segmenters.ProjectSegmenterDeleted{
				ProjectId:     projectId,
				SegmenterName: segmenter.Name,
			},
		},
	}
	return proto.Marshal(&updateClientState)
}
