package service

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type MessageQueue interface {
	PublishNewExperiment(experiment schema.Experiment, segmentersType map[string]schema.SegmenterType) error
	UpdateExperiment(experiment schema.Experiment, segmentersType map[string]schema.SegmenterType) error
	UpdateProjectSettings(settings schema.ProjectSettings) error
}

type PubSubMessageQueue struct {
	context      context.Context
	updatesTopic *pubsub.Topic
	config       PubSubConfig
}

type PubSubConfig struct {
	GCPProject string
	TopicName  string
}

func NewPubSubMessageQueue(config PubSubConfig) (*PubSubMessageQueue, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GCPProject)
	if err != nil {
		return nil, err
	}
	return &PubSubMessageQueue{
		context:      ctx,
		updatesTopic: client.Topic(config.TopicName),
		config:       config,
	}, nil
}

func (i *PubSubMessageQueue) PublishNewExperiment(
	experiment schema.Experiment,
	segmentersType map[string]schema.SegmenterType,
) error {
	protoRecord, err := models.OpenAPIExperimentSpecToProtobuf(experiment, segmentersType)
	if err != nil {
		return err
	}

	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ExperimentCreated{
			ExperimentCreated: &_pubsub.ExperimentCreated{
				Experiment: protoRecord,
			},
		},
	}
	return i.publishMessage(&updateClientState)
}

func (i *PubSubMessageQueue) UpdateExperiment(
	experiment schema.Experiment,
	segmentersType map[string]schema.SegmenterType,
) error {
	protoRecord, err := models.OpenAPIExperimentSpecToProtobuf(experiment, segmentersType)
	if err != nil {
		return err
	}

	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ExperimentUpdated{
			ExperimentUpdated: &_pubsub.ExperimentUpdated{
				Experiment: protoRecord,
			},
		},
	}
	return i.publishMessage(&updateClientState)
}

func (i *PubSubMessageQueue) UpdateProjectSettings(settings schema.ProjectSettings) error {
	updateClientState := _pubsub.MessagePublishState{
		Update: &_pubsub.MessagePublishState_ProjectSettingsUpdated{
			ProjectSettingsUpdated: &_pubsub.ProjectSettingsUpdated{
				ProjectSettings: models.OpenAPIProjectSettingsSpecToProtobuf(settings),
			},
		},
	}
	return i.publishMessage(&updateClientState)
}

func (i *PubSubMessageQueue) publishMessage(messageProto proto.Message) error {
	payload, err := proto.Marshal(messageProto)
	if err != nil {
		return err
	}
	message := pubsub.Message{
		Data: payload,
	}
	_, err = i.updatesTopic.Publish(i.context, &message).Get(i.context)
	if err != nil {
		return err
	}
	return nil
}
