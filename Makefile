.PHONY: all build dev test clean lint up down db/up db/down get-go-version install-tools fmt check-fmt wire

# central Go version
# Default to the stable project version. Change here to pin Go for CI/dev.
GO_VERSION ?= 1.24

# local tooling/.bin directory
BIN ?= .bin

# Tool versions (managed here)
GOFUMPT_VERSION ?= v0.6.0
GOLANGCI_VERSION ?= v1.64.8
WIRE_VERSION ?= v0.7.0

# Tool paths
GOFUMPT := $(BIN)/gofumpt
GOLANGCI := $(BIN)/golangci-lint
WIRE := $(BIN)/wire

# Development
all: up dev

dev:
	air -c .air.toml

build:
	CGO_ENABLED=0 go build -o .bin/api ./cmd/api

test:
	make testdb/up
	make wait-for-testdb
	TZ="UTC" TEST_DATABASE_DSN="user=postgres password=postgres host=localhost port=5433 dbname=tsudzuri_test sslmode=disable" go test ./...
	make testdb/down

clean:
	rm -rf .bin/
	go clean -testcache

# Docker
up:
	# pass the configured GO_VERSION into docker compose so builds use the managed version
	GO_VERSION=$(GO_VERSION) docker compose up --build -d

down:
	docker compose down

# Database (container helpers)
db/up:
	# Start only the db service so we can iterate on DB without rebuilding other services
	GO_VERSION=$(GO_VERSION) docker compose up -d --build db

db/down:
	# Stop and remove the db service (data volume preserved by default)
	docker compose stop db || true
	docker compose rm -f -v db || true

migration: db/down db/up
	@echo "Migration completed: database container restarted with fresh schema"


testdb/up:
	# Start only the testdb service so we can iterate on test DB without rebuilding other services
	GO_VERSION=$(GO_VERSION) docker compose up -d --build testdb

testdb/down:
	# Stop and remove the testdb service (data volume preserved by default)
	docker compose stop testdb || true
	docker compose rm -f -v testdb || true

wait-for-testdb:
	@echo "Waiting for test database to be ready..."
	# Use pg_isready inside the container to avoid requiring local Postgres client tools
	@timeout=60; \
	while ! docker compose exec -T testdb pg_isready -U postgres -d tsudzuri_test > /dev/null 2>&1; do \
		if [ $$timeout -le 0 ]; then \
			echo "Timed out waiting for test database to be ready."; \
			exit 1; \
		fi; \
		echo "Test database is not ready yet. Waiting..."; \
		timeout=$$((timeout-1)); \
		sleep 1; \
	done; \
	echo "Test database is ready."

# Install developer tools (golangci-lint, gofumpt)
install-tools:
	@echo "Installing developer tooling into $(BIN)..."
	@mkdir -p $(BIN)
	@GOBIN=$(abspath $(BIN)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_VERSION}
	@GOBIN=$(abspath $(BIN)) go install mvdan.cc/gofumpt@${GOFUMPT_VERSION}
	@GOBIN=$(abspath $(BIN)) go install github.com/google/wire/cmd/wire@${WIRE_VERSION}
	@echo "Installed: $(GOLANGCI) $(GOFUMPT)"
	@echo "Installed: $(WIRE)"

# Linting
lint:
	@echo "Running golangci-lint..."
	@$(GOLANGCI) run ./...
	@echo "golangci-lint completed."

# Format code using gofumpt
fmt:
	@echo "Running gofumpt..."
	@$(GOFUMPT) -w .
	@echo "gofumpt completed."

# Dependencies
deps:
	go mod tidy
	go mod verify

# Print the configured Go version (for CI to consume)
get-go-version:
	@echo $(GO_VERSION)

# Generate code (e.g., mocks)
generate:
	# Generate all code (Ent + mocks, etc.) via go:generate directives
	go generate ./...

wire:
	@$(WIRE) ./cmd/api
