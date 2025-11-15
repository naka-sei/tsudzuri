.PHONY: all build dev test clean lint up down db/up db/down get-go-version install-tools fmt check-fmt wire swagger swagger/down

# central Go version
# Default to the stable project version. Change here to pin Go for CI/dev.
GO_VERSION ?= 1.24

# local tooling/.bin directory
BIN ?= $(CURDIR)/.bin

# UNAME info
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

# Tool versions (managed here)
GOFUMPT_VERSION ?= v0.6.0
GOLANGCI_VERSION ?= v1.64.8
WIRE_VERSION ?= v0.7.0
BUF_VERSION ?= v1.59.0

# Tool paths
GOFUMPT := $(BIN)/gofumpt
GOLANGCI := $(BIN)/golangci-lint
WIRE := $(BIN)/wire
BUF := $(BIN)/buf

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
	GO_VERSION=$(GO_VERSION) docker compose up -d --build api db swagger-ui

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
	@curl -sSL "https://github.com/bufbuild/buf/releases/download/$(BUF_VERSION)/buf-$(UNAME_OS)-$(UNAME_ARCH)" -o $(BUF)
	@chmod +x $(BUF)
	@echo "Installed: $(BUF)"

# Linting
lint:
	@echo "Running golangci-lint..."
	@$(GOLANGCI) run ./...
	@echo "golangci-lint completed."

# Format code using gofumpt and proto formatting
fmt:
	@echo "Running gofumpt..."
	@$(GOFUMPT) -w .
	@echo "gofumpt completed."
	@echo "Running buf format..."
	@cd api/protobuf && $(BUF) format -w; cd -
	@echo "buf format completed."

# Dependencies
deps:
	go mod tidy
	go mod verify
	go mod vendor

# Print the configured Go version (for CI to consume)
get-go-version:
	@echo $(GO_VERSION)

# generate/protobuf/go
generate/protobuf/go:
	@cd api/protobuf && $(BUF) dep update && $(BUF) build && $(BUF) generate; cd -
	@echo "Protobuf code generation completed."

# Generate TypeScript types from generated OpenAPI (for frontend)
generate/typescript:
	@echo "Generating TypeScript types from OpenAPI..."
	@# Convert Swagger v2 -> OpenAPI v3 then generate types
	@echo "Converting Swagger v2 to OpenAPI v3..."
	@npx swagger2openapi api/protobuf/gen/openapi/tsudzuri/v1/tsudzuri.swagger.json -o api/protobuf/gen/openapi/tsudzuri/v1/tsudzuri.openapi.json --yaml=false
	@echo "Generating TypeScript types from converted OpenAPI v3..."
	@npx openapi-typescript api/protobuf/gen/openapi/tsudzuri/v1/tsudzuri.openapi.json -o api/protobuf/gen/openapi/tsudzuri/v1/types.ts
	@echo "TypeScript types generated at api/protobuf/gen/openapi/tsudzuri/v1/types.ts"

# Generate code (e.g., mocks)
generate:
	# Generate all code (Ent + mocks, etc.) via go:generate directives
	go generate ./...
	make generate/protobuf/go
	make deps

generate/wire:
	@$(WIRE) ./cmd/api

# Swagger UI
swagger:
	# Start only the swagger-ui service to browse the generated OpenAPI spec
	GO_VERSION=$(GO_VERSION) docker compose up -d --build swagger-ui

swagger/down:
	# Stop and remove the swagger-ui service
	docker compose stop swagger-ui || true
	docker compose rm -f -v swagger-ui || true
