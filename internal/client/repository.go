package client

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statictoad/vtc-platform/internal/client/db"
)

// Repository defines the data access contract for clients.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Client, error)
	FindByClerkUserID(ctx context.Context, clerkUserID string) (*Client, error)
	Create(ctx context.Context, c *Client) error
	Update(ctx context.Context, id string, input UpdateClientInput, updatedAt time.Time) error
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

func toBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
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

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*Client, error) {
	row, err := r.q.GetClientByID(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	return toClient(row), nil
}

func (r *postgresRepository) FindByClerkUserID(ctx context.Context, clerkUserID string) (*Client, error) {
	row, err := r.q.GetClientByClerkUserID(ctx, clerkUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	return toClient(row), nil
}

func (r *postgresRepository) Create(ctx context.Context, c *Client) error {
	err := r.q.CreateClient(ctx, db.CreateClientParams{
		ID:          toUUID(c.ID),
		ClerkUserID: c.ClerkUserID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		Email:       c.Email,
		Phone:       toText(c.Phone),
		Notes:       toText(c.Notes),
		CanPayLater: c.CanPayLater,
		CreatedAt:   toTimestamptz(c.CreatedAt),
		UpdatedAt:   toTimestamptz(c.UpdatedAt),
	})
	if err != nil {
		if isDuplicateError(err) {
			return ErrDuplicateClerkUserID
		}
		return err
	}
	return nil
}

func (r *postgresRepository) Update(ctx context.Context, id string, input UpdateClientInput, updatedAt time.Time) error {
	rows, err := r.q.UpdateClient(ctx, db.UpdateClientParams{
		ID:          toUUID(id),
		Phone:       toText(input.Phone),
		Notes:       toText(input.Notes),
		CanPayLater: toBool(input.CanPayLater),
		UpdatedAt:   toTimestamptz(updatedAt),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrClientNotFound
	}
	return nil
}

// toClient maps a generated db.ClientClient to the domain Client.
func toClient(row db.ClientClient) *Client {
	return &Client{
		ID:          fromUUID(row.ID),
		ClerkUserID: row.ClerkUserID,
		FirstName:   row.FirstName,
		LastName:    row.LastName,
		Email:       row.Email,
		Phone:       fromText(row.Phone),
		Notes:       fromText(row.Notes),
		CanPayLater: row.CanPayLater,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

// isDuplicateError checks if the error is a Postgres unique constraint violation.
func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
