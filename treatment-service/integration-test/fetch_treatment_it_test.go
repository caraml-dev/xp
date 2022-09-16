package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/xp/clients/management"
	"github.com/caraml-dev/xp/clients/treatment"
	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/common/testutils"
	"github.com/caraml-dev/xp/treatment-service/models"
	"github.com/caraml-dev/xp/treatment-service/monitoring"
	"github.com/caraml-dev/xp/treatment-service/server"
	mgmtSvcServer "github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement/server"
	mgmtSvc "github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement/service"
)

const (
	TreatmentServerPort = 8080
	PubSubProject       = "test"
	TopicName           = "update"
)

type request struct {
	Longitude float64 `json:"longitude,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	OrderId   *string `json:"order-id,omitempty"`
	Timezone  *string `json:"tz,omitempty"`
}

var (
	orderId = "1234"
	tz      = "Asia/Singapore"
)

type TreatmentServiceTestSuite struct {
	suite.Suite

	managementServiceClient *management.ClientWithResponses
	managementServiceServer *httptest.Server

	treatmentServiceServer *server.Server
	treatmentServiceClient *treatment.ClientWithResponses

	terminationChannel chan bool
	ctx                context.Context
	emulator           testcontainers.Container
	kafka              testcontainers.DockerCompose
}

func int32Ptr(value int32) *int32 {
	return &value
}

func getStartEndTime() (time.Time, time.Time) {
	start := time.Now().AddDate(0, 0, -1)                // Start experiment a day ago
	end := time.Now().Truncate(time.Hour).Add(time.Hour) // End experiment an hour from now

	return start, end
}

func generateExperiments() []schema.Experiment {
	startTime, endTime := getStartEndTime()

	vals := map[int64]map[string]interface{}{
		1: {
			"id":           int64(1),
			"project-id":   int64(1),
			"name":         "sg-exp-1",
			"description":  "SG Experiment",
			"interval":     int32(rand.Int()),
			"days_of_week": []int64{},
			"hours_of_day": []int64{},
			"s2Ids":        []int64{3592210809859604480},
			"treatments": []schema.ExperimentTreatment{
				{
					Configuration: map[string]interface{}{
						"key1": "default-treatment-config",
					},
					Name:    "default-sg-treatment",
					Traffic: int32Ptr(50),
				},
				{
					Configuration: map[string]interface{}{
						"key1": "treatment-1-config",
					},
					Name:    "treatment-sg-1",
					Traffic: int32Ptr(25),
				},
				{
					Configuration: map[string]interface{}{
						"key1": "treatment-2-config",
					},
					Name:    "treatment-sg-2",
					Traffic: int32Ptr(25),
				},
			},
			"start-time": startTime,
			"end-time":   endTime,
			"type":       schema.ExperimentTypeAB,
		},
		2: {
			"id":           int64(2),
			"project-id":   int64(2),
			"name":         "id-exp-1",
			"description":  "ID Experiment",
			"interval":     int32(rand.Int()),
			"days_of_week": []int64{1, 2, 3, 4, 5, 6, 7},
			"hours_of_day": []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
			"s2Ids":        []int64{},
			"treatments": []schema.ExperimentTreatment{
				{
					Configuration: map[string]interface{}{
						"key1": "default-treatment-config",
					},
					Name:    "default-id-treatment",
					Traffic: int32Ptr(40),
				},
				{
					Configuration: map[string]interface{}{
						"key1": "treatment-1-config",
					},
					Name:    "treatment-id-1",
					Traffic: int32Ptr(30),
				},
				{
					Configuration: map[string]interface{}{
						"key1": "treatment-2-config",
					},
					Name:    "treatment-id-2",
					Traffic: int32Ptr(30),
				},
			},
			"start-time": startTime,
			"end-time":   endTime,
			"type":       schema.ExperimentTypeAB,
		},
		3: {
			"id":           int64(3),
			"project-id":   int64(3),
			"name":         "sg-exp-3",
			"description":  "SG Experiment",
			"interval":     int32(30),
			"days_of_week": []int64{1, 2, 3, 4, 5, 6, 7},
			"hours_of_day": []int64{},
			"s2Ids":        []int64{},
			"treatments": []schema.ExperimentTreatment{
				{
					Configuration: map[string]interface{}{
						"key1": "default-treatment-config",
					},
					Name:    "default-sg-treatment",
					Traffic: int32Ptr(70),
				},
				{
					Configuration: map[string]interface{}{
						"key1": "treatment-1-config",
					},
					Name:    "treatment-sg-1",
					Traffic: int32Ptr(20),
				},
				{
					Configuration: map[string]interface{}{
						"key1": "treatment-2-config",
					},
					Name:    "treatment-sg-2",
					Traffic: int32Ptr(10),
				},
			},
			"start-time": startTime,
			"end-time":   endTime,
			"type":       schema.ExperimentTypeSwitchback,
		},
	}

	var experiments []schema.Experiment
	for _, val := range vals {
		id := val["id"].(int64)
		projectId := val["project-id"].(int64)
		description := val["description"].(string)
		daysOfWeek := val["days_of_week"].([]int64)
		hoursOfDay := val["hours_of_day"].([]int64)
		interval := val["interval"].(int32)
		name := val["name"].(string)
		s2Ids := val["s2Ids"].([]int64)
		treatments := val["treatments"].([]schema.ExperimentTreatment)
		experimentType := val["type"].(schema.ExperimentType)
		startTime := val["start-time"].(time.Time)
		endTime := val["end-time"].(time.Time)

		experiment := schema.Experiment{
			Id:          id,
			ProjectId:   projectId,
			Description: &description,
			EndTime:     endTime,
			Interval:    &interval,
			Name:        name,
			Segment: schema.ExperimentSegment{
				"days_of_week": &daysOfWeek,
				"hours_of_day": &hoursOfDay,
				"s2_ids":       &s2Ids,
			},
			StartTime:  startTime,
			Treatments: treatments,
			Type:       experimentType,
			Version:    int64(1),
		}
		experiments = append(experiments, experiment)
	}
	return experiments
}

func setupManagementServiceClient() (*management.ClientWithResponses, *httptest.Server) {
	projectSettings := []schema.ProjectSettings{}
	segmenters := map[int64]schema.ProjectSegmenters{
		1: {
			Names: []string{"s2_ids"},
			Variables: schema.ProjectSegmenters_Variables{
				AdditionalProperties: map[string][]string{
					"s2_ids": {"latitude", "longitude"},
				},
			},
		},
		2: {
			Names: []string{"days_of_week", "hours_of_day"},
			Variables: schema.ProjectSegmenters_Variables{
				AdditionalProperties: map[string][]string{
					"days_of_week": {"tz"},
					"hours_of_day": {"tz"},
				},
			},
		},
		3: {
			Names: []string{"days_of_week"},
			Variables: schema.ProjectSegmenters_Variables{
				AdditionalProperties: map[string][]string{
					"days_of_week": {"tz"},
				},
			},
		},
		4: {
			Names: []string{},
			Variables: schema.ProjectSegmenters_Variables{
				AdditionalProperties: map[string][]string{},
			},
		},
		5: {
			Names: []string{},
			Variables: schema.ProjectSegmenters_Variables{
				AdditionalProperties: map[string][]string{},
			},
		},
	}
	for i := 1; i <= 5; i++ {
		settings := schema.ProjectSettings{
			EnableS2idClustering: false,
			ProjectId:            int64(i),
			Username:             fmt.Sprintf("ProjectSettings%v", i),
			RandomizationKey:     "order-id",
			Passkey:              "test_project_1234",
			Segmenters:           segmenters[int64(i)],
		}
		projectSettings = append(projectSettings, settings)
	}

	experiments := generateExperiments()

	segmentersType := map[string]schema.SegmenterType{
		"s2_ids":       "INTEGER",
		"days_of_week": "INTEGER",
		"hours_of_day": "INTEGER",
	}

	messageQueue, err := mgmtSvc.NewPubSubMessageQueue(mgmtSvc.PubSubConfig{
		GCPProject: PubSubProject,
		TopicName:  TopicName,
	})
	if err != nil {
		log.Fatalf("fail to instantiate message queue: %s", err.Error())
	}
	store, err := mgmtSvc.NewInMemoryStore(experiments, projectSettings, messageQueue, segmentersType)
	if err != nil {
		log.Fatalf("fail to instantiate in memory store: %s", err.Error())
	}
	managementServer := mgmtSvcServer.NewServer(store)
	managementClient, _ := management.NewClientWithResponses(managementServer.URL)

	return managementClient, managementServer
}

func setupTreatmentService(managementServiceServerURL string) (chan bool, *server.Server) {
	os.Setenv("MANAGEMENTSERVICE::URL", managementServiceServerURL)
	treatmentServer, err := server.NewServer([]string{"test.yaml"})
	if err != nil {
		log.Fatalf("fail to instantiate treatment server: %s", err.Error())
	}

	c := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-c:
				close(c)
				return
			default:
				treatmentServer.Start()
			}
		}
	}()

	return c, treatmentServer
}

func setupTreatmentServiceClient(treatmentServiceServerAddr string) *treatment.ClientWithResponses {
	addr := fmt.Sprintf("http://%s/v1", treatmentServiceServerAddr)
	treatmentClient, err := treatment.NewClientWithResponses(addr)
	if err != nil {
		log.Fatalf("fail to instantiate treatment settings: %s", err.Error())
	}

	return treatmentClient
}

func (suite *TreatmentServiceTestSuite) SetupSuite() {
	ctx := context.Background()
	os.Setenv("PORT", strconv.Itoa(TreatmentServerPort))
	os.Setenv("PROJECTIDS", "1,2,3,4,5")
	os.Setenv("PUBSUB::PROJECT", PubSubProject)
	os.Setenv("PUBSUB::TOPICNAME", TopicName)

	emulator, pubsubClient, err := testutils.StartPubSubEmulator(context.Background(), PubSubProject)
	if err != nil {
		panic(err)
	}

	err = testutils.CreatePubsubTopic(
		pubsubClient,
		ctx,
		[]string{TopicName},
	)
	if err != nil {
		panic(err)
	}

	suite.emulator = emulator

	managementClient, managementServer := setupManagementServiceClient()
	suite.managementServiceClient = managementClient
	suite.managementServiceServer = managementServer

	// Docker compose file copied from official confluentinc repository.
	// See: https://github.com/confluentinc/cp-all-in-one/blob/7.0.1-post/cp-all-in-one-kraft/docker-compose.yml
	composeFilePaths := []string{"docker-compose/kafka/docker-compose.yaml"}
	kafka := testcontainers.NewLocalDockerCompose(composeFilePaths, "kafka")
	execError := kafka.
		WithCommand([]string{"up", "-d"}).
		Invoke()
	err = execError.Error
	if err != nil {
		panic(err)
	}
	suite.kafka = kafka
	os.Setenv("ASSIGNEDTREATMENTLOGGER::KAFKACONFIG::BROKERS", "localhost:9092")

	c, treatmentServer := setupTreatmentService(suite.managementServiceServer.URL)
	waitForServerToListen := func() bool {
		conn, err := net.Dial("tcp", net.JoinHostPort("", strconv.Itoa(TreatmentServerPort)))
		if conn != nil {
			conn.Close()
		}
		return err == nil
	}
	suite.Require().Eventuallyf(waitForServerToListen, 3*time.Second, 1*time.Second, "treatment service failed to start")

	suite.terminationChannel = c
	suite.treatmentServiceServer = treatmentServer

	treatmentClient := setupTreatmentServiceClient(suite.treatmentServiceServer.Addr)
	suite.treatmentServiceClient = treatmentClient

	suite.ctx = context.Background()
}

func TestTreatmentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TreatmentServiceTestSuite))
}

func (suite *TreatmentServiceTestSuite) TearDownSuite() {
	suite.treatmentServiceServer.Close()
	suite.terminationChannel <- true

	_ = suite.kafka.Down()
	_ = suite.emulator.Terminate(context.Background())
}

func (suite *TreatmentServiceTestSuite) TestAdditionalFilters() {
	projectId := int64(1)
	params := treatment.FetchTreatmentParams{PassKey: "test_project_1234"}

	postBody, _ := json.Marshal(request{
		Longitude: 103.8998991137485,
		Latitude:  1.2537040223936706,
		OrderId:   &orderId,
	})
	requestReader := bytes.NewReader(postBody)
	resp, err := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		requestReader,
	)

	expectedBody := schema.SelectedTreatment{
		ExperimentName: "sg-exp-1",
		ExperimentId:   1,
		Treatment: schema.SelectedTreatmentData{
			Name: "default-sg-treatment",
			Configuration: map[string]interface{}{
				"key1": "default-treatment-config",
			},
			Traffic: int32Ptr(50),
		},
		Metadata: schema.SelectedTreatmentMetadata{
			ExperimentVersion: int64(1),
			ExperimentType:    schema.ExperimentTypeAB,
		},
	}

	suite.Require().NoError(err)
	suite.Require().Equal(200, resp.StatusCode())
	suite.Require().Equal(expectedBody, *resp.JSON200.Data)

	consumer, err := kafka.NewConsumer(
		&kafka.ConfigMap{
			"bootstrap.servers":    "localhost:9092",
			"group.id":             "test-group",
			"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"},
		})
	suite.Require().NoError(err)
	err = consumer.Subscribe("local-testing", nil)
	suite.Require().NoError(err)

	startTime := time.Now()
	decodedResultLogMessage := &monitoring.TreatmentServiceResultLogMessage{}
W:
	for {
		ev := consumer.Poll(1000)
		switch e := ev.(type) {
		case *kafka.Message:
			fmt.Printf("%% Message on %s:\n%s\n",
				e.TopicPartition, string(e.Value))
			err = proto.Unmarshal(e.Value, decodedResultLogMessage)
			break W
		case kafka.PartitionEOF:
			fmt.Printf("%% Reached %v\n", e)
			break W
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			break W
		default:
			if time.Since(startTime).Seconds() > 10 {
				break W
			}
		}
	}
	consumer.Close()
	suite.Require().NoError(err)
	suite.Require().Equal(
		"{\"s2_ids\":[3592210809859604480,3592210814154571776,3592210796974702592,3592210865694179328,3592211140572086272]}",
		decodedResultLogMessage.Segment,
	)
}

func (suite *TreatmentServiceTestSuite) TestTimeFilter() {
	projectId := int64(2)
	params := treatment.FetchTreatmentParams{PassKey: "test_project_1234"}

	postBody, _ := json.Marshal(request{
		OrderId:  &orderId,
		Timezone: &tz,
	})
	requestReader := bytes.NewReader(postBody)
	resp, err := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		requestReader,
	)

	expectedBody := schema.SelectedTreatment{
		ExperimentName: "id-exp-1",
		ExperimentId:   2,
		Treatment: schema.SelectedTreatmentData{
			Name: "default-id-treatment",
			Configuration: map[string]interface{}{
				"key1": "default-treatment-config",
			},
			Traffic: int32Ptr(40),
		},
		Metadata: schema.SelectedTreatmentMetadata{
			ExperimentVersion: int64(1),
			ExperimentType:    schema.ExperimentTypeAB,
		},
	}

	suite.Require().NoError(err)
	suite.Require().Equal(200, resp.StatusCode())
	suite.Require().Equal(expectedBody, *resp.JSON200.Data)
}

func (suite *TreatmentServiceTestSuite) TestExperimentUpdates() {
	projectId := int64(4)
	experimentId := int64(4)
	traffic := int32Ptr(1)
	interval := int32Ptr(1)
	orderId := "1234"
	s2ids := []int64{3592210809859604480}
	experimentName := "new-sg-exp"
	treatmentName := "new-sg-treatment"
	updatedBy := ""
	body := management.CreateExperimentJSONRequestBody{
		EndTime: time.Now().Add(time.Duration(24) * time.Hour),
		Name:    experimentName,
		Segment: schema.ExperimentSegment{
			"s2_ids": &s2ids,
		},
		StartTime: time.Now(),
		Interval:  interval,
		Treatments: []schema.ExperimentTreatment{
			{
				Name:    treatmentName,
				Traffic: traffic,
				Configuration: map[string]interface{}{
					"key1": "new-treatment-config",
				},
			},
		},
		UpdatedBy: &updatedBy,
	}
	response, err := suite.managementServiceClient.CreateExperimentWithResponse(suite.ctx, projectId, body)
	suite.Require().Equal(http.StatusOK, response.StatusCode())
	suite.Require().NoError(err)

	time.Sleep(3 * time.Second)
	params := treatment.FetchTreatmentParams{
		PassKey: "test_project_1234",
	}

	reqBody := request{
		OrderId: &orderId,
	}
	postBody, _ := json.Marshal(reqBody)
	resp, err := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		bytes.NewReader(postBody),
	)

	expectedBody := schema.SelectedTreatment{
		ExperimentName: "new-sg-exp",
		ExperimentId:   experimentId,
		Treatment: schema.SelectedTreatmentData{
			Name: "new-sg-treatment",
			Configuration: map[string]interface{}{
				"key1": "new-treatment-config",
			},
			Traffic: traffic,
		},
		Metadata: schema.SelectedTreatmentMetadata{
			ExperimentVersion: int64(1),
			ExperimentType:    schema.ExperimentTypeAB,
		},
	}

	suite.Require().NoError(err)
	suite.Require().Equal(200, resp.StatusCode())
	suite.Require().Equal(expectedBody, *resp.JSON200.Data)

	newTraffic := int32Ptr(11)
	newDescription := "updated description"
	updateExperimentBody := management.UpdateExperimentJSONRequestBody{
		EndTime:     time.Now().Add(time.Duration(24) * time.Hour),
		Description: &newDescription,
		Segment: schema.ExperimentSegment{
			"s2_ids": &s2ids,
		},
		Interval: interval,
		Treatments: []schema.ExperimentTreatment{
			{
				Name:    treatmentName,
				Traffic: newTraffic,
				Configuration: map[string]interface{}{
					"key1": "new-treatment-config",
				},
			},
		},
	}
	updateExperimentResponse, err := suite.managementServiceClient.UpdateExperimentWithResponse(suite.ctx, projectId, experimentId, updateExperimentBody)
	suite.Require().Equal(http.StatusOK, updateExperimentResponse.StatusCode())
	suite.Require().NoError(err)

	time.Sleep(3 * time.Second)
	params = treatment.FetchTreatmentParams{
		PassKey: "test_project_1234",
	}

	updatedResp, err := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		bytes.NewReader(postBody),
	)

	expectedBody = schema.SelectedTreatment{
		ExperimentName: "new-sg-exp",
		ExperimentId:   experimentId,
		Treatment: schema.SelectedTreatmentData{
			Name: "new-sg-treatment",
			Configuration: map[string]interface{}{
				"key1": "new-treatment-config",
			},
			Traffic: newTraffic,
		},
		Metadata: schema.SelectedTreatmentMetadata{
			ExperimentVersion: int64(2),
			ExperimentType:    schema.ExperimentTypeAB,
		},
	}

	suite.Require().NoError(err)
	suite.Require().Equal(200, resp.StatusCode())
	suite.Require().Equal(expectedBody, *updatedResp.JSON200.Data)

	updateExperimentBody = management.UpdateExperimentJSONRequestBody{
		Status: "inactive",
	}
	disableExpResp, err := suite.managementServiceClient.UpdateExperiment(
		suite.ctx, projectId, resp.JSON200.Data.ExperimentId, updateExperimentBody,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, disableExpResp.StatusCode)
	_ = disableExpResp.Body.Close()

	time.Sleep(3 * time.Second)
	resp, err = suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		bytes.NewReader(postBody),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, resp.StatusCode())
	suite.Require().Nil(resp.JSON200.Data)
}

func (suite *TreatmentServiceTestSuite) TestProjectSettingsUpdates() {
	projectId := int64(5)
	projectSegmenters := schema.ProjectSegmenters{
		Names: []string{"s2_ids"},
		Variables: schema.ProjectSegmenters_Variables{
			AdditionalProperties: map[string][]string{
				"s2_ids": {"latitude", "longitude"},
			},
		},
	}

	updateProjectSettingsResponse, err := suite.managementServiceClient.UpdateProjectSettingsWithResponse(
		suite.ctx, projectId, management.UpdateProjectSettingsJSONRequestBody{
			EnableS2idClustering: nil,
			RandomizationKey:     "order-id",
			Segmenters:           projectSegmenters,
		})
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, updateProjectSettingsResponse.StatusCode())
	time.Sleep(3 * time.Second)

	params := treatment.FetchTreatmentParams{
		PassKey: "test_project_1234",
	}

	reqBody := request{
		Longitude: 103.8998991137485,
		Latitude:  1.2537040223936706,
		OrderId:   &orderId,
	}
	postBody, _ := json.Marshal(reqBody)
	fetchTreatmentResponse, err := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		bytes.NewReader(postBody),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusBadRequest, fetchTreatmentResponse.StatusCode())
}

func (suite *TreatmentServiceTestSuite) TestNoExperiment() {
	projectId := int64(1)
	params := treatment.FetchTreatmentParams{PassKey: "test_project_1234"}

	// Incorrect Longitude value hence no matching experiment
	postBody, _ := json.Marshal(request{
		Longitude: 179.8998991137485,
		Latitude:  1.2537040223936706,
		OrderId:   &orderId,
		Timezone:  &tz,
	})
	requestReader := bytes.NewReader(postBody)
	resp, _ := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		requestReader,
	)

	suite.Require().Equal(200, resp.StatusCode())
	suite.Require().Nil(resp.JSON200.Data)
}

func (suite *TreatmentServiceTestSuite) TestIncorrectRequestParam() {
	projectId := int64(1)
	params := treatment.FetchTreatmentParams{PassKey: "test_project_1234"}
	requestReader := bytes.NewReader([]byte(`{"latitude": 103.45, "longitude": "*"}`))
	resp, err := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		requestReader,
	)

	suite.Require().NoError(err)
	suite.Require().Equal(400, resp.StatusCode())
	suite.Require().Equal("unable to cast \"*\" of type string to float64", resp.JSON400.Error)
}

func (suite *TreatmentServiceTestSuite) TestAllFiltersSwitchback() {
	projectId := int64(3)
	params := treatment.FetchTreatmentParams{PassKey: "test_project_1234"}

	postBody, _ := json.Marshal(request{
		Longitude: 103.8998991137485,
		Latitude:  1.2537040223936706,
		OrderId:   &orderId,
		Timezone:  &tz,
	})
	requestReader := bytes.NewReader(postBody)
	resp, err := suite.treatmentServiceClient.FetchTreatmentWithBodyWithResponse(
		suite.ctx,
		projectId,
		&params,
		"application/json",
		requestReader,
	)

	// Calculate switchback window id
	startTime, _ := getStartEndTime()
	windowId := int64(math.Floor(time.Since(startTime).Minutes() / float64(30)))
	expectedBody := schema.SelectedTreatment{
		ExperimentName: "sg-exp-3",
		ExperimentId:   3,
		Treatment: schema.SelectedTreatmentData{
			Name: "default-sg-treatment",
			Configuration: map[string]interface{}{
				"key1": "default-treatment-config",
			},
			Traffic: int32Ptr(70),
		},
		Metadata: schema.SelectedTreatmentMetadata{
			ExperimentVersion:  int64(1),
			ExperimentType:     schema.ExperimentTypeSwitchback,
			SwitchbackWindowId: &windowId,
		},
	}

	suite.Require().NoError(err)
	suite.Require().Equal(200, resp.StatusCode())
	suite.Require().Equal(expectedBody, *resp.JSON200.Data)
}

func (suite *TreatmentServiceTestSuite) TestLocalStorage() {
	storage, err := models.NewLocalStorage([]models.ProjectId{1}, suite.managementServiceServer.URL, false)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(storage)

	resp, err := suite.managementServiceClient.GetExperimentWithResponse(suite.ctx, 1, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(resp.JSON200.Data.Name, storage.Experiments[1][0].Experiment.Name)
}
