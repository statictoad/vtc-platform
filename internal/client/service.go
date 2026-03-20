package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/statictoad/vtc-platform/pkg/events"
)

// Service holds the business logic for clients.
type Service struct {
	repo      Repository
	publisher events.Publisher
}

// NewService returns a new client Service.
func NewService(repo Repository, publisher events.Publisher) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
	}
}

// GetByID returns a client by internal ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Client, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetByClerkUserID returns a client by Clerk user ID.
// Used by the gateway to resolve the current user from a Clerk JWT.
func (s *Service) GetByClerkUserID(ctx context.Context, clerkUserID string) (*Client, error) {
	c, err := s.repo.FindByClerkUserID(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// HandleWebhook processes an incoming Clerk webhook payload.
// Supports "user.created" and "user.deleted" event types.
// "user.updated" is intentionally ignored — email/name changes in Clerk
// are not synced back here to avoid overwriting operator-set data.
func (s *Service) HandleWebhook(ctx context.Context, payload WebhookClientCreated) error {
	switch payload.Type {
	case "user.created":
		return s.create(ctx, payload.ToCreateInput())
	case "user.deleted":
		// Soft delete is handled at the booking level via events.
		// client-service does not hard delete clients — historical bookings
		// must remain readable.
		return nil
	default:
		return nil
	}
}

// create creates a new client from a Clerk webhook payload.
// Not exported — callers should use HandleWebhook.
func (s *Service) create(ctx context.Context, input CreateClientInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("client.create: generate id: %w", err)
	}

	now := time.Now()
	c := &Client{
		ID:          id.String(),
		ClerkUserID: input.ClerkUserID,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Email:       input.Email,
		Phone:       nil,
		Notes:       nil,
		CanPayLater: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return fmt.Errorf("client.create: %w", err)
	}

	eventID, _ := uuid.NewV7()
	evt := events.ClientCreated{
		EventID:    eventID.String(),
		OccurredAt: now,
		ClientID:   c.ID,
		FirstName:  c.FirstName,
		LastName:   c.LastName,
		Email:      c.Email,
	}

	if err := s.publisher.Publish(ctx, events.StreamClientCreated, evt); err != nil {
		// Non-fatal — log and continue.
		// TODO: implement outbox pattern for guaranteed delivery.
		fmt.Printf("warning: failed to publish ClientCreated event: %v\n", err)
	}

	return nil
}

// Update updates a client's profile fields.
// Only Phone, Notes, and CanPayLater can be updated.
// ClerkUserID and Email are immutable — owned by Clerk.
func (s *Service) Update(ctx context.Context, id string, input UpdateClientInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	// Verify the client exists before updating.
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, id, input, time.Now()); err != nil {
		return fmt.Errorf("client.Update: %w", err)
	}

	return nil
}
