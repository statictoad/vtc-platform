package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/statictoad/vtc-platform/pkg/events"
)

// Service holds the business logic for bookings.
// It depends on abstractions — Repository and events.Publisher —
// never on concrete implementations.
type Service struct {
	repo      Repository
	publisher events.Publisher
}

// NewService returns a new booking Service.
func NewService(repo Repository, publisher events.Publisher) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
	}
}

// GetBooking returns a booking by ID.
func (s *Service) GetBooking(ctx context.Context, id string) (*Booking, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ListByClient returns all bookings for a given client, ordered by scheduled date.
func (s *Service) ListByClient(ctx context.Context, clientID string) ([]Booking, error) {
	bookings, err := s.repo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("booking.ListByClient: %w", err)
	}
	return bookings, nil
}

// Create creates a new booking in PENDING status.
// The caller is responsible for validating the input before calling this method.
func (s *Service) Create(ctx context.Context, input CreateBookingInput) (*Booking, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("booking.Create: generate id: %w", err)
	}

	now := time.Now()
	b := &Booking{
		ID:                 id.String(),
		ClientID:           input.ClientID,
		VehicleID:          input.VehicleID,
		ScheduledAt:        input.ScheduledAt,
		PickupAddress:      input.PickupAddress,
		DropoffAddress:     input.DropoffAddress,
		Status:             StatusPending,
		TotalAmount:        input.TotalAmount,
		TaxRateBasisPoints: nil,
		EstimatedDistance:  input.EstimatedDistance,
		Co2Grams:           input.Co2Grams,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, fmt.Errorf("booking.Create: %w", err)
	}

	return b, nil
}

// Confirm transitions a booking from PENDING to CONFIRMED.
// Publishes a BookingConfirmed event consumed by proof-service,
// notification-service, and billing-service.
func (s *Service) Confirm(ctx context.Context, id string) error {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if b.Status != StatusPending {
		return ErrInvalidTransition
	}

	now := time.Now()
	if err := s.repo.UpdateStatus(ctx, id, StatusConfirmed, now); err != nil {
		return fmt.Errorf("booking.Confirm: %w", err)
	}

	eventID, _ := uuid.NewV7()
	evt := events.BookingConfirmed{
		EventID:        eventID.String(),
		OccurredAt:     now,
		BookingID:      b.ID,
		ClientID:       b.ClientID,
		VehicleID:      b.VehicleID,
		ScheduledAt:    b.ScheduledAt,
		PickupAddress:  b.PickupAddress,
		DropoffAddress: b.DropoffAddress,
	}

	if err := s.publisher.Publish(ctx, events.StreamBookingConfirmed, evt); err != nil {
		// Log but do not fail — the booking is confirmed in the DB.
		// A background reconciliation job can re-publish missed events.
		// TODO: implement outbox pattern for guaranteed delivery.
		fmt.Printf("warning: failed to publish BookingConfirmed event: %v\n", err)
	}

	return nil
}

// Cancel transitions a booking to CANCELLED.
// Only PENDING or CONFIRMED bookings can be cancelled.
func (s *Service) Cancel(ctx context.Context, id string) error {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if b.Status == StatusCompleted || b.Status == StatusCancelled {
		return ErrInvalidTransition
	}

	now := time.Now()
	if err := s.repo.UpdateStatus(ctx, id, StatusCancelled, now); err != nil {
		return fmt.Errorf("booking.Cancel: %w", err)
	}

	eventID, _ := uuid.NewV7()
	evt := events.BookingCancelled{
		EventID:    eventID.String(),
		OccurredAt: now,
		BookingID:  id,
		ClientID:   b.ClientID,
	}

	if err := s.publisher.Publish(ctx, events.StreamBookingCancelled, evt); err != nil {
		fmt.Printf("warning: failed to publish BookingCancelled event: %v\n", err)
	}

	return nil
}

// Complete transitions a booking from CONFIRMED to COMPLETED.
func (s *Service) Complete(ctx context.Context, id string) error {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if b.Status != StatusConfirmed {
		return ErrInvalidTransition
	}

	now := time.Now()
	if err := s.repo.UpdateStatus(ctx, id, StatusCompleted, now); err != nil {
		return fmt.Errorf("booking.Complete: %w", err)
	}

	eventID, _ := uuid.NewV7()
	evt := events.BookingCompleted{
		EventID:    eventID.String(),
		OccurredAt: now,
		BookingID:  id,
		ClientID:   b.ClientID,
	}

	if err := s.publisher.Publish(ctx, events.StreamBookingCompleted, evt); err != nil {
		fmt.Printf("warning: failed to publish BookingCompleted event: %v\n", err)
	}

	return nil
}
