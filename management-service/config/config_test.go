package config

import (
	"strings"
	"testing"
	"time"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigs(t *testing.T) {
	zeroSecond, _ := time.ParseDuration("0s")
	emptyInterfaceMap := make(map[string]interface{})
	emptyStringMap := make(map[string]string)
	defaultCfg := Config{
		Port:           3000,
		AllowedOrigins: []string{"*"},
		DbConfig: &DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "xp",
			Password:        "xp",
			Database:        "xp",
			MigrationsPath:  "file://database/db-migrations",
			ConnMaxIdleTime: zeroSecond,
			ConnMaxLifetime: zeroSecond,
			MaxIdleConns:    0,
			MaxOpenConns:    0,
		},
		SegmenterConfig: make(map[string]interface{}),
		MLPConfig: &MLPConfig{
			URL: "",
		},
		MessageQueueConfig: &common_mq_config.MessageQueueConfig{
			Kind: "",
			PubSubConfig: &common_mq_config.PubSubConfig{
				Project:              "dev",
				TopicName:            "xp-update",
				PubSubTimeoutSeconds: 30,
			},
		},
		ValidationConfig: ValidationConfig{
			ValidationUrlTimeoutSeconds: 5,
		},
		OpenAPISpecsPath: ".",
		DeploymentConfig: DeploymentConfig{
			EnvironmentType: "local",
		},
		NewRelicConfig: newrelic.Config{
			Enabled:           false,
			AppName:           "",
			License:           "",
			IgnoreStatusCodes: []int{},
			Labels:            emptyInterfaceMap,
		},
		SentryConfig: sentry.Config{Enabled: false, Labels: emptyStringMap},
		XpUIConfig: &XpUIConfig{
			Homepage: "/xp",
		},
	}
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, defaultCfg, *cfg)
	assert.Equal(t, ":3000", cfg.ListenAddress())
}

// TestLoadConfigFiles verifies that when multiple configs are passed in
// they are consumed in the correct order
func TestLoadConfigFiles(t *testing.T) {
	oneSecond, _ := time.ParseDuration("1s")
	twoSecond, _ := time.ParseDuration("2s")
	tests := []struct {
		name        string
		configFiles []string
		errString   string
		expected    Config
	}{
		{
			name:        "success | load multiple config files",
			configFiles: []string{"../testdata/config1.yaml", "../testdata/config2.yaml"},
			expected: Config{
				Port:           3000,
				AllowedOrigins: []string{"host-1", "host-2"},
				DbConfig: &DatabaseConfig{
					Host:            "localhost",
					Port:            5432,
					User:            "admin",
					Password:        "password",
					Database:        "xp",
					MigrationsPath:  "file://test-db-migrations",
					ConnMaxIdleTime: oneSecond,
					ConnMaxLifetime: twoSecond,
					MaxIdleConns:    3,
					MaxOpenConns:    4,
				},
				SegmenterConfig: map[string]interface{}{
					"s2_ids": map[string]interface{}{
						"mins2celllevel": 9,
						"maxs2celllevel": 12,
					},
				},
				MLPConfig: &MLPConfig{
					URL: "test-mlp-url",
				},
				MessageQueueConfig: &common_mq_config.MessageQueueConfig{
					Kind: "pubsub",
					PubSubConfig: &common_mq_config.PubSubConfig{
						Project:              "test-pubsub-project",
						TopicName:            "test-pubsub-topic",
						PubSubTimeoutSeconds: 30,
					},
				},
				ValidationConfig: ValidationConfig{
					ValidationUrlTimeoutSeconds: 5,
				},
				OpenAPISpecsPath: "test-path",
				DeploymentConfig: DeploymentConfig{
					EnvironmentType: "dev",
				},
				NewRelicConfig: newrelic.Config{
					Enabled:           true,
					AppName:           "xp",
					License:           "amazing-license",
					IgnoreStatusCodes: []int{403, 404, 405},
					Labels:            map[string]interface{}{"env": "dev"},
				},
				SentryConfig: sentry.Config{Enabled: false, Labels: map[string]string{"app": "xp-management-service"}},
				XpUIConfig: &XpUIConfig{
					AppDirectory: "ui",
					Homepage:     "/testxp",
				},
			},
		},
		{
			name:        "failure | bad config",
			configFiles: []string{"../testdata/config3.yaml"},
			errString: strings.Join([]string{"failed to update viper config: failed to unmarshal config values: 1 error(s) decoding:\n\n* cannot ",
				"parse 'DbConfig.Port' as int: strconv.ParseInt: parsing \"abc\": invalid syntax"}, ""),
		},
		{
			name:        "failure | file read",
			configFiles: []string{"../testdata/config4.yaml"},
			errString: strings.Join([]string{"failed to update viper config: failed to read config from file '../testdata/config4.yaml': ",
				"While parsing config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `bad_config` ",
				"into map[string]interface {}"}, ""),
		},
	}

	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			cfg, err := Load(data.configFiles...)
			if data.errString == "" {
				// Success
				require.NoError(t, err)
				assert.Equal(t, data.expected, *cfg)
			} else {
				assert.EqualError(t, err, data.errString)
				assert.Nil(t, cfg)
			}
		})
	}
}
