package booking

import "time"

// BookingStatus represents the lifecycle of a booking.
type BookingStatus string

const (
	StatusPending   BookingStatus = "PENDING"
	StatusConfirmed BookingStatus = "CONFIRMED"
	StatusCompleted BookingStatus = "COMPLETED"
	StatusCancelled BookingStatus = "CANCELLED"
)

// Booking is the core domain model for booking-service.
// It is owned entirely by this service — no other service imports this struct.
type Booking struct {
	ID                 string
	ClientID           string
	VehicleID          string
	ScheduledAt        time.Time
	PickupAddress      string
	DropoffAddress     string
	Status             BookingStatus
	TotalAmount        *int     // optional — not known at booking creation time
	TaxRateBasisPoints *int     // optional — set later
	EstimatedKm        *float64 // optional — may not be calculated
	Co2Grams           *int     // optional — regulatory display, may be absent
	ExternalInvoiceID  *string  // set later by billing-service
	InvoiceURL         *string  // set later by billing-service
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Client is what booking-service needs to know about a client.
// This is NOT imported from client-service — it is a local read model.
// It is populated by calling client-service via HTTP before creating a booking.
type Client struct {
	ID          string
	FirstName   string
	LastName    string
	Phone       string
	CanPayLater bool
}

// Vehicle is what booking-service needs to know about a vehicle.
// Populated by calling fleet-service via HTTP before creating a booking.
type Vehicle struct {
	ID          string
	NumberPlate string
	Brand       string
	Model       string
	Active      bool
}

// CreateBookingInput carries the data needed to create a new booking.
// Received from the HTTP handler after decoding and validating the request body.
type CreateBookingInput struct {
	ClientID       string
	VehicleID      string
	ScheduledAt    time.Time
	PickupAddress  string
	DropoffAddress string
	TotalAmount    *int
	EstimatedKm    *float64
	Co2Grams       *int
}

// Validate checks that the input is complete and coherent.
func (i CreateBookingInput) Validate() error {
	if i.ClientID == "" {
		return ErrMissingClientID
	}
	if i.VehicleID == "" {
		return ErrMissingVehicleID
	}
	if i.ScheduledAt.IsZero() {
		return ErrMissingScheduledAt
	}
	if i.ScheduledAt.Before(time.Now()) {
		return ErrScheduledAtInPast
	}
	if i.PickupAddress == "" {
		return ErrMissingPickupAddress
	}
	if i.DropoffAddress == "" {
		return ErrMissingDropoffAddress
	}
	return nil
}

// =============================================================================
// Domain errors
// =============================================================================

type domainError string

func (e domainError) Error() string {
	return string(e)
}

const (
	ErrMissingClientID       domainError = "client_id is required"
	ErrMissingVehicleID      domainError = "vehicle_id is required"
	ErrMissingScheduledAt    domainError = "scheduled_at is required"
	ErrScheduledAtInPast     domainError = "scheduled_at must be in the future"
	ErrMissingPickupAddress  domainError = "pickup_address is required"
	ErrMissingDropoffAddress domainError = "dropoff_address is required"
	ErrBookingNotFound       domainError = "booking not found"
	ErrInvalidTransition     domainError = "invalid status transition"
)
