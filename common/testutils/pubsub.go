package testutils

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartPubSubEmulator(ctx context.Context, project string) (testcontainers.Container, *pubsub.Client, error) {
	pubsubPort, _ := nat.NewPort("tcp", "8085")
	pubsubHostPort := fmt.Sprintf("0.0.0.0:%d", pubsubPort.Int())
	req := testcontainers.ContainerRequest{
		Image:        "google/cloud-sdk",
		ExposedPorts: []string{strconv.Itoa(pubsubPort.Int())},
		WaitingFor:   wait.ForLog("Server started"),
		Cmd:          []string{"gcloud", "beta", "emulators", "pubsub", "start", "--host-port", pubsubHostPort, "--project", project},
		// Reaper is unable to start up when using lima + nerdctl on MacOS. So, disabling it.
		SkipReaper: true,
	}
	emulator, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	var pubsubEmulatorAddress string
	if runtime.GOOS == "linux" {
		containerIp, err := emulator.ContainerIP(ctx)
		if err != nil {
			return nil, nil, err
		}
		pubsubEmulatorAddress = fmt.Sprintf("%s:%d", containerIp, pubsubPort.Int())
	} else {
		mappedPort, err := emulator.MappedPort(ctx, pubsubPort)
		if err != nil {
			return nil, nil, err
		}
		pubsubEmulatorAddress = fmt.Sprintf("%s:%d", "localhost", mappedPort.Int())
	}

	err = os.Setenv("PUBSUB_EMULATOR_HOST", pubsubEmulatorAddress)
	if err != nil {
		return nil, nil, err
	}
	err = os.Setenv("PUBSUB_PROJECT_ID", project)
	if err != nil {
		return nil, nil, err
	}

	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, nil, err
	}

	return emulator, client, nil
}

func CreatePubsubTopic(client *pubsub.Client, ctx context.Context, topicNames []string) error {
	for _, topicName := range topicNames {
		_, err := client.CreateTopic(ctx, string(topicName))
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateSubscriptions(client *pubsub.Client, ctx context.Context, topicNames []string) (map[string]*pubsub.Subscription, error) {
	subscriptions := make(map[string]*pubsub.Subscription)
	for _, topicName := range topicNames {
		subscriptionId := fmt.Sprintf("%s_subscription", topicName)
		topic := client.Topic(topicName)
		subscription, err := client.CreateSubscription(ctx,
			subscriptionId,
			pubsub.SubscriptionConfig{Topic: topic, ExpirationPolicy: time.Hour * 24, EnableMessageOrdering: true})
		if err != nil {
			return nil, err
		}
		subscriptions[topicName] = subscription
	}
	return subscriptions, nil
}
