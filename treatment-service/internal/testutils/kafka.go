package testutils

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/testcontainers/testcontainers-go"
)

const (
	CLUSTER_NETWORK_NAME = "kafka-cluster"
	ZOOKEEPER_PORT       = "2181"
	KAFKA_BROKER_PORT    = "9092"
	KAFKA_CLIENT_PORT    = "9093"
	ZOOKEEPER_IMAGE      = "confluentinc/cp-zookeeper:5.2.1"
	KAFKA_IMAGE          = "confluentinc/cp-kafka:5.2.1"
)

type KafkaCluster struct {
	kafkaContainer     testcontainers.Container
	zookeeperContainer testcontainers.Container
}

func (kc *KafkaCluster) StartCluster() error {
	ctx := context.Background()

	err := kc.zookeeperContainer.Start(ctx)
	if err != nil {
		return err
	}
	err = kc.kafkaContainer.Start(ctx)
	if err != nil {
		return err
	}
	err = kc.startKafka()
	if err != nil {
		return err
	}

	return nil
}

func (kc *KafkaCluster) GetKafkaHost() string {
	ctx := context.Background()
	host, err := kc.kafkaContainer.Host(ctx)
	if err != nil {
		panic(err)
	}
	port, err := kc.kafkaContainer.MappedPort(ctx, KAFKA_CLIENT_PORT)
	if err != nil {
		panic(err)
	}

	// returns the exposed kafka host:port
	return host + ":" + port.Port()
}

func (kc *KafkaCluster) startKafka() error {
	ctx := context.Background()

	kafkaStartFile, err := ioutil.TempFile("", "testcontainers_start.sh")
	if err != nil {
		return err
	}
	defer os.Remove(kafkaStartFile.Name())

	// needs to set KAFKA_ADVERTISED_LISTENERS with the exposed kafka port
	exposedHost := kc.GetKafkaHost()
	_, err = kafkaStartFile.WriteString("#!/bin/bash \n")
	if err != nil {
		return err
	}
	_, err = kafkaStartFile.WriteString("export KAFKA_ADVERTISED_LISTENERS='PLAINTEXT://" + exposedHost + ",BROKER://kafka:" + KAFKA_BROKER_PORT + "'\n")
	if err != nil {
		return err
	}
	_, err = kafkaStartFile.WriteString(". /etc/confluent/docker/bash-config \n")
	if err != nil {
		return err
	}
	_, err = kafkaStartFile.WriteString("/etc/confluent/docker/configure \n")
	if err != nil {
		return err
	}
	_, err = kafkaStartFile.WriteString("/etc/confluent/docker/launch \n")
	if err != nil {
		return err
	}

	err = kc.kafkaContainer.CopyFileToContainer(ctx, kafkaStartFile.Name(), "testcontainers_start.sh", 0700)
	if err != nil {
		return err
	}
	return nil
}

func NewKafkaCluster() *KafkaCluster {
	ctx := context.Background()

	// creates a network, so kafka and zookeeper can communicate directly
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{Name: CLUSTER_NETWORK_NAME},
	})
	if err != nil {
		panic(err)
	}

	dockerNetwork := network.(*testcontainers.DockerNetwork)

	zookeeperContainer := createZookeeperContainer(dockerNetwork)
	kafkaContainer := createKafkaContainer(dockerNetwork)

	return &KafkaCluster{
		zookeeperContainer: zookeeperContainer,
		kafkaContainer:     kafkaContainer,
	}
}

func createZookeeperContainer(network *testcontainers.DockerNetwork) testcontainers.Container {
	ctx := context.Background()

	// creates the zookeeper container, but do not start it yet
	zookeeperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          ZOOKEEPER_IMAGE,
			ExposedPorts:   []string{ZOOKEEPER_PORT},
			Env:            map[string]string{"ZOOKEEPER_CLIENT_PORT": ZOOKEEPER_PORT, "ZOOKEEPER_TICK_TIME": "2000"},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"zookeeper"}},
		},
	})
	if err != nil {
		panic(err)
	}

	return zookeeperContainer
}

func createKafkaContainer(network *testcontainers.DockerNetwork) testcontainers.Container {
	ctx := context.Background()

	// creates the kafka container, but do not start it yet
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        KAFKA_IMAGE,
			ExposedPorts: []string{KAFKA_CLIENT_PORT},
			Env: map[string]string{
				"KAFKA_BROKER_ID":                        "1",
				"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:" + ZOOKEEPER_PORT,
				"KAFKA_LISTENERS":                        "PLAINTEXT://0.0.0.0:" + KAFKA_CLIENT_PORT + ",BROKER://0.0.0.0:" + KAFKA_BROKER_PORT,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
				"KAFKA_INTER_BROKER_LISTENER_NAME":       "BROKER",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"kafka"}},
			// the container only starts when it finds and run /testcontainers_start.sh
			Cmd: []string{"sh", "-c", "while [ ! -f /testcontainers_start.sh ]; do sleep 0.1; done; /testcontainers_start.sh"},
		},
	})
	if err != nil {
		panic(err)
	}

	return kafkaContainer
}
