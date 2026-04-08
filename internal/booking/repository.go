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
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{q: db.New(pool), pool: pool}
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

func toNullInt2(i *int) pgtype.Int2 {
	if i == nil {
		return pgtype.Int2{Valid: false}
	}
	return pgtype.Int2{Int16: int16(*i), Valid: true}
}

func fromNullInt2(i pgtype.Int2) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int16)
	return &v
}

// =============================================================================
// Repository methods
// =============================================================================

func (r *postgresRepository) FindByID(ctx context.Context, id string) (*Booking, error) {
	// Raw SQL to also fetch coordinates via ST_X / ST_Y
	const q = `
		SELECT id, client_id, vehicle_id,
		       scheduled_at,
		       pickup_street, pickup_city, pickup_country, pickup_details,
		       ST_X(pickup_coordinates::geometry)  AS pickup_lng,
		       ST_Y(pickup_coordinates::geometry)  AS pickup_lat,
		       dropoff_street, dropoff_city, dropoff_country, dropoff_details,
		       ST_X(dropoff_coordinates::geometry) AS dropoff_lng,
		       ST_Y(dropoff_coordinates::geometry) AS dropoff_lat,
		       passengers, suitcases, notes,
		       status, total_amount, tax_rate_basis_points,
		       estimated_distance, co2_grams,
		       external_invoice_id, invoice_url,
		       created_at, updated_at
		FROM booking.bookings
		WHERE id = $1`

	b := &Booking{}
	var pickupLng, pickupLat, dropoffLng, dropoffLat *float64

	err := r.pool.QueryRow(ctx, q, toUUID(id)).Scan(
		&b.ID, &b.ClientID, &b.VehicleID,
		&b.ScheduledAt,
		&b.PickupStreet, &b.PickupCity, &b.PickupCountry, &b.PickupDetails,
		&pickupLng, &pickupLat,
		&b.DropoffStreet, &b.DropoffCity, &b.DropoffCountry, &b.DropoffDetails,
		&dropoffLng, &dropoffLat,
		&b.Passengers, &b.Suitcases, &b.Notes,
		&b.Status, &b.TotalAmount, &b.TaxRateBasisPoints,
		&b.EstimatedDistance, &b.Co2Grams,
		&b.ExternalInvoiceID, &b.InvoiceURL,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	b.PickupLat = pickupLat
	b.PickupLng = pickupLng
	b.DropoffLat = dropoffLat
	b.DropoffLng = dropoffLng

	return b, nil
}

func (r *postgresRepository) FindByClientID(ctx context.Context, clientID string) ([]Booking, error) {
	const q = `
		SELECT id, client_id, vehicle_id,
		       scheduled_at,
		       pickup_street, pickup_city, pickup_country, pickup_details,
		       ST_X(pickup_coordinates::geometry)  AS pickup_lng,
		       ST_Y(pickup_coordinates::geometry)  AS pickup_lat,
		       dropoff_street, dropoff_city, dropoff_country, dropoff_details,
		       ST_X(dropoff_coordinates::geometry) AS dropoff_lng,
		       ST_Y(dropoff_coordinates::geometry) AS dropoff_lat,
		       passengers, suitcases, notes,
		       status, total_amount, tax_rate_basis_points,
		       estimated_distance, co2_grams,
		       external_invoice_id, invoice_url,
		       created_at, updated_at
		FROM booking.bookings
		WHERE client_id = $1
		ORDER BY scheduled_at DESC`

	rows, err := r.pool.Query(ctx, q, toUUID(clientID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		b := Booking{}
		var pickupLng, pickupLat, dropoffLng, dropoffLat *float64

		if err := rows.Scan(
			&b.ID, &b.ClientID, &b.VehicleID,
			&b.ScheduledAt,
			&b.PickupStreet, &b.PickupCity, &b.PickupCountry, &b.PickupDetails,
			&pickupLng, &pickupLat,
			&b.DropoffStreet, &b.DropoffCity, &b.DropoffCountry, &b.DropoffDetails,
			&dropoffLng, &dropoffLat,
			&b.Passengers, &b.Suitcases, &b.Notes,
			&b.Status, &b.TotalAmount, &b.TaxRateBasisPoints,
			&b.EstimatedDistance, &b.Co2Grams,
			&b.ExternalInvoiceID, &b.InvoiceURL,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}

		b.PickupLat = pickupLat
		b.PickupLng = pickupLng
		b.DropoffLat = dropoffLat
		b.DropoffLng = dropoffLng

		bookings = append(bookings, b)
	}

	return bookings, rows.Err()
}

func (r *postgresRepository) Create(ctx context.Context, b *Booking) error {
	// Insert core fields via sqlc
	err := r.q.CreateBooking(ctx, db.CreateBookingParams{
		ID:                 toUUID(b.ID),
		ClientID:           toUUID(b.ClientID),
		VehicleID:          toUUID(b.VehicleID),
		ScheduledAt:        toTimestamptz(b.ScheduledAt),
		PickupStreet:       b.PickupStreet,
		PickupCity:         b.PickupCity,
		PickupCountry:      b.PickupCountry,
		PickupDetails:      toText(b.PickupDetails),
		DropoffStreet:      b.DropoffStreet,
		DropoffCity:        b.DropoffCity,
		DropoffCountry:     b.DropoffCountry,
		DropoffDetails:     toText(b.DropoffDetails),
		Passengers:         int16(b.Passengers),
		Suitcases:          int16(b.Suitcases),
		Notes:              toText(b.Notes),
		Status:             db.BookingBookingStatus(b.Status),
		TotalAmount:        toInt4(b.TotalAmount),
		TaxRateBasisPoints: toNullInt2(b.TaxRateBasisPoints),
		EstimatedDistance:  toInt4(b.EstimatedDistance),
		Co2Grams:           toInt4(b.Co2Grams),
		CreatedAt:          toTimestamptz(b.CreatedAt),
		UpdatedAt:          toTimestamptz(b.UpdatedAt),
	})
	if err != nil {
		return err
	}

	// Set coordinates via raw SQL if provided
	if b.PickupLng != nil && b.PickupLat != nil && b.DropoffLng != nil && b.DropoffLat != nil {
		const coordSQL = `
			UPDATE booking.bookings
			SET pickup_coordinates  = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
			    dropoff_coordinates = ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography,
			    updated_at          = $6
			WHERE id = $1`

		_, err = r.pool.Exec(ctx, coordSQL,
			toUUID(b.ID),
			*b.PickupLng, *b.PickupLat,
			*b.DropoffLng, *b.DropoffLat,
			b.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return nil
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
