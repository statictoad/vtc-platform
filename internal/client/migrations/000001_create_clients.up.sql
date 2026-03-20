CREATE SCHEMA IF NOT EXISTS client;

CREATE TABLE client.clients (
    id            UUID        NOT NULL,
    clerk_user_id TEXT        NOT NULL,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    email         TEXT        NOT NULL,
    phone         TEXT,
    notes         TEXT,
    can_pay_later BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT clients_pkey          PRIMARY KEY (id),
    CONSTRAINT clients_clerk_user_id UNIQUE      (clerk_user_id)
);