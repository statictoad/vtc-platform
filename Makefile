COMPOSE_FILE := deployments/local/compose.yml
MIGRATE      := migrate
DB_URL       := postgres://vtc:vtc@localhost:5432/vtc?sslmode=disable

# =============================================================================
# Infrastructure
# =============================================================================

.PHONY: up
up:
	podman compose -f $(COMPOSE_FILE) up -d

.PHONY: down
down:
	podman compose -f $(COMPOSE_FILE) down

.PHONY: down-v
down-v:
	podman compose -f $(COMPOSE_FILE) down -v

.PHONY: logs
logs:
	podman compose -f $(COMPOSE_FILE) logs -f

.PHONY: reset
reset: down-v up
	@echo "Waiting for Postgres to be ready..."
	@sleep 3
	$(MAKE) migrate-all

# =============================================================================
# Migrations
# =============================================================================

.PHONY: create-schemas
create-schemas:
	podman exec -it vtc-postgres psql -U vtc -d vtc -c \
	  "CREATE SCHEMA IF NOT EXISTS booking; \
	   CREATE SCHEMA IF NOT EXISTS client; \
	   CREATE SCHEMA IF NOT EXISTS fleet; \
	   CREATE SCHEMA IF NOT EXISTS proof;"

.PHONY: migrate-booking-up
migrate-booking-up:
	$(MIGRATE) -path internal/booking/migrations \
	           -database "$(DB_URL)&search_path=booking&x-migrations-table=schema_migrations" up

.PHONY: migrate-booking-down
migrate-booking-down:
	$(MIGRATE) -path internal/booking/migrations \
	           -database "$(DB_URL)&search_path=booking&x-migrations-table=schema_migrations" down 1

.PHONY: migrate-client-up
migrate-client-up:
	$(MIGRATE) -path internal/client/migrations \
	           -database "$(DB_URL)&search_path=client&x-migrations-table=schema_migrations" up

.PHONY: migrate-client-down
migrate-client-down:
	$(MIGRATE) -path internal/client/migrations \
	           -database "$(DB_URL)&search_path=client&x-migrations-table=schema_migrations" down 1

.PHONY: migrate-fleet-up
migrate-fleet-up:
	$(MIGRATE) -path internal/fleet/migrations \
	           -database "$(DB_URL)&search_path=fleet&x-migrations-table=schema_migrations" up

.PHONY: migrate-fleet-down
migrate-fleet-down:
	$(MIGRATE) -path internal/fleet/migrations \
	           -database "$(DB_URL)&search_path=fleet&x-migrations-table=schema_migrations" down 1

.PHONY: migrate-proof-up
migrate-proof-up:
	$(MIGRATE) -path internal/proof/migrations \
	           -database "$(DB_URL)&search_path=proof&x-migrations-table=schema_migrations" up

.PHONY: migrate-proof-down
migrate-proof-down:
	$(MIGRATE) -path internal/proof/migrations \
	           -database "$(DB_URL)&search_path=proof&x-migrations-table=schema_migrations" down 1

.PHONY: migrate-all
migrate-all: create-schemas migrate-booking-up migrate-client-up migrate-fleet-up migrate-proof-up

# =============================================================================
# Queries
# =============================================================================

.PHONY: sqlc
sqlc:
	cd internal/booking && sqlc generate
	cd internal/client  && sqlc generate
	cd internal/fleet   && sqlc generate
	cd internal/proof   && sqlc generate

# =============================================================================
# Seeds
# =============================================================================

.PHONY: seed-client
seed-client:
	podman cp internal/client/seeds/clients-data.sql vtc-postgres:/tmp/clients-data.sql
	podman exec vtc-postgres psql -U vtc -d vtc -f /tmp/clients-data.sql

.PHONY: seed-fleet
seed-fleet:
	podman cp internal/fleet/seeds/vehicles-data.sql vtc-postgres:/tmp/vehicles-data.sql
	podman exec vtc-postgres psql -U vtc -d vtc -f /tmp/vehicles-data.sql

.PHONY: seed
seed: seed-client seed-fleet

# =============================================================================
# Dev
# =============================================================================

.PHONY: run-booking
run-booking:
	set -a && . ./.env && set +a && \
	DATABASE_URL=$$BOOKING_SERVICE_DB_URL PORT=$$BOOKING_SERVICE_PORT \
	go run ./cmd/booking-api

.PHONY: run-client
run-client:
	set -a && . ./.env && set +a && \
	DATABASE_URL=$$CLIENT_SERVICE_DB_URL PORT=$$CLIENT_SERVICE_PORT \
	go run ./cmd/client-api

.PHONY: run-fleet
run-fleet:
	set -a && . ./.env && set +a && \
	DATABASE_URL=$$FLEET_SERVICE_DB_URL PORT=$$FLEET_SERVICE_PORT \
	go run ./cmd/fleet-api

.PHONY: run-gateway
run-gateway:
	set -a && . ./.env && set +a && \
	PORT=$$GATEWAY_PORT \
	go run ./cmd/gateway

.PHONY: dev
dev:
	overmind start -f Procfile

# =============================================================================
# Build
# =============================================================================

.PHONY: build
build:
	go build -o bin/gateway             ./cmd/gateway
	go build -o bin/booking-api         ./cmd/booking-api
	go build -o bin/client-api          ./cmd/client-api
	go build -o bin/fleet-api           ./cmd/fleet-api
	go build -o bin/proof-api           ./cmd/proof-api
	go build -o bin/billing-api         ./cmd/billing-api
	go build -o bin/notification-worker ./cmd/notification-worker

# =============================================================================
# Test
# =============================================================================

.PHONY: test
test:
	go test ./...

.PHONY: test-booking
test-booking:
	go test ./internal/booking/...

.PHONY: test-fleet
test-fleet:
	go test ./internal/fleet/...
