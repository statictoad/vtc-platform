CREATE SCHEMA IF NOT EXISTS booking;

CREATE EXTENSION IF NOT EXISTS postgis;

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
    pickup_street         VARCHAR(200)           NOT NULL,
    pickup_city           VARCHAR(100)           NOT NULL,
    pickup_country        VARCHAR(100)           NOT NULL,
    pickup_details        VARCHAR(500),
    pickup_coordinates    GEOGRAPHY(POINT, 4326),
    dropoff_street        VARCHAR(200)           NOT NULL,
    dropoff_city          VARCHAR(100)           NOT NULL,
    dropoff_country       VARCHAR(100)           NOT NULL,
    dropoff_details       VARCHAR(500),
    dropoff_coordinates   GEOGRAPHY(POINT, 4326),
    passengers            SMALLINT               NOT NULL DEFAULT 1 CHECK (passengers BETWEEN 1 AND 6),
    suitcases             SMALLINT               NOT NULL DEFAULT 0 CHECK (suitcases BETWEEN 0 AND 10),
    notes                 TEXT,
    status                booking.booking_status NOT NULL DEFAULT 'PENDING',
    total_amount          INTEGER,
    tax_rate_basis_points SMALLINT,
    estimated_distance    INTEGER,
    co2_grams             INTEGER,
    external_invoice_id   TEXT,
    invoice_url           TEXT,
    created_at            TIMESTAMPTZ            NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ            NOT NULL DEFAULT NOW(),

    CONSTRAINT bookings_pkey PRIMARY KEY (id)
);