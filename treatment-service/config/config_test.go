package config

import (
	"testing"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
)

func TestDefaultConfigs(t *testing.T) {
	emptyInterfaceMap := make(map[string]interface{})
	emptyStringMap := make(map[string]string)
	defaultCfg := Config{
		Port: 8080,
		SwaggerConfig: SwaggerConfig{
			Enabled:          false,
			AllowedOrigins:   []string{"*"},
			OpenAPISpecsPath: ".",
		},
		ProjectIds: []string{},
		ManagementService: ManagementServiceConfig{
			URL:                  "http://localhost:3000/v1",
			AuthorizationEnabled: false,
		},
		DeploymentConfig: DeploymentConfig{
			EnvironmentType: "local",
			MaxGoRoutines:   100,
		},
		AssignedTreatmentLogger: AssignedTreatmentLoggerConfig{
			Kind:                 "",
			QueueLength:          100,
			FlushIntervalSeconds: 1,
			BQConfig:             &BigqueryConfig{},
			KafkaConfig: &KafkaConfig{
				Brokers:          "",
				Topic:            "",
				MaxMessageBytes:  1048588,
				CompressionType:  "none",
				ConnectTimeoutMS: 1000,
			},
		},
		DebugConfig: DebugConfig{
			OutputPath: "/tmp",
		},
		MessageQueueConfig: &common_mq_config.MessageQueueConfig{
			Kind: "",
			PubSubConfig: &common_mq_config.PubSubConfig{
				Project:              "dev",
				TopicName:            "xp-update",
				PubSubTimeoutSeconds: 30,
			},
		},
		MonitoringConfig: Monitoring{MetricLabels: []string{}},
		NewRelicConfig: newrelic.Config{
			Enabled:           false,
			AppName:           "",
			License:           "",
			IgnoreStatusCodes: []int{},
			Labels:            emptyInterfaceMap,
		},
		SentryConfig:    sentry.Config{Enabled: false, Labels: emptyStringMap},
		SegmenterConfig: make(map[string]interface{}),
	}
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, defaultCfg, *cfg)
	assert.Equal(t, defaultCfg.ListenAddress(), cfg.ListenAddress())
	assert.Equal(t, defaultCfg.GetProjectIds(), cfg.GetProjectIds())
}

func TestLoadMultipleConfigs(t *testing.T) {
	configFiles := []string{"../testdata/config1.yaml", "../testdata/config2.yaml"}
	expected := Config{
		Port: 8080,
		SwaggerConfig: SwaggerConfig{
			Enabled:          false,
			AllowedOrigins:   []string{"host-1", "host-2"},
			OpenAPISpecsPath: "test-path",
		},
		ProjectIds: []string{"1", "2"},
		ManagementService: ManagementServiceConfig{
			URL:                  "localhost:3000/v1",
			AuthorizationEnabled: true,
		},
		DeploymentConfig: DeploymentConfig{
			EnvironmentType:                    "dev",
			MaxGoRoutines:                      200,
			GoogleApplicationCredentialsEnvVar: "GOOGLE_APPLICATION_CREDENTIALS_EXPERIMENT_ENGINE",
		},
		AssignedTreatmentLogger: AssignedTreatmentLoggerConfig{
			Kind:                 "bq",
			QueueLength:          100,
			FlushIntervalSeconds: 1,
			BQConfig: &BigqueryConfig{
				Project: "dev",
				Dataset: "xp-test-dataset",
				Table:   "xp-test-table",
			},
			KafkaConfig: &KafkaConfig{
				MaxMessageBytes:  1048588,
				CompressionType:  "none",
				ConnectTimeoutMS: 1000,
			},
		},
		DebugConfig: DebugConfig{
			OutputPath: "/tmp1",
		},
		MessageQueueConfig: &common_mq_config.MessageQueueConfig{
			Kind: "",
			PubSubConfig: &common_mq_config.PubSubConfig{
				Project:              "dev",
				TopicName:            "xp-update",
				PubSubTimeoutSeconds: 30,
			},
		},
		MonitoringConfig: Monitoring{MetricLabels: []string{}},
		NewRelicConfig: newrelic.Config{
			Enabled:           true,
			AppName:           "xp-treatment-service-test",
			License:           "amazing-license",
			IgnoreStatusCodes: []int{403, 404, 405},
			Labels:            map[string]interface{}{"env": "dev"},
		},
		SentryConfig:    sentry.Config{Enabled: true, DSN: "my.amazing.sentry.dsn", Labels: map[string]string{"app": "xp-treatment-service"}},
		SegmenterConfig: map[string]interface{}{"s2_ids": map[string]interface{}{"mins2celllevel": 9, "maxs2celllevel": 15}},
	}

	cfg, err := Load(configFiles...)
	require.NoError(t, err)
	assert.Equal(t, expected, *cfg)
	assert.Equal(t, cfg.GetProjectIds(), []uint32{1, 2})
}

func TestMissingConfigs(t *testing.T) {
	cfg, err := Load("")
	require.Nil(t, cfg)
	require.Error(t, err, "failed to update viper config:")
}
