package config

import (
	"fmt"
	"time"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"

	common_config "github.com/caraml-dev/xp/common/config"
	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
)

type Config struct {
	OpenAPISpecsPath string `default:"."`
	Port             int    `default:"3000"`

	AllowedOrigins      []string `default:"*"`
	AuthorizationConfig *AuthorizationConfig
	DbConfig            *DatabaseConfig
	MLPConfig           *MLPConfig
	MessageQueueConfig  *common_mq_config.MessageQueueConfig
	SegmenterConfig     map[string]interface{}
	ValidationConfig    ValidationConfig
	DeploymentConfig    DeploymentConfig
	NewRelicConfig      newrelic.Config
	SentryConfig        sentry.Config
	XpUIConfig          *XpUIConfig
	PollerConfig        *PollerConfig
}

// AuthorizationConfig captures the config for MLP authz
type AuthorizationConfig struct {
	Enabled bool
	URL     string
	Caching *InMemoryCacheConfig
}

type InMemoryCacheConfig struct {
	Enabled                     bool
	KeyExpirySeconds            int `default:"600"`
	CacheCleanUpIntervalSeconds int `default:"900"`
}

// DatabaseConfig captures the XP database config
type DatabaseConfig struct {
	Host           string `default:"localhost"`
	Port           int    `default:"5432"`
	User           string `default:"xp"`
	Password       string `default:"xp"`
	Database       string `default:"xp"`
	MigrationsPath string `default:"file://database/db-migrations"`

	ConnMaxIdleTime time.Duration `default:"0s"`
	ConnMaxLifetime time.Duration `default:"0s"`
	MaxIdleConns    int           `default:"0"`
	MaxOpenConns    int           `default:"0"`
}

// MLPConfig captures the configuration used to connect to the MLP API server
type MLPConfig struct {
	URL string
}

// ValidationConfig captures the config related to the validation of schemas
type ValidationConfig struct {
	ValidationUrlTimeoutSeconds int `default:"5"`
}

// DeploymentConfig captures the config related to the deployment of Management Service
type DeploymentConfig struct {
	EnvironmentType string `default:"local"`
}

// XpUIConfig captures config related to serving XP UI files
type XpUIConfig struct {
	// Optional. If configured, xp management service API will serve static files
	// of the xp-ui React app.
	AppDirectory string
	// Optional. Defines the relative path under which the app will be accessible.
	// This should match `homepage` value from the `package.json` file of the CRA app
	Homepage string `default:"/xp"`
}

type PollerConfig struct {
	Enabled bool `default:"false"`
}

// ListenAddress returns the Management API app's port
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
