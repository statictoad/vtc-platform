-- name: GetBookingByID :one
SELECT id, client_id, vehicle_id,
       scheduled_at,
       pickup_street, pickup_city, pickup_country, pickup_details,
       dropoff_street, dropoff_city, dropoff_country, dropoff_details,
       passengers, suitcases, notes,
       status, total_amount, tax_rate_basis_points,
       estimated_distance, co2_grams,
       external_invoice_id, invoice_url,
       created_at, updated_at
FROM booking.bookings
WHERE id = $1;

-- name: ListBookingsByClientID :many
SELECT id, client_id, vehicle_id,
       scheduled_at,
       pickup_street, pickup_city, pickup_country, pickup_details,
       dropoff_street, dropoff_city, dropoff_country, dropoff_details,
       passengers, suitcases, notes,
       status, total_amount, tax_rate_basis_points,
       estimated_distance, co2_grams,
       external_invoice_id, invoice_url,
       created_at, updated_at
FROM booking.bookings
WHERE client_id = $1
ORDER BY scheduled_at DESC;

-- name: CreateBooking :exec
INSERT INTO booking.bookings (
    id, client_id, vehicle_id,
    scheduled_at,
    pickup_street, pickup_city, pickup_country, pickup_details,
    dropoff_street, dropoff_city, dropoff_country, dropoff_details,
    passengers, suitcases, notes,
    status, total_amount, tax_rate_basis_points,
    estimated_distance, co2_grams,
    created_at, updated_at
) VALUES (
    $1, $2, $3,
    $4,
    $5, $6, $7, $8,
    $9, $10, $11, $12,
    $13, $14, $15,
    $16, $17, $18,
    $19, $20,
    $21, $22
);

-- name: UpdateBookingCoordinates :exec
UPDATE booking.bookings
SET pickup_coordinates  = ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
    dropoff_coordinates = ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography,
    updated_at          = $6
WHERE id = $1;

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