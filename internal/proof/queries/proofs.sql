-- name: GetProofByBookingID :one
SELECT id, booked_at,
       operator_name, operator_siret, operator_evtc_number,
       vehicle_number_plate,
       client_first_name, client_last_name, client_phone,
       pickup_address, dropoff_address, scheduled_at,
       proof_hash,
       booking_id,
       created_at, updated_at
FROM proof.proofs
WHERE booking_id = $1;

-- name: GetProofByID :one
SELECT id, booked_at,
       operator_name, operator_siret, operator_evtc_number,
       vehicle_number_plate,
       client_first_name, client_last_name, client_phone,
       pickup_address, dropoff_address, scheduled_at,
       proof_hash,
       booking_id,
       created_at, updated_at
FROM proof.proofs
WHERE id = $1;

-- name: CreateProof :exec
INSERT INTO proof.proofs (
    id, booked_at,
    operator_name, operator_siret, operator_evtc_number,
    vehicle_number_plate,
    client_first_name, client_last_name, client_phone,
    pickup_address, dropoff_address, scheduled_at,
    proof_hash,
    booking_id,
    created_at, updated_at
) VALUES (
    $1, $2,
    $3, $4, $5,
    $6,
    $7, $8, $9,
    $10, $11, $12,
    $13,
    $14,
    $15, $16
);

-- name: UpdateProofHash :execrows
UPDATE proof.proofs
SET proof_hash = $2,
    updated_at = $3
WHERE id = $1;