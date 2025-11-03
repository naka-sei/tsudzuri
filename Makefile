.PHONY: all build dev test clean lint migrate migrate-down up down get-go-version install-tools fmt check-fmt

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

# Database
migrate:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/tsudzuri?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/tsudzuri?sslmode=disable" down

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