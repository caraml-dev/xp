package messagequeue

import "context"

// noopMQService is the struct for no operation to event updates
type noopMQ struct{}

// NewNoopMQService initializes a noopMQ struct
func NewNoopMQService() (MessageQueueService, error) {
	return &noopMQ{}, nil
}

func (k *noopMQ) SubscribeToManagementService(ctx context.Context) error {
	return nil
}

func (k *noopMQ) DeleteSubscriptions(ctx context.Context) error {
	return nil
}
