.PHONY: all build dev test clean lint up down db/up db/down get-go-version install-tools fmt check-fmt

# central Go version
# Default to the stable project version. Change here to pin Go for CI/dev.
GO_VERSION ?= 1.24

# local tooling/.bin directory
BIN ?= .bin

# Tool versions (managed here)
GOFUMPT_VERSION ?= v0.6.0
GOLANGCI_VERSION ?= v1.64.8

# Tool paths
GOFUMPT := $(BIN)/gofumpt
GOLANGCI := $(BIN)/golangci-lint

# Development
all: up dev

dev:
	air -c .air.toml

build:
	CGO_ENABLED=0 go build -o .bin/api cmd/api/main.go

test:
	go test -v -race -cover ./...

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

# Install developer tools (golangci-lint, gofumpt)
install-tools:
	@echo "Installing developer tooling into $(BIN)..."
	@mkdir -p $(BIN)
	@GOBIN=$(abspath $(BIN)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_VERSION}
	@GOBIN=$(abspath $(BIN)) go install mvdan.cc/gofumpt@${GOFUMPT_VERSION}
	@echo "Installed: $(GOLANGCI) $(GOFUMPT)"

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
	go generate ./...
