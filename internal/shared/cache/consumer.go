package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Consumer reads domain events from a Valkey Stream using consumer groups.
// Implements pkg/events.Consumer.
//
// Each service that consumes events should use a unique group name so that
// events are delivered to each service independently.
type Consumer struct {
	client   valkey.Client
	group    string // consumer group name — unique per service (e.g. "proof-service")
	consumer string // consumer name — unique per instance (e.g. "proof-service-1")
}

// NewConsumer returns a new Consumer.
//
//   - group: name of the consumer group (e.g. "proof-service")
//   - consumer: name of this consumer instance (e.g. hostname or pod name)
func NewConsumer(client valkey.Client, group, consumer string) *Consumer {
	return &Consumer{
		client:   client,
		group:    group,
		consumer: consumer,
	}
}

// EnsureGroup creates the consumer group for the given stream if it does not
// already exist. Must be called before Consume.
// Uses "$" as the start ID so only new messages are delivered after creation.
func (c *Consumer) EnsureGroup(ctx context.Context, stream string) error {
	cmd := c.client.B().XgroupCreate().
		Key(stream).
		Group(c.group).
		Id("$").
		Mkstream().
		Build()

	err := c.client.Do(ctx, cmd).Error()
	if err != nil && !isGroupExistsError(err) {
		return fmt.Errorf("cache.Consumer: XGROUP CREATE %s/%s: %w", stream, c.group, err)
	}

	return nil
}

// Consume reads messages from the stream in a blocking loop.
// For each message, it calls handler with the raw JSON payload bytes.
// If handler returns nil, the message is acknowledged (XACK).
// If handler returns an error, the message remains in the pending list
// and will be redelivered on the next restart.
//
// Consume blocks until the context is cancelled.
func (c *Consumer) Consume(ctx context.Context, stream string, handler func(ctx context.Context, payload []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		cmd := c.client.B().Xreadgroup().
			Group(c.group, c.consumer).
			Count(10).
			Block(2000). // block for 2 seconds, then loop
			Streams().
			Key(stream).
			Id(">"). // only undelivered messages
			Build()

		results, err := c.client.Do(ctx, cmd).AsXRead()
		if err != nil {
			if valkey.IsValkeyNil(err) {
				// timeout — no new messages, loop again
				continue
			}
			slog.Error("cache.Consumer: XREADGROUP failed",
				"stream", stream,
				"error", err,
			)
			time.Sleep(time.Second) // back off before retrying
			continue
		}

		for _, messages := range results {
			for _, msg := range messages {
				payload, ok := msg.FieldValues["payload"]
				if !ok {
					slog.Warn("cache.Consumer: message missing payload field",
						"stream", stream,
						"id", msg.ID,
					)
					c.ack(ctx, stream, msg.ID)
					continue
				}

				if err := handler(ctx, []byte(payload)); err != nil {
					slog.Error("cache.Consumer: handler failed",
						"stream", stream,
						"id", msg.ID,
						"error", err,
					)
					// Do not ACK — message stays in pending list for redelivery.
					continue
				}

				c.ack(ctx, stream, msg.ID)
			}
		}
	}
}

// ack acknowledges a processed message.
func (c *Consumer) ack(ctx context.Context, stream, id string) {
	cmd := c.client.B().Xack().
		Key(stream).
		Group(c.group).
		Id(id).
		Build()

	if err := c.client.Do(ctx, cmd).Error(); err != nil {
		slog.Error("cache.Consumer: XACK failed",
			"stream", stream,
			"id", id,
			"error", err,
		)
	}
}

// isGroupExistsError checks if the error is a BUSYGROUP error —
// returned by Valkey when the consumer group already exists.
func isGroupExistsError(err error) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) > 9 && err.Error()[:9] == "BUSYGROUP"
}
