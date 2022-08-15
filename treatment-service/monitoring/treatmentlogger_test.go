package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

func TestNoopAssignedTreatmentLogger(t *testing.T) {
	logger, err := NewNoopAssignedTreatmentLogger()
	var expected *AssignedTreatmentLogger

	assert.NoError(t, nil, err)
	assert.Equal(t, expected, logger)
}

func TestNewProtobufKafkaLogEntry(t *testing.T) {
	treatmentCfg, _ := structpb.NewStruct(map[string]interface{}{
		"treatment-key": "treatment-value",
	})
	assignedTreatmentLog := &AssignedTreatmentLog{
		ProjectID: 0,
		RequestID: "1",
		Experiment: &_pubsub.Experiment{
			Id:   1,
			Name: "test-exp",
		},
		Treatment: &_pubsub.ExperimentTreatment{
			Name:   "test-treatment",
			Config: treatmentCfg,
		},
		Request: &Request{},
		Segmenters: []models.SegmentFilter{
			{Key: "key", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "value"}}}},
		},
	}

	// Run newProtobufKafkaLogEntry and validate
	key, message, err := newProtobufKafkaLogEntry(assignedTreatmentLog)
	assert.NoError(t, err)

	// Unmarshall serialised key
	decodedResultLogKey := &TreatmentServiceResultLogKey{}
	err = proto.Unmarshal(key, decodedResultLogKey)
	assert.NoError(t, err)

	m := protojson.MarshalOptions{}
	actualKeyJSON, err := m.Marshal(decodedResultLogKey)
	assert.NoError(t, err)

	// Unmarshall serialised message
	decodedResultLogMessage := &TreatmentServiceResultLogMessage{}
	err = proto.Unmarshal(message, decodedResultLogMessage)
	assert.NoError(t, err)

	m = protojson.MarshalOptions{}
	actualValueJSON, err := m.Marshal(decodedResultLogMessage)
	assert.NoError(t, err)

	// Convert expected and actual log entries to JSON for comparison
	assignedTreatmentLogKeyJSON := map[string]interface{}{
		"eventTimestamp": decodedResultLogKey.EventTimestamp.AsTime(),
		"requestId":      "1",
	}
	expectedKeyJSON, err := json.Marshal(assignedTreatmentLogKeyJSON)
	assert.NoError(t, err)
	assignedTreatmentLogValueJSON := map[string]interface{}{
		"eventTimestamp":  decodedResultLogMessage.EventTimestamp.AsTime(),
		"experimentId":    "1",
		"experimentName":  "test-exp",
		"request":         map[string]interface{}{},
		"requestId":       "1",
		"segment":         "{\"key\":[\"value\"]}",
		"treatmentConfig": "{\"treatment-key\":\"treatment-value\"}",
		"treatmentName":   "test-treatment",
	}
	expectedValueJSON, err := json.Marshal(assignedTreatmentLogValueJSON)
	assert.NoError(t, err)

	// Compare logEntry data
	assert.JSONEq(t, string(expectedKeyJSON), string(actualKeyJSON))
	assert.JSONEq(t, string(expectedValueJSON), string(actualValueJSON))
}

type BQLoggerSuite struct {
	suite.Suite

	target   *bigquery.Table
	bqClient *bigquery.Client

	logger *AssignedTreatmentLogger
}

func (s *BQLoggerSuite) SetupTest() {
	var err error
	config := config.BigqueryConfig{
		Project: s.target.ProjectID,
		Dataset: s.target.DatasetID,
		Table:   s.target.TableID,
	}
	s.logger, err = NewBQAssignedTreatmentLogger(config, 100, time.Millisecond)
	if err != nil {
		panic(err)
	}
}

/*
Strictly speaking this test is broken. It needs real BQ table and ideally already created in advance, since
schema update is lagging with streaming inserts and NOT_FOUND error might occur if table was just created.

But I decided to keep it here since it's still might be useful for development and local testing.

In order to activate tests env BQ_LOGS_TARGET should be set with BQ table reference
*/
func (s *BQLoggerSuite) TestLogSentToBQ() {
	log := &AssignedTreatmentLog{
		ProjectID:  uint32(1),
		RequestID:  uuid.New().String(),
		Experiment: &_pubsub.Experiment{Id: 1, Name: "ExperimentName"},
		Treatment:  &_pubsub.ExperimentTreatment{Name: "optionA", Traffic: 30},
		Request:    &Request{},
		Segmenters: []models.SegmentFilter{
			{
				Key:   "days_of_week",
				Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(1)}}},
			},
		},
	}
	err := s.logger.Append(log)

	s.Require().Nil(err)

	s.Require().Eventually(func() bool {
		q := s.bqClient.Query(fmt.Sprintf("SELECT count(*) FROM %s.%s WHERE request_id = '%s'", s.target.DatasetID, s.target.TableID, log.RequestID))
		it, _ := q.Read(context.TODO())
		if it == nil {
			return false
		}
		var values []bigquery.Value
		_ = it.Next(&values)
		return values[0].(int64) == 1
	}, 5*time.Second, time.Second)
}

func TestTreatmentLoggerWithBQPublisherSuite(t *testing.T) {
	bqTarget := os.Getenv("BQ_LOGS_TARGET")
	if bqTarget == "" {
		t.Skip("BQ Dataset is not defined")
	}

	parts := strings.Split(bqTarget, ":")
	project := parts[0]
	table := parts[1]

	parts = strings.Split(table, ".")
	dataset := parts[0]
	table = parts[1]

	bqClient, _ := bigquery.NewClient(context.Background(), project)

	s := new(BQLoggerSuite)
	s.target = bqClient.Dataset(dataset).Table(table)
	s.bqClient = bqClient
	suite.Run(t, s)
}
