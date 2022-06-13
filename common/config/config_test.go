package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	tu "github.com/gojek/turing-experiments/common/testutils"
)

func TestDefaultConfigs(t *testing.T) {
	host := "localhost"
	port := 3000
	pubSubProject := "dev"
	pubSubTopicName := "update"
	sentryUrl := "https://xx.xx.xx"
	sentryLabels := map[string]string{}

	cfg := tu.Config{}
	var filePaths []string
	err := ParseConfig(&cfg, filePaths)

	require.NoError(t, err)
	require.Equal(t, host, cfg.Host)
	require.Equal(t, port, cfg.Port)
	require.Equal(t, pubSubProject, cfg.PubSub.Project)
	require.Equal(t, pubSubTopicName, cfg.PubSub.TopicName)
	require.Equal(t, sentryUrl, cfg.Sentry.Url)
	require.Equal(t, sentryLabels, cfg.Sentry.Labels)
}

func TestFileConfigs(t *testing.T) {
	host := "localhost"
	port := 3030
	pubSubProject := "test-project"
	pubSubTopicName := "test-topic"
	sentryUrl := "https://yy.yy.yy"
	sentryLabels := map[string]string{"env": "local"}

	cfg := tu.Config{}
	filePaths := []string{"../testdata/config1.yaml"}

	err := ParseConfig(&cfg, filePaths)
	require.NoError(t, err)
	require.Equal(t, host, cfg.Host)
	require.Equal(t, port, cfg.Port)
	require.Equal(t, pubSubProject, cfg.PubSub.Project)
	require.Equal(t, pubSubTopicName, cfg.PubSub.TopicName)
	require.Equal(t, sentryUrl, cfg.Sentry.Url)
	require.Equal(t, sentryLabels, cfg.Sentry.Labels)
}

func TestEnvConfigs(t *testing.T) {
	host := "envhost"
	port := "9999"
	pubSubProject := "test-project-1234"
	pubSubTopicName := "test-topic-1234"
	sentryUrl := "https://zz.zz.zz"
	sentryLabels := map[string]string{}

	os.Setenv("HOST", host)
	os.Setenv("PORT", port)
	os.Setenv("PUBSUB::PROJECT", pubSubProject)
	os.Setenv("PUBSUB::TOPICNAME", pubSubTopicName)
	os.Setenv("SENTRY::URL", sentryUrl)

	cfg := tu.Config{}
	var filePaths []string

	err := ParseConfig(&cfg, filePaths)
	require.NoError(t, err)
	require.Equal(t, host, cfg.Host)
	require.Equal(t, port, fmt.Sprint(cfg.Port))
	require.Equal(t, pubSubProject, cfg.PubSub.Project)
	require.Equal(t, pubSubTopicName, cfg.PubSub.TopicName)
	require.Equal(t, sentryUrl, cfg.Sentry.Url)
	require.Equal(t, sentryLabels, cfg.Sentry.Labels)
}
