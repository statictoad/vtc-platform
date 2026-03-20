package proof

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/statictoad/vtc-platform/internal/proof/upstream"
	"github.com/statictoad/vtc-platform/pkg/events"
)

// Service holds the business logic for proof-service.
type Service struct {
	repo      Repository
	operator  OperatorConfig
	clientSvc *upstream.ClientServiceClient
	fleetSvc  *upstream.FleetServiceClient
}

// NewService returns a new proof Service.
// operator is validated at construction time — the service refuses to start
// if any operator field is missing.
func NewService(repo Repository, operator OperatorConfig, clientSvc *upstream.ClientServiceClient, fleetSvc *upstream.FleetServiceClient) (*Service, error) {
	if err := operator.Validate(); err != nil {
		return nil, err
	}
	return &Service{
		repo:      repo,
		operator:  operator,
		clientSvc: clientSvc,
		fleetSvc:  fleetSvc,
	}, nil
}

// GetByBookingID returns the proof for a given booking.
// Used by the HTTP handler for GET /proofs/:bookingId.
func (s *Service) GetByBookingID(ctx context.Context, bookingID string) (*Proof, error) {
	p, err := s.repo.FindByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// HandleBookingConfirmed processes a booking.confirmed event.
// Creates an immutable proof snapshot and computes its SHA-256 hash.
// Idempotent — if a proof already exists for this booking, it is a no-op.
func (s *Service) HandleBookingConfirmed(ctx context.Context, payload []byte) error {
	var evt events.BookingConfirmed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("proof.HandleBookingConfirmed: unmarshal event: %w", err)
	}

	// Idempotency check — proof may already exist if the event was redelivered.
	existing, err := s.repo.FindByBookingID(ctx, evt.BookingID)
	if err != nil && err != ErrProofNotFound {
		return fmt.Errorf("proof.HandleBookingConfirmed: check existing: %w", err)
	}
	if existing != nil {
		slog.Info("proof already exists, skipping",
			"booking_id", evt.BookingID,
		)
		return nil
	}

	// Fetch client details from client-service
	client, err := s.clientSvc.GetClient(ctx, evt.ClientID)
	if err != nil {
		return fmt.Errorf("proof.HandleBookingConfirmed: fetch client: %w", err)
	}

	// Fetch vehicle details from fleet-service
	vehicle, err := s.fleetSvc.GetVehicle(ctx, evt.VehicleID)
	if err != nil {
		return fmt.Errorf("proof.HandleBookingConfirmed: fetch vehicle: %w", err)
	}

	input := CreateProofInput{
		BookingID:          evt.BookingID,
		BookedAt:           evt.OccurredAt,
		OperatorName:       s.operator.Name,
		OperatorSiret:      s.operator.Siret,
		OperatorEvtcNumber: s.operator.EvtcNumber,
		VehicleNumberPlate: vehicle.NumberPlate,
		ClientFirstName:    client.FirstName,
		ClientLastName:     client.LastName,
		ClientPhone:        client.Phone,
		PickupAddress:      evt.PickupAddress,
		DropoffAddress:     evt.DropoffAddress,
		ScheduledAt:        evt.ScheduledAt,
	}

	return s.create(ctx, input)
}

// create creates a new proof and computes its hash.
func (s *Service) create(ctx context.Context, input CreateProofInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("proof.create: generate id: %w", err)
	}

	now := time.Now()
	p := &Proof{
		ID:                 id.String(),
		BookedAt:           input.BookedAt,
		OperatorName:       input.OperatorName,
		OperatorSiret:      input.OperatorSiret,
		OperatorEvtcNumber: input.OperatorEvtcNumber,
		VehicleNumberPlate: input.VehicleNumberPlate,
		ClientFirstName:    input.ClientFirstName,
		ClientLastName:     input.ClientLastName,
		ClientPhone:        input.ClientPhone,
		PickupAddress:      input.PickupAddress,
		DropoffAddress:     input.DropoffAddress,
		ScheduledAt:        input.ScheduledAt,
		ProofHash:          nil, // computed after insert
		BookingID:          input.BookingID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		if err == ErrProofAlreadyExists {
			// Race condition — another instance created the proof concurrently.
			// Safe to ignore.
			return nil
		}
		return fmt.Errorf("proof.create: %w", err)
	}

	// Compute and store the hash after the proof is persisted.
	hash, err := p.ComputeHash()
	if err != nil {
		return fmt.Errorf("proof.create: compute hash: %w", err)
	}

	if err := s.repo.UpdateHash(ctx, p.ID, hash, time.Now()); err != nil {
		// Non-fatal — proof exists, hash can be recomputed later.
		// TODO: add a background job to fix proofs with missing hashes.
		slog.Error("proof.create: failed to store hash",
			"proof_id", p.ID,
			"error", err,
		)
	}

	slog.Info("proof created",
		"proof_id", p.ID,
		"booking_id", p.BookingID,
	)

	return nil
}
