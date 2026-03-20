package booking

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/statictoad/vtc-platform/internal/booking/db"
)

// Repository defines the data access contract for bookings.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Booking, error)
	FindByClientID(ctx context.Context, clientID string) ([]Booking, error)
	Create(ctx context.Context, b *Booking) error
	UpdateStatus(ctx context.Context, id string, status BookingStatus, updatedAt time.Time) error
	UpdateInvoice(ctx context.Context, id string, externalInvoiceID string, invoiceURL string) error
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

func toInt4(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

func fromInt4(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int32)
	return &v
}

func toFloat8(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *f, Valid: true}
}

func fromFloat8(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// =============================================================================
// Repository methods
// =============================================================================

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*Booking, error) {
	row, err := r.q.GetBookingByID(ctx, toUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	return toBooking(row), nil
}

func (r *postgresRepository) FindByClientID(ctx context.Context, clientID string) ([]Booking, error) {
	rows, err := r.q.ListBookingsByClientID(ctx, toUUID(clientID))
	if err != nil {
		return nil, err
	}
	bookings := make([]Booking, len(rows))
	for i, row := range rows {
		bookings[i] = *toBooking(row)
	}
	return bookings, nil
}

func (r *postgresRepository) Create(ctx context.Context, b *Booking) error {
	return r.q.CreateBooking(ctx, db.CreateBookingParams{
		ID:                 toUUID(b.ID),
		ClientID:           toUUID(b.ClientID),
		VehicleID:          toUUID(b.VehicleID),
		ScheduledAt:        toTimestamptz(b.ScheduledAt),
		PickupAddress:      b.PickupAddress,
		DropoffAddress:     b.DropoffAddress,
		Status:             db.BookingBookingStatus(b.Status),
		TotalAmount:        toInt4(b.TotalAmount),
		TaxRateBasisPoints: toInt4(b.TaxRateBasisPoints),
		EstimatedKm:        toFloat8(b.EstimatedKm),
		Co2Grams:           toInt4(b.Co2Grams),
		CreatedAt:          toTimestamptz(b.CreatedAt),
		UpdatedAt:          toTimestamptz(b.UpdatedAt),
	})
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id string, status BookingStatus, updatedAt time.Time) error {
	rows, err := r.q.UpdateBookingStatus(ctx, db.UpdateBookingStatusParams{
		ID:        toUUID(id),
		Status:    db.BookingBookingStatus(status),
		UpdatedAt: toTimestamptz(updatedAt),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *postgresRepository) UpdateInvoice(ctx context.Context, id string, externalInvoiceID string, invoiceURL string) error {
	rows, err := r.q.UpdateBookingInvoice(ctx, db.UpdateBookingInvoiceParams{
		ID:                toUUID(id),
		ExternalInvoiceID: toText(&externalInvoiceID),
		InvoiceUrl:        toText(&invoiceURL),
		UpdatedAt:         toTimestamptz(time.Now()),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBookingNotFound
	}
	return nil
}

// toBooking maps a generated db.BookingBooking to the domain Booking.
func toBooking(row db.BookingBooking) *Booking {
	return &Booking{
		ID:                 fromUUID(row.ID),
		ClientID:           fromUUID(row.ClientID),
		VehicleID:          fromUUID(row.VehicleID),
		ScheduledAt:        row.ScheduledAt.Time,
		PickupAddress:      row.PickupAddress,
		DropoffAddress:     row.DropoffAddress,
		Status:             BookingStatus(row.Status),
		TotalAmount:        fromInt4(row.TotalAmount),
		TaxRateBasisPoints: fromInt4(row.TaxRateBasisPoints),
		EstimatedKm:        fromFloat8(row.EstimatedKm),
		Co2Grams:           fromInt4(row.Co2Grams),
		ExternalInvoiceID:  fromText(row.ExternalInvoiceID),
		InvoiceURL:         fromText(row.InvoiceUrl),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}
