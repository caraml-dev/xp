package config

import (
	"fmt"
	"strconv"

	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"

	common_config "github.com/caraml-dev/xp/common/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type AssignedTreatmentLoggerKind = string

const (
	KafkaLogger AssignedTreatmentLoggerKind = "kafka"
	BQLogger    AssignedTreatmentLoggerKind = "bq"
	NoopLogger  AssignedTreatmentLoggerKind = ""
)

type Config struct {
	Port       int      `json:"port" default:"8080"`
	ProjectIds []string `json:"project_ids" default:""`

	AssignedTreatmentLogger AssignedTreatmentLoggerConfig `json:"assigned_treatment_logger"`
	DebugConfig             DebugConfig                   `json:"debug_config"`
	NewRelicConfig          newrelic.Config               `json:"new_relic_config"`
	SentryConfig            sentry.Config                 `json:"sentry_config"`
	DeploymentConfig        DeploymentConfig              `json:"deployment_config"`
	PubSub                  PubSub                        `json:"pub_sub"`
	ManagementService       ManagementServiceConfig       `json:"management_service"`
	MonitoringConfig        Monitoring                    `json:"monitoring_config"`
	SwaggerConfig           SwaggerConfig                 `json:"swagger_config"`
	SegmenterConfig         map[string]interface{}        `json:"segmenter_config"`
}

type AssignedTreatmentLoggerConfig struct {
	Kind                 AssignedTreatmentLoggerKind `json:"kind" default:""`
	QueueLength          int                         `json:"queue_length" default:"100"`
	FlushIntervalSeconds int                         `json:"flush_interval_seconds" default:"1"`

	BQConfig    *BigqueryConfig `json:"bq_config"`
	KafkaConfig *KafkaConfig    `json:"kafka_config"`
}

type BigqueryConfig struct {
	Project string `json:"project"`
	Dataset string `json:"dataset"`
	Table   string `json:"table"`
}

type KafkaConfig struct {
	Brokers          string `json:"brokers"`
	Topic            string `json:"topics"`
	MaxMessageBytes  int    `json:"max_message_bytes" default:"1048588"`
	CompressionType  string `json:"compression_type" default:"none"`
	ConnectTimeoutMS int    `json:"connect_timeout_ms" default:"1000"`
}

type DebugConfig struct {
	OutputPath string `json:"output_path" default:"/tmp"`
}

type SwaggerConfig struct {
	Enabled          bool     `json:"enabled" default:"false"`
	AllowedOrigins   []string `json:"allowed_origins" default:"*"`
	OpenAPISpecsPath string   `json:"open_api_specs_path" default:"."`
}

// DeploymentConfig captures the config related to the deployment of Treatment Service
type DeploymentConfig struct {
	EnvironmentType string `json:"environment_type" default:"local"`
	MaxGoRoutines   int    `json:"max_go_routines" default:"100"`
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
	Project              string `json:"project" default:"dev"`
	TopicName            string `json:"topic_name" default:"xp-update"`
	PubSubTimeoutSeconds int    `json:"pub_sub_timeout_seconds" default:"30"`
}

type ManagementServiceConfig struct {
	URL                  string `json:"url" default:"http://localhost:3000/v1"`
	AuthorizationEnabled bool   `json:"authorization_enabled"`
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
