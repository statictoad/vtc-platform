package fleet

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statictoad/vtc-platform/internal/fleet/db"
)

// Repository defines the data access contract for vehicles.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Vehicle, error)
	ListActive(ctx context.Context) ([]Vehicle, error)
	Create(ctx context.Context, v *Vehicle) error
	Update(ctx context.Context, id string, input UpdateVehicleInput, updatedAt time.Time) error
}

type postgresRepository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{q: db.New(pool)}
}

// =============================================================================
// Type conversion helpers
// =============================================================================

func toUUID(s string) pgtype.UUID {
	id, _ := uuid.Parse(s)
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromUUID(u pgtype.UUID) string {
	return uuid.UUID(u.Bytes).String()
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// =============================================================================
// Repository methods
// =============================================================================

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*Vehicle, error) {
	row, err := r.q.GetVehicleByID(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, err
	}
	return toVehicle(row), nil
}

func (r *postgresRepository) ListActive(ctx context.Context) ([]Vehicle, error) {
	rows, err := r.q.ListActiveVehicles(ctx)
	if err != nil {
		return nil, err
	}
	vehicles := make([]Vehicle, len(rows))
	for i, row := range rows {
		vehicles[i] = *toVehicle(row)
	}
	return vehicles, nil
}

func (r *postgresRepository) Create(ctx context.Context, v *Vehicle) error {
	err := r.q.CreateVehicle(ctx, db.CreateVehicleParams{
		ID:          toUUID(v.ID),
		NumberPlate: v.NumberPlate,
		Brand:       v.Brand,
		Model:       v.Model,
		Year:        v.Year,
		Active:      v.Active,
		CreatedAt:   toTimestamptz(v.CreatedAt),
		UpdatedAt:   toTimestamptz(v.UpdatedAt),
	})
	if err != nil {
		if isDuplicateError(err) {
			return ErrDuplicateNumberPlate
		}
		return err
	}
	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id string, input UpdateVehicleInput, updatedAt time.Time) error {
	// UpdateVehicleParams uses non-pointer types because sqlc.narg() is not
	// used yet in the query. Fields not provided in input fall back to their
	// current DB value via COALESCE — we must fetch the current vehicle first
	// to fill in unchanged fields.
	current, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	numberPlate := current.NumberPlate
	if input.NumberPlate != nil {
		numberPlate = *input.NumberPlate
	}

	brand := current.Brand
	if input.Brand != nil {
		brand = *input.Brand
	}

	model := current.Model
	if input.Model != nil {
		model = *input.Model
	}

	year := current.Year
	if input.Year != nil {
		year = *input.Year
	}

	active := current.Active
	if input.Active != nil {
		active = *input.Active
	}

	rows, err := r.q.UpdateVehicle(ctx, db.UpdateVehicleParams{
		ID:          toUUID(id),
		NumberPlate: numberPlate,
		Brand:       brand,
		Model:       model,
		Year:        year,
		Active:      active,
		UpdatedAt:   toTimestamptz(updatedAt),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrVehicleNotFound
	}
	return nil
}

// toVehicle maps a generated db.FleetVehicle to the domain Vehicle.
func toVehicle(row db.FleetVehicle) *Vehicle {
	return &Vehicle{
		ID:          fromUUID(row.ID),
		NumberPlate: row.NumberPlate,
		Brand:       row.Brand,
		Model:       row.Model,
		Year:        row.Year,
		Active:      row.Active,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

// isDuplicateError checks if the error is a Postgres unique constraint violation.
func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
