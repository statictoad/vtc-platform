CREATE SCHEMA IF NOT EXISTS proof;

CREATE TABLE proof.proofs (
    id                    UUID        NOT NULL,
    booked_at             TIMESTAMPTZ NOT NULL,

    -- Operator snapshot
    operator_name         TEXT        NOT NULL,
    operator_siret        TEXT        NOT NULL,
    operator_evtc_number  TEXT        NOT NULL,

    -- Vehicle snapshot
    vehicle_number_plate  TEXT        NOT NULL,

    -- Client snapshot
    client_first_name     TEXT        NOT NULL,
    client_last_name      TEXT        NOT NULL,
    client_phone          TEXT,

    -- Ride snapshot
    pickup_address        TEXT        NOT NULL,
    dropoff_address       TEXT        NOT NULL,
    scheduled_at          TIMESTAMPTZ NOT NULL,

    -- Tamper detection
    proof_hash            TEXT,

    -- Reference to the booking (ID only — no FK, cross-service boundary)
    booking_id            UUID        NOT NULL,

    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT proofs_pkey       PRIMARY KEY (id),
    CONSTRAINT proofs_booking_id UNIQUE      (booking_id)
);