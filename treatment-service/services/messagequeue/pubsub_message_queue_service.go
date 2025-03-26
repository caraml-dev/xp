package messagequeue

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type PubsubSubscriber struct {
	localStorage *models.LocalStorage
	subscription *pubsub.Subscription
	projectIds   []models.ProjectId
}

type PubsubSubscriberConfig struct {
	Project         string
	UpdateTopicName string
	ProjectIds      []models.ProjectId
}

func newSubscriptionId(topic string) string {
	return fmt.Sprintf("%s_sub_%s", topic, uuid.NewString())
}

func newPubsubSubscription(ctx context.Context, client *pubsub.Client, topic string) (*pubsub.Subscription, error) {
	return client.CreateSubscription(
		ctx, newSubscriptionId(topic), pubsub.SubscriptionConfig{
			Topic:                 client.Topic(topic),
			ExpirationPolicy:      time.Hour * 24,
			EnableMessageOrdering: true,
		},
	)
}

func NewPubsubMQService(
	ctx context.Context,
	storage *models.LocalStorage,
	config PubsubSubscriberConfig,
	googleApplicationCredentialsEnvVar string,
) (*PubsubSubscriber, error) {
	var client *pubsub.Client
	var err error
	if filepath := os.Getenv(googleApplicationCredentialsEnvVar); filepath != "" {
		client, err = pubsub.NewClient(ctx, config.Project, option.WithCredentialsFile(filepath))
	} else {
		client, err = pubsub.NewClient(ctx, config.Project)
	}

	if err != nil {
		return nil, err
	}

	subscription, err := newPubsubSubscription(ctx, client, config.UpdateTopicName)
	if err != nil {
		return nil, err
	}

	return &PubsubSubscriber{
		localStorage: storage,
		subscription: subscription,
		projectIds:   config.ProjectIds,
	}, nil
}

func (u *PubsubSubscriber) SubscribeToManagementService(ctx context.Context) error {
	return u.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		defer msg.Ack()
		update := _pubsub.MessagePublishState{}
		err := proto.Unmarshal(msg.Data, &update)
		if err != nil {
			log.Println("Warning: unable to unmarshal message for new experiment:", err)
			msg.Ack()
		}

		updateType := update.Update
		switch updateType.(type) {
		case *_pubsub.MessagePublishState_ExperimentCreated:
			experiment := update.GetExperimentCreated().Experiment
			if models.ContainsProjectId(u.projectIds, models.ProjectId(experiment.ProjectId)) {
				u.localStorage.InsertExperiment(experiment)
			}
		case *_pubsub.MessagePublishState_ExperimentUpdated:
			experiment := update.GetExperimentUpdated().Experiment
			if models.ContainsProjectId(u.projectIds, models.ProjectId(experiment.ProjectId)) {
				u.localStorage.UpdateExperiment(experiment)
			}
		case *_pubsub.MessagePublishState_ProjectSettingsCreated:
			if err := u.localStorage.InsertProjectSettings(update.GetProjectSettingsCreated().ProjectSettings); err != nil {
				log.Println("Warning: unable to insert segmenters for new project settings:", err)
				return
			}
		case *_pubsub.MessagePublishState_ProjectSettingsUpdated:
			u.localStorage.UpdateProjectSettings(update.GetProjectSettingsUpdated().ProjectSettings)
		case *_pubsub.MessagePublishState_ProjectSegmenterCreated:
			u.localStorage.UpdateProjectSegmenters(
				update.GetProjectSegmenterCreated().ProjectSegmenter,
				update.GetProjectSegmenterCreated().ProjectId)
		case *_pubsub.MessagePublishState_ProjectSegmenterUpdated:
			u.localStorage.UpdateProjectSegmenters(
				update.GetProjectSegmenterUpdated().ProjectSegmenter,
				update.GetProjectSegmenterUpdated().ProjectId)
		case *_pubsub.MessagePublishState_ProjectSegmenterDeleted:
			u.localStorage.DeleteProjectSegmenters(
				update.GetProjectSegmenterDeleted().SegmenterName,
				update.GetProjectSegmenterDeleted().ProjectId)
		}
	})
}

func (u *PubsubSubscriber) DeleteSubscriptions(ctx context.Context) error {
	if err := u.subscription.Delete(ctx); err != nil {
		return err
	}
	return nil
}
