package config

import (
	"fmt"
	"strconv"

	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"

	common_config "github.com/gojek/turing-experiments/common/config"
	"github.com/gojek/turing-experiments/treatment-service/models"
)

type AssignedTreatmentLoggerKind = string

const (
	KafkaLogger AssignedTreatmentLoggerKind = "kafka"
	BQLogger    AssignedTreatmentLoggerKind = "bq"
	NoopLogger  AssignedTreatmentLoggerKind = ""
)

type Config struct {
	Port       int      `default:"8080"`
	ProjectIds []string `default:""`

	AssignedTreatmentLogger AssignedTreatmentLoggerConfig
	DebugConfig             DebugConfig
	NewRelicConfig          newrelic.Config
	SentryConfig            sentry.Config
	DeploymentConfig        DeploymentConfig
	PubSub                  PubSub
	ManagementService       ManagementServiceConfig
	MonitoringConfig        Monitoring
	SwaggerConfig           SwaggerConfig
	SegmenterConfig         map[string]interface{}
}

type AssignedTreatmentLoggerConfig struct {
	Kind                 AssignedTreatmentLoggerKind `default:""`
	QueueLength          int                         `default:"1073741824"`
	FlushIntervalSeconds int                         `default:"1"`

	BQConfig    *BigqueryConfig
	KafkaConfig *KafkaConfig
}

type BigqueryConfig struct {
	Project string
	Dataset string
	Table   string
}

type KafkaConfig struct {
	Brokers          string
	Topic            string
	MaxMessageBytes  int    `default:"1048588"`
	CompressionType  string `default:"none"`
	ConnectTimeoutMS int    `default:"1000"`
}

type DebugConfig struct {
	OutputPath string `default:"/tmp"`
}

type SwaggerConfig struct {
	Enabled          bool     `default:"false"`
	AllowedOrigins   []string `default:"*"`
	OpenAPISpecsPath string   `default:"."`
}

// DeploymentConfig captures the config related to the deployment of Treatment Service
type DeploymentConfig struct {
	EnvironmentType string `default:"local"`
}

type MetricSinkKind = string

const (
	PrometheusMetricSink MetricSinkKind = "prometheus"
	NoopMetricSink       MetricSinkKind = ""
)

type Monitoring struct {
	Kind         MetricSinkKind `default:""`
	MetricLabels []string       `default:""`
}

type PubSub struct {
	Project              string `default:"dev"`
	TopicName            string `default:"xp-update"`
	PubSubTimeoutSeconds int    `default:"30"`
}

type ManagementServiceConfig struct {
	URL                  string `default:"http://localhost:3000/v1"`
	AuthorizationEnabled bool
}

func (c *Config) GetProjectIds() []models.ProjectId {
	projectIds := make([]models.ProjectId, 0)
	for _, projectIdString := range c.ProjectIds {
		projectId, _ := strconv.Atoi(projectIdString)
		projectIds = append(projectIds, uint32(projectId))
	}

	return projectIds
}

// ListenAddress returns the Treatment API app's port
func (c *Config) ListenAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

func Load(filepaths ...string) (*Config, error) {
	var cfg Config
	err := common_config.ParseConfig(&cfg, filepaths)
	if err != nil {
		return nil, fmt.Errorf("failed to update viper config: %s", err)
	}

	return &cfg, nil
}
