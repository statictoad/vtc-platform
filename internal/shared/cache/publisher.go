package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/valkey-io/valkey-go"
)

// Publisher publishes domain events to Valkey Streams.
// Implements pkg/events.Publisher.
type Publisher struct {
	client valkey.Client
}

// NewPublisher returns a new Publisher backed by the given Valkey client.
func NewPublisher(client valkey.Client) *Publisher {
	return &Publisher{client: client}
}

// Publish serialises payload as JSON and appends it to the given stream
// using XADD with an auto-generated ID (*).
//
// The entry has a single field "payload" containing the JSON bytes.
// Consumers read this field and deserialise it into the expected event type.
func (p *Publisher) Publish(ctx context.Context, stream string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cache.Publisher: marshal payload: %w", err)
	}

	cmd := p.client.B().Xadd().
		Key(stream).
		Id("*").
		FieldValue().
		FieldValue("payload", string(data)).
		Build()

	if err := p.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("cache.Publisher: XADD %s: %w", stream, err)
	}

	return nil
}
