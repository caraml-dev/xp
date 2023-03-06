package messagequeue

// MessageQueueKind describes the message queue for transmitting event updates to and fro Treatment Service
type MessageQueueKind = string

const (
	// NoopMQ is a No-Op Message Queue
	NoopMQ MessageQueueKind = ""
	// PubSubMQ is a PubSub Message Queue
	PubSubMQ MessageQueueKind = "pubsub"
)

type MessageQueueConfig struct {
	// The type of Message Queue for event updates
	Kind MessageQueueKind `default:""`

	// PubSubConfig captures the config related to publishing and subscribing to a PubSub Message Queue
	PubSubConfig *PubSubConfig
}

type PubSubConfig struct {
	Project   string `json:"project" default:"dev" validate:"required"`
	TopicName string `json:"topic_name" default:"xp-update" validate:"required"`
	// PubSubTimeoutSeconds is the duration beyond which subscribing to a topic will time out
	PubSubTimeoutSeconds int `json:"pub_sub_timeout_seconds" default:"30" validate:"required"`
}
