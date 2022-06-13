package monitoring

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	_utils "github.com/gojek/turing-experiments/common/utils"
)

// kafkaProducer contains GetMetadata and Produce methods for mocking in unit tests
type kafkaProducer interface {
	GetMetadata(*string, bool, int) (*kafka.Metadata, error)
	Produce(*kafka.Message, chan kafka.Event) error
}

type KafkaLogPublisher struct {
	topic    string
	producer kafkaProducer
}

func (p *KafkaLogPublisher) Publish(logs []*AssignedTreatmentLog) error {
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	for _, l := range logs {
		keyBytes, valueBytes, err := newProtobufKafkaLogEntry(l)
		if err != nil {
			return err
		}

		err = p.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &p.topic,
				Partition: kafka.PartitionAny},
			Value: valueBytes,
			Key:   keyBytes,
		}, deliveryChan)
		if err != nil {
			return err
		}

		// Get delivery response
		event := <-deliveryChan
		msg := event.(*kafka.Message)
		if msg.TopicPartition.Error != nil {
			err = fmt.Errorf("Delivery failed: %v\n", msg.TopicPartition.Error)
			return err
		}
	}

	return nil
}

func NewKafkaLogPublisher(
	kafkaBrokers string,
	kafkaTopic string,
	kafkaMaxMessageBytes int,
	kafkaCompressionType string,
	KafkaConnectTimeoutMS int,
) (*KafkaLogPublisher, error) {
	// Create Kafka Producer
	producer, err := newKafkaProducer(kafkaBrokers, kafkaMaxMessageBytes, kafkaCompressionType)
	if err != nil {
		return nil, err
	}
	// Test that we are able to query the broker on the topic. If the topic
	// does not already exist on the broker, this should create it.
	_, err = producer.GetMetadata(&kafkaTopic, false, KafkaConnectTimeoutMS)
	if err != nil {
		return nil, fmt.Errorf("error Querying topic %s from Kafka broker(s): %s", kafkaTopic, err)
	}
	// Create Kafka Logger
	return &KafkaLogPublisher{
		topic:    kafkaTopic,
		producer: producer,
	}, nil
}

func newKafkaProducer(
	kafkaBrokers string,
	kafkaMaxMessageBytes int,
	kafkaCompressionType string,
) (kafkaProducer, error) {
	producer, err := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": kafkaBrokers,
			"message.max.bytes": kafkaMaxMessageBytes,
			"compression.type":  kafkaCompressionType,
		},
	)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

// newProtobufKafkaLogEntry converts a given AssignedTreatmentLog to the Protobuf format and marshals it,
// for writing to a Kafka topic
func newProtobufKafkaLogEntry(
	log *AssignedTreatmentLog,
) (keyBytes []byte, valueBytes []byte, err error) {
	timestamp := &timestamppb.Timestamp{Seconds: time.Now().Unix()}
	// Create the Kafka key
	key := &TreatmentServiceResultLogKey{
		RequestId:      log.RequestID,
		EventTimestamp: timestamp,
	}
	// Marshal the key
	keyBytes, err = proto.Marshal(key)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal log entry key, %s", err)
	}

	segments := make(map[string]interface{})
	for _, s := range log.Segmenters {
		allValues := []interface{}{}
		for _, v := range s.Value {
			allValues = append(allValues, _utils.SegmenterValueToInterface(v))
		}
		segments[s.Key] = allValues
	}

	segmentsJson, err := json.Marshal(segments)
	if err != nil {
		return nil, nil, err
	}
	message := &TreatmentServiceResultLogMessage{
		EventTimestamp: timestamp,
		ProjectId:      log.ProjectID,
		RequestId:      log.RequestID,
		Request:        log.Request,
		Segment:        string(segmentsJson),
	}

	if log.Experiment != nil {
		message.ExperimentId = log.Experiment.Id
		message.ExperimentName = log.Experiment.Name
	}

	if log.Treatment != nil {
		treatmentConfigJson, err := json.Marshal(log.Treatment.Config)
		if err != nil {
			return nil, nil, err
		}
		treatmentConfig := string(treatmentConfigJson)

		message.TreatmentName = log.Treatment.Name
		message.TreatmentConfig = treatmentConfig
	}

	if log.Error != nil {
		errorJson, err := json.Marshal(log.Error)
		if err != nil {
			return nil, nil, err
		}

		message.Error = string(errorJson)
	}

	// Marshal the message
	valueBytes, err = proto.Marshal(message)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal log entry value, %s", err)
	}

	return keyBytes, valueBytes, nil
}
