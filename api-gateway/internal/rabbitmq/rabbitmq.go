package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Logger  *zap.Logger
	Tracer  trace.Tracer
}

func NewRabbitMQClient(url string, logger *zap.Logger) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open RabbitMQ channel", zap.Error(err))
		conn.Close()
		return nil, err
	}
	logger.Info("Connected to RabbitMQ")
	return &Client{
		Conn:    conn,
		Channel: ch,
		Logger:  logger,
		Tracer:  otel.Tracer("rabbitmq"),
	}, nil
}

func (c *Client) Setup(exchange, queue string) error {
	err := c.Channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
	if err != nil {
		c.Logger.Error("Failed to declare exchange", zap.String("exchange", exchange), zap.Error(err))
		return err
	}
	q, err := c.Channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		c.Logger.Error("Failed to declare queue", zap.String("queue", queue), zap.Error(err))
		return err
	}
	err = c.Channel.QueueBind(q.Name, "", exchange, false, nil)
	if err != nil {
		c.Logger.Error("Failed to bind queue", zap.String("queue", q.Name), zap.String("exchange", exchange), zap.Error(err))
		return err
	}
	c.Logger.Info("RabbitMQ setup completed", zap.String("exchange", exchange), zap.String("queue", queue))
	return nil
}

func (c *Client) Publish(ctx context.Context, exchange, message string) error {
	ctx, span := c.Tracer.Start(ctx, "rabbitmq.publish",
		trace.WithAttributes(
			attribute.String("exchange", exchange),
			attribute.String("message", message),
		))
	defer span.End()

	err := c.Channel.PublishWithContext(ctx,
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		c.Logger.Error("Failed to publish message", zap.String("exchange", exchange), zap.String("message", message), zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	c.Logger.Info("Message published", zap.String("exchange", exchange), zap.String("message", message))
	return nil
}

func (c *Client) Consume(queue string, consumerFunc func(context.Context, string)) error {
	msgs, err := c.Channel.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		c.Logger.Error("Failed to consume messages", zap.String("queue", queue), zap.Error(err))
		return err
	}
	go func() {
		for msg := range msgs {
			ctx, span := c.Tracer.Start(context.Background(), "rabbitmq.consume",
				trace.WithAttributes(
					attribute.String("queue", queue),
					attribute.String("body", string(msg.Body)),
				))
			c.Logger.Info("Message consumed", zap.String("queue", queue), zap.String("body", string(msg.Body)))
			consumerFunc(ctx, string(msg.Body))
			span.End()
		}
	}()
	c.Logger.Info("Started consuming messages", zap.String("queue", queue))
	return nil
}

func (c *Client) PublishLogoutEvent(ctx context.Context, userID string) error {
    message := "logout:" + userID
    return c.Publish(ctx, "user_events", message)
}



// Start consuming logout events
// client.Consume("user_logout_queue", handleLogoutEvent)

func (c *Client) Close() {
	c.Channel.Close()
	c.Conn.Close()
	c.Logger.Info("RabbitMQ connection closed")
}
