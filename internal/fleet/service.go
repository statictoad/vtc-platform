package fleet

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service holds the business logic for fleet-service.
type Service struct {
	repo Repository
}

// NewService returns a new fleet Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetByID returns a vehicle by ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Vehicle, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ListActive returns all active vehicles.
func (s *Service) ListActive(ctx context.Context) ([]Vehicle, error) {
	vehicles, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("fleet.ListActive: %w", err)
	}
	return vehicles, nil
}

// Create registers a new vehicle in the fleet.
func (s *Service) Create(ctx context.Context, input CreateVehicleInput) (*Vehicle, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("fleet.Create: generate id: %w", err)
	}

	now := time.Now()
	v := &Vehicle{
		ID:          id.String(),
		NumberPlate: input.NumberPlate,
		Brand:       input.Brand,
		Model:       input.Model,
		Year:        input.Year,
		Active:      true, // all vehicles start active
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return nil, fmt.Errorf("fleet.Create: %w", err)
	}

	return v, nil
}

// Update updates a vehicle's fields.
func (s *Service) Update(ctx context.Context, id string, input UpdateVehicleInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, id, input, time.Now()); err != nil {
		return fmt.Errorf("fleet.Update: %w", err)
	}

	return nil
}

// Deactivate marks a vehicle as inactive.
// Inactive vehicles cannot be assigned to new bookings.
func (s *Service) Deactivate(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	active := false
	if err := s.repo.Update(ctx, id, UpdateVehicleInput{Active: &active}, time.Now()); err != nil {
		return fmt.Errorf("fleet.Deactivate: %w", err)
	}

	return nil
}

// Activate marks a vehicle as active.
func (s *Service) Activate(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	active := true
	if err := s.repo.Update(ctx, id, UpdateVehicleInput{Active: &active}, time.Now()); err != nil {
		return fmt.Errorf("fleet.Activate: %w", err)
	}

	return nil
}
