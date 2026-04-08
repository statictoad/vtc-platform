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
	PickupStreet       string
	PickupCity         string
	PickupCountry      string
	PickupDetails      *string
	PickupLat          *float64 // from GEOGRAPHY(POINT, 4326)
	PickupLng          *float64
	DropoffStreet      string
	DropoffCity        string
	DropoffCountry     string
	DropoffDetails     *string
	DropoffLat         *float64
	DropoffLng         *float64
	Passengers         int
	Suitcases          int
	Notes              *string
	Status             BookingStatus
	TotalAmount        *int    // optional — not known at booking creation time
	TaxRateBasisPoints *int    // optional — set later
	EstimatedDistance  *int    // optional — in meters
	Co2Grams           *int    // optional — regulatory display
	ExternalInvoiceID  *string // set later by billing-service
	InvoiceURL         *string // set later by billing-service
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Client is what booking-service needs to know about a client.
// This is NOT imported from client-service — it is a local read model.
// Populated by calling client-service via HTTP before creating a booking.
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
	ClientID          string
	VehicleID         string
	ScheduledAt       time.Time
	PickupStreet      string
	PickupCity        string
	PickupCountry     string
	PickupDetails     *string
	PickupLat         *float64
	PickupLng         *float64
	DropoffStreet     string
	DropoffCity       string
	DropoffCountry    string
	DropoffDetails    *string
	DropoffLat        *float64
	DropoffLng        *float64
	Passengers        int
	Suitcases         int
	Notes             *string
	TotalAmount       *int
	EstimatedDistance *int
	Co2Grams          *int
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
	if i.PickupStreet == "" {
		return ErrMissingPickupStreet
	}
	if i.PickupCity == "" {
		return ErrMissingPickupCity
	}
	if i.PickupCountry == "" {
		return ErrMissingPickupCountry
	}
	if i.DropoffStreet == "" {
		return ErrMissingDropoffStreet
	}
	if i.DropoffCity == "" {
		return ErrMissingDropoffCity
	}
	if i.DropoffCountry == "" {
		return ErrMissingDropoffCountry
	}
	if i.Passengers < 1 || i.Passengers > 6 {
		return ErrInvalidPassengers
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
	ErrMissingPickupStreet   domainError = "pickup_street is required"
	ErrMissingPickupCity     domainError = "pickup_city is required"
	ErrMissingPickupCountry  domainError = "pickup_country is required"
	ErrMissingDropoffStreet  domainError = "dropoff_street is required"
	ErrMissingDropoffCity    domainError = "dropoff_city is required"
	ErrMissingDropoffCountry domainError = "dropoff_country is required"
	ErrInvalidPassengers     domainError = "passengers must be between 1 and 6"
	ErrBookingNotFound       domainError = "booking not found"
	ErrInvalidTransition     domainError = "invalid status transition"
)
