package proof

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Proof is the core domain model for proof-service.
// It is an immutable snapshot generated when a booking is confirmed.
// It acts as the legal "bon de transport" required by VTC regulation.
type Proof struct {
	ID       string
	BookedAt time.Time

	// Operator snapshot — captured at confirmation time.
	// Remains valid even if company details change later.
	OperatorName       string
	OperatorSiret      string
	OperatorEvtcNumber string

	// Vehicle snapshot — license plate at the time of booking.
	VehicleNumberPlate string

	// Client snapshot — captured at confirmation time.
	ClientFirstName string
	ClientLastName  string
	ClientPhone     *string

	// Ride snapshot
	PickupAddress  string
	DropoffAddress string
	ScheduledAt    time.Time

	// SHA-256 hash of the legally significant fields.
	// Computed after creation and stored back via UpdateProofHash.
	// A mismatch indicates tampering.
	ProofHash *string

	// Reference to the booking — ID only, no FK across service boundary.
	BookingID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateProofInput carries the data needed to create a new proof.
// Populated from the booking.confirmed event + operator config.
type CreateProofInput struct {
	BookingID string
	BookedAt  time.Time

	// From operator config (environment variables)
	OperatorName       string
	OperatorSiret      string
	OperatorEvtcNumber string

	// From booking.confirmed event
	VehicleNumberPlate string
	ClientFirstName    string
	ClientLastName     string
	ClientPhone        *string
	PickupAddress      string
	DropoffAddress     string
	ScheduledAt        time.Time
}

// Validate checks that the input is complete.
func (i CreateProofInput) Validate() error {
	if i.BookingID == "" {
		return ErrMissingBookingID
	}
	if i.OperatorName == "" {
		return ErrMissingOperatorName
	}
	if i.OperatorSiret == "" {
		return ErrMissingOperatorSiret
	}
	if i.OperatorEvtcNumber == "" {
		return ErrMissingOperatorEvtcNumber
	}
	if i.VehicleNumberPlate == "" {
		return ErrMissingVehicleNumberPlate
	}
	if i.ClientFirstName == "" {
		return ErrMissingClientFirstName
	}
	if i.ClientLastName == "" {
		return ErrMissingClientLastName
	}
	if i.PickupAddress == "" {
		return ErrMissingPickupAddress
	}
	if i.DropoffAddress == "" {
		return ErrMissingDropoffAddress
	}
	if i.ScheduledAt.IsZero() {
		return ErrMissingScheduledAt
	}
	return nil
}

// OperatorConfig holds the operator information read from environment variables.
// Injected into the service at startup.
type OperatorConfig struct {
	Name       string
	Siret      string
	EvtcNumber string
}

// Validate checks that all operator fields are set.
func (o OperatorConfig) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("OPERATOR_NAME environment variable is required")
	}
	if o.Siret == "" {
		return fmt.Errorf("OPERATOR_SIRET environment variable is required")
	}
	if o.EvtcNumber == "" {
		return fmt.Errorf("OPERATOR_EVTC_NUMBER environment variable is required")
	}
	return nil
}

// =============================================================================
// Hashing
// =============================================================================

// hashable defines the legally significant fields used to compute the proof hash.
// Field order is fixed — changing it would invalidate existing hashes.
// ID, ProofHash, CreatedAt, UpdatedAt are intentionally excluded.
type hashable struct {
	BookingID          string  `json:"booking_id"`
	BookedAt           string  `json:"booked_at"` // RFC3339 for determinism
	OperatorName       string  `json:"operator_name"`
	OperatorSiret      string  `json:"operator_siret"`
	OperatorEvtcNumber string  `json:"operator_evtc_number"`
	VehicleNumberPlate string  `json:"vehicle_number_plate"`
	ClientFirstName    string  `json:"client_first_name"`
	ClientLastName     string  `json:"client_last_name"`
	ClientPhone        *string `json:"client_phone"`
	PickupAddress      string  `json:"pickup_address"`
	DropoffAddress     string  `json:"dropoff_address"`
	ScheduledAt        string  `json:"scheduled_at"` // RFC3339 for determinism
}

// ComputeHash returns the SHA-256 hex hash of the proof's legally significant fields.
func (p *Proof) ComputeHash() (string, error) {
	h := hashable{
		BookingID:          p.BookingID,
		BookedAt:           p.BookedAt.UTC().Format(time.RFC3339),
		OperatorName:       p.OperatorName,
		OperatorSiret:      p.OperatorSiret,
		OperatorEvtcNumber: p.OperatorEvtcNumber,
		VehicleNumberPlate: p.VehicleNumberPlate,
		ClientFirstName:    p.ClientFirstName,
		ClientLastName:     p.ClientLastName,
		ClientPhone:        p.ClientPhone,
		PickupAddress:      p.PickupAddress,
		DropoffAddress:     p.DropoffAddress,
		ScheduledAt:        p.ScheduledAt.UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("proof.ComputeHash: marshal: %w", err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// =============================================================================
// Domain errors
// =============================================================================

type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrProofNotFound             domainError = "proof not found"
	ErrProofAlreadyExists        domainError = "proof already exists for this booking"
	ErrMissingBookingID          domainError = "booking_id is required"
	ErrMissingOperatorName       domainError = "operator_name is required"
	ErrMissingOperatorSiret      domainError = "operator_siret is required"
	ErrMissingOperatorEvtcNumber domainError = "operator_evtc_number is required"
	ErrMissingVehicleNumberPlate domainError = "vehicle_number_plate is required"
	ErrMissingClientFirstName    domainError = "client_first_name is required"
	ErrMissingClientLastName     domainError = "client_last_name is required"
	ErrMissingPickupAddress      domainError = "pickup_address is required"
	ErrMissingDropoffAddress     domainError = "dropoff_address is required"
	ErrMissingScheduledAt        domainError = "scheduled_at is required"
)
