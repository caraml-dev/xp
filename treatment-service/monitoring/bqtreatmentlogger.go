package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	_utils "github.com/caraml-dev/xp/common/utils"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/util"
	"go.einride.tech/protobuf-bigquery/encoding/protobq"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type BQLogPublisher struct {
	table *bigquery.Table
}

type BQLogRow struct {
	EventTimestamp time.Time `bigquery:"event_timestamp"`

	ProjectId models.ProjectId `bigquery:"project_id"`
	RequestId string           `bigquery:"request_id"`

	Request *Request `bigquery:"request"`
	Segment string   `bigquery:"segment"`

	ExperimentId   int64  `bigquery:"experiment_id"`
	ExperimentName string `bigquery:"experiment_name"`

	TreatmentName     string `bigquery:"treatment_name"`
	TreatmentConfig   string `bigquery:"treatment_config"`
	TreatmentMetadata string `bigquery:"treatment_metadata"`

	Error string `bigquery:"error"`
}

func (p *BQLogPublisher) Publish(logs []*AssignedTreatmentLog) error {
	bqLogs := make([]*BQLogRow, 0)

	for _, l := range logs {
		segments := make(map[string]interface{})
		for _, s := range l.Segmenters {
			allValues := []interface{}{}
			for _, v := range s.Value {
				allValues = append(allValues, _utils.SegmenterValueToInterface(v))
			}
			segments[s.Key] = allValues
		}

		segmentsJson, err := json.Marshal(segments)
		if err != nil {
			return err
		}

		bqlogRow := &BQLogRow{
			EventTimestamp: time.Now(),
			ProjectId:      l.ProjectID,
			RequestId:      l.RequestID,
			Request:        l.Request,
			Segment:        string(segmentsJson),
		}

		if l.Experiment != nil {
			bqlogRow.ExperimentId = l.Experiment.Id
			bqlogRow.ExperimentName = l.Experiment.Name
		}

		if l.Treatment != nil {
			treatmentConfigJson, err := json.Marshal(l.Treatment.Config)
			if err != nil {
				return err
			}
			treatmentConfig := string(treatmentConfigJson)

			bqlogRow.TreatmentName = l.Treatment.Name
			bqlogRow.TreatmentConfig = treatmentConfig
		}

		if l.TreatmentMetadata != nil {
			treatmentMetadata, err := json.Marshal(l.TreatmentMetadata)
			if err != nil {
				return err
			}
			bqlogRow.TreatmentMetadata = string(treatmentMetadata)
		}

		if l.Error != nil {
			errorJson, err := json.Marshal(l.Error)
			if err != nil {
				return err
			}

			bqlogRow.Error = string(errorJson)
		}

		bqLogs = append(bqLogs, bqlogRow)
	}
	return p.table.Inserter().Put(context.TODO(), bqLogs)
}

func createBQTable(
	ctx *context.Context,
	table *bigquery.Table,
	schema *bigquery.Schema,
) error {
	// Set partitioning
	metaData := &bigquery.TableMetadata{
		Schema: *schema,
		TimePartitioning: &bigquery.TimePartitioning{
			Field: "event_timestamp",
		},
		RequirePartitionFilter: false,
	}

	if err := table.Create(*ctx, metaData); err != nil {
		return err
	}

	ok := util.WaitFor(func() bool {
		_, err := table.Metadata(*ctx)
		return err == nil
	}, time.Minute, time.Second)
	if !ok {
		return fmt.Errorf("Couldn't create BQ Log table: %v", table)
	}

	return nil
}

func setupBQTable(
	ctx *context.Context,
	bqClient *bigquery.Client,
	schema *bigquery.Schema,
	datasetName string,
	tableName string,
) (*bigquery.Table, error) {
	// Check that dataset exists
	dataset := bqClient.Dataset(datasetName)
	_, err := dataset.Metadata(*ctx)
	if err != nil {
		return nil, fmt.Errorf("BigQuery dataset %s not found", datasetName)
	}

	// Check if the table exists
	table := dataset.Table(tableName)
	_, err = table.Metadata(*ctx)

	// If not, create
	if err != nil && err.(*googleapi.Error).Code == 404 {
		err = createBQTable(ctx, table, schema)
		if err != nil {
			return nil, err
		}
	}

	return table, nil
}

func NewBQLogPublisher(project string, dataset string, tableName string) (*BQLogPublisher, error) {
	ctx := context.Background()

	var client *bigquery.Client
	var err error
	if filepath := os.Getenv(models.ExpGoogleApplicationCredentials); filepath != "" {
		client, err = bigquery.NewClient(ctx, project, option.WithCredentialsFile(filepath))
	} else {
		client, err = bigquery.NewClient(ctx, project)
	}

	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}

	schema := getLogResultTableSchema()

	table, err := setupBQTable(&ctx, client, schema, dataset, tableName)
	if err != nil {
		return nil, err
	}

	return &BQLogPublisher{table: table}, nil
}

// getLogResultTableSchema returns the expected schema defined for logging results to BigQuery
func getLogResultTableSchema() *bigquery.Schema {
	schema := protobq.InferSchema(&TreatmentServiceResultLogMessage{})
	return &schema
}
