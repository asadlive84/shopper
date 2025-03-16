package ports

import "context"

type MessagingPort interface {
	Publish(ctx context.Context, exchange, message string) error
	Consume(queue string, consumerFunc func(context.Context, string)) error
	Close()
}