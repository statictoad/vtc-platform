CREATE SCHEMA IF NOT EXISTS fleet;

CREATE TABLE fleet.vehicles (
    id           UUID    NOT NULL,
    number_plate TEXT    NOT NULL,
    brand        TEXT    NOT NULL,
    model        TEXT    NOT NULL,
    year         BIGINT  NOT NULL, -- int64 in Go via sqlc, this avoids casting
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vehicles_pkey         PRIMARY KEY (id),
    CONSTRAINT vehicles_number_plate UNIQUE      (number_plate)
);