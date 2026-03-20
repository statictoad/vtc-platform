package events

import (
	"context"
	"time"
)

// Stream names — used as Valkey stream keys.
const (
	StreamBookingConfirmed = "booking.confirmed"
	StreamBookingCancelled = "booking.cancelled"
	StreamBookingCompleted = "booking.completed"
	StreamClientCreated    = "client.created"
)

// Publisher is the contract for publishing domain events.
// The concrete implementation uses Valkey Streams.
type Publisher interface {
	Publish(ctx context.Context, stream string, payload any) error
}

// Consumer is the contract for consuming domain events.
type Consumer interface {
	Consume(ctx context.Context, stream string, handler func(ctx context.Context, payload []byte) error) error
}

// =============================================================================
// Booking events
// =============================================================================

// BookingConfirmed is published when a booking transitions to CONFIRMED.
// Consumed by: proof-service, notification-service, billing-service.
type BookingConfirmed struct {
	EventID        string    `json:"event_id"`
	OccurredAt     time.Time `json:"occurred_at"`
	BookingID      string    `json:"booking_id"`
	ClientID       string    `json:"client_id"`
	VehicleID      string    `json:"vehicle_id"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	PickupAddress  string    `json:"pickup_address"`
	DropoffAddress string    `json:"dropoff_address"`
}

// BookingCancelled is published when a booking is cancelled.
// Consumed by: notification-service.
type BookingCancelled struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	BookingID  string    `json:"booking_id"`
	ClientID   string    `json:"client_id"`
}

// BookingCompleted is published when a ride is completed.
// Consumed by: billing-service.
type BookingCompleted struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	BookingID  string    `json:"booking_id"`
	ClientID   string    `json:"client_id"`
}

// =============================================================================
// Client events
// =============================================================================

// ClientCreated is published when a new client registers.
// Consumed by: notification-service.
type ClientCreated struct {
	EventID    string    `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
	ClientID   string    `json:"client_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
}
