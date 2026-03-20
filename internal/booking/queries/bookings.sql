-- name: GetBookingByID :one
SELECT id, client_id, vehicle_id,
       scheduled_at, pickup_address, dropoff_address,
       status, total_amount, tax_rate_basis_points,
       estimated_km, co2_grams,
       external_invoice_id, invoice_url,
       created_at, updated_at
FROM booking.bookings
WHERE id = $1;

-- name: ListBookingsByClientID :many
SELECT id, client_id, vehicle_id,
       scheduled_at, pickup_address, dropoff_address,
       status, total_amount, tax_rate_basis_points,
       estimated_km, co2_grams,
       external_invoice_id, invoice_url,
       created_at, updated_at
FROM booking.bookings
WHERE client_id = $1
ORDER BY scheduled_at DESC;

-- name: CreateBooking :exec
INSERT INTO booking.bookings (
    id, client_id, vehicle_id,
    scheduled_at, pickup_address, dropoff_address,
    status, total_amount, tax_rate_basis_points,
    estimated_km, co2_grams,
    created_at, updated_at
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9,
    $10, $11,
    $12, $13
);

-- name: UpdateBookingStatus :execrows
UPDATE booking.bookings
SET status     = $2,
    updated_at = $3
WHERE id = $1;

-- name: UpdateBookingInvoice :execrows
UPDATE booking.bookings
SET external_invoice_id = $2,
    invoice_url         = $3,
    updated_at          = $4
WHERE id = $1;