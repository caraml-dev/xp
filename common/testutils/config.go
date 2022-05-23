package testutils

type Config struct {
	Host   string `default:"localhost"`
	Port   int    `default:"3000"`
	PubSub PubSub
	Sentry Sentry
}

type PubSub struct {
	Project   string `default:"dev"`
	TopicName string `default:"update"`
}

type Sentry struct {
	Url    string `default:"https://xx.xx.xx"`
	Labels map[string]string
}
