package proof

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statictoad/vtc-platform/internal/proof/db"
)

// Repository defines the data access contract for proofs.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Proof, error)
	FindByBookingID(ctx context.Context, bookingID string) (*Proof, error)
	Create(ctx context.Context, p *Proof) error
	UpdateHash(ctx context.Context, id string, hash string, updatedAt time.Time) error
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

func toText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func fromText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// =============================================================================
// Repository methods
// =============================================================================

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*Proof, error) {
	row, err := r.q.GetProofByID(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProofNotFound
		}
		return nil, err
	}
	return toProof(row), nil
}

func (r *postgresRepository) FindByBookingID(ctx context.Context, bookingID string) (*Proof, error) {
	row, err := r.q.GetProofByBookingID(ctx, toUUID(bookingID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProofNotFound
		}
		return nil, err
	}
	return toProof(row), nil
}

func (r *postgresRepository) Create(ctx context.Context, p *Proof) error {
	err := r.q.CreateProof(ctx, db.CreateProofParams{
		ID:                 toUUID(p.ID),
		BookedAt:           toTimestamptz(p.BookedAt),
		OperatorName:       p.OperatorName,
		OperatorSiret:      p.OperatorSiret,
		OperatorEvtcNumber: p.OperatorEvtcNumber,
		VehicleNumberPlate: p.VehicleNumberPlate,
		ClientFirstName:    p.ClientFirstName,
		ClientLastName:     p.ClientLastName,
		ClientPhone:        toText(p.ClientPhone),
		PickupAddress:      p.PickupAddress,
		DropoffAddress:     p.DropoffAddress,
		ScheduledAt:        toTimestamptz(p.ScheduledAt),
		ProofHash:          toText(p.ProofHash),
		BookingID:          toUUID(p.BookingID),
		CreatedAt:          toTimestamptz(p.CreatedAt),
		UpdatedAt:          toTimestamptz(p.UpdatedAt),
	})
	if err != nil {
		if isDuplicateError(err) {
			return ErrProofAlreadyExists
		}
		return err
	}
	return nil
}

func (r *postgresRepository) UpdateHash(ctx context.Context, id string, hash string, updatedAt time.Time) error {
	rows, err := r.q.UpdateProofHash(ctx, db.UpdateProofHashParams{
		ID:        toUUID(id),
		ProofHash: toText(&hash),
		UpdatedAt: toTimestamptz(updatedAt),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProofNotFound
	}
	return nil
}

// toProof maps a generated db.ProofProof to the domain Proof.
func toProof(row db.ProofProof) *Proof {
	return &Proof{
		ID:                 fromUUID(row.ID),
		BookedAt:           row.BookedAt.Time,
		OperatorName:       row.OperatorName,
		OperatorSiret:      row.OperatorSiret,
		OperatorEvtcNumber: row.OperatorEvtcNumber,
		VehicleNumberPlate: row.VehicleNumberPlate,
		ClientFirstName:    row.ClientFirstName,
		ClientLastName:     row.ClientLastName,
		ClientPhone:        fromText(row.ClientPhone),
		PickupAddress:      row.PickupAddress,
		DropoffAddress:     row.DropoffAddress,
		ScheduledAt:        row.ScheduledAt.Time,
		ProofHash:          fromText(row.ProofHash),
		BookingID:          fromUUID(row.BookingID),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}

// isDuplicateError checks if the error is a Postgres unique constraint violation.
func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
