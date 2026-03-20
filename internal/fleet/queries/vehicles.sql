-- name: GetVehicleByID :one
SELECT id, number_plate, brand, model, year, active, created_at, updated_at
FROM fleet.vehicles
WHERE id = $1;

-- name: ListActiveVehicles :many
SELECT id, number_plate, brand, model, year, active, created_at, updated_at
FROM fleet.vehicles
WHERE active = true
ORDER BY brand, model;

-- name: CreateVehicle :exec
INSERT INTO fleet.vehicles (
    id, number_plate, brand, model, year, active, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: UpdateVehicle :execrows
UPDATE fleet.vehicles
SET
    number_plate = COALESCE($2, number_plate),
    brand        = COALESCE($3, brand),
    model        = COALESCE($4, model),
    year         = COALESCE($5, year),
    active       = COALESCE($6, active),
    updated_at   = $7
WHERE id = $1;