CREATE SCHEMA IF NOT EXISTS booking;

CREATE TYPE booking.booking_status AS ENUM (
    'PENDING',
    'CONFIRMED',
    'COMPLETED',
    'CANCELLED'
);

CREATE TABLE booking.bookings (
    id                    UUID                   NOT NULL,
    client_id             UUID                   NOT NULL,
    vehicle_id            UUID                   NOT NULL,
    scheduled_at          TIMESTAMPTZ            NOT NULL,
    pickup_address        TEXT                   NOT NULL,
    dropoff_address       TEXT                   NOT NULL,
    status                booking.booking_status NOT NULL DEFAULT 'PENDING',
    total_amount          INTEGER,
    tax_rate_basis_points INTEGER,
    estimated_km          FLOAT,
    co2_grams             INTEGER,
    external_invoice_id   TEXT,
    invoice_url           TEXT,
    created_at            TIMESTAMPTZ            NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ            NOT NULL DEFAULT NOW(),

    CONSTRAINT bookings_pkey PRIMARY KEY (id)
);