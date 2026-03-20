package fleet_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/statictoad/vtc-platform/internal/fleet"
)

// fakeRepository is an in-memory implementation of Repository for testing.
type fakeRepository struct {
	vehicles map[string]*fleet.Vehicle
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{vehicles: make(map[string]*fleet.Vehicle)}
}

func (r *fakeRepository) FindByID(ctx context.Context, id string) (*fleet.Vehicle, error) {
	v, ok := r.vehicles[id]
	if !ok {
		return nil, fleet.ErrVehicleNotFound
	}
	return v, nil
}

func (r *fakeRepository) ListActive(ctx context.Context) ([]fleet.Vehicle, error) {
	var result []fleet.Vehicle
	for _, v := range r.vehicles {
		if v.Active {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (r *fakeRepository) Create(ctx context.Context, v *fleet.Vehicle) error {
	r.vehicles[v.ID] = v
	return nil
}

func (r *fakeRepository) Update(ctx context.Context, id string, input fleet.UpdateVehicleInput, updatedAt time.Time) error {
	v, ok := r.vehicles[id]
	if !ok {
		return fleet.ErrVehicleNotFound
	}
	if input.Active != nil {
		v.Active = *input.Active
	}
	v.UpdatedAt = updatedAt
	return nil
}

func TestCreate_Valid(t *testing.T) {
	repo := newFakeRepository()
	svc := fleet.NewService(repo)

	v, err := svc.Create(context.Background(), fleet.CreateVehicleInput{
		NumberPlate: "AB-123-CD",
		Brand:       "Mercedes-Benz",
		Model:       "Classe E",
		Year:        2022,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v.ID == "" {
		t.Error("expected ID to be set")
	}
	if !v.Active {
		t.Error("expected vehicle to be active on creation")
	}
}

func TestCreate_MissingBrand(t *testing.T) {
	svc := fleet.NewService(newFakeRepository())

	_, err := svc.Create(context.Background(), fleet.CreateVehicleInput{
		NumberPlate: "AB-123-CD",
		Brand:       "",
		Model:       "Classe E",
		Year:        2022,
	})

	if !errors.Is(err, fleet.ErrMissingBrand) {
		t.Errorf("expected ErrMissingBrand, got %v", err)
	}
}

func TestDeactivate_VehicleNotFound(t *testing.T) {
	svc := fleet.NewService(newFakeRepository())

	err := svc.Deactivate(context.Background(), "non-existent-id")

	if !errors.Is(err, fleet.ErrVehicleNotFound) {
		t.Errorf("expected ErrVehicleNotFound, got %v", err)
	}
}

func TestDeactivate_SetsActiveFalse(t *testing.T) {
	repo := newFakeRepository()
	svc := fleet.NewService(repo)

	v, _ := svc.Create(context.Background(), fleet.CreateVehicleInput{
		NumberPlate: "AB-123-CD",
		Brand:       "Mercedes-Benz",
		Model:       "Classe E",
		Year:        2022,
	})

	err := svc.Deactivate(context.Background(), v.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := repo.FindByID(context.Background(), v.ID)
	if updated.Active {
		t.Error("expected vehicle to be inactive after deactivate")
	}
}
