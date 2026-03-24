# ZarishSphere — Makefile
.PHONY: dev build test lint clean docker-build docker-push helm-lint

GO_VERSION := 1.26.1
BINARY_NAME := server
DOCKER_REGISTRY := ghcr.io/zarishsphere

# Start the service in development mode (with live reload)
dev:
	air -c .air.toml

# Build the binary
build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/$(BINARY_NAME) ./cmd/server/

# Run all tests
test:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic

# Run linter
lint:
	golangci-lint run ./...

# Run database migrations
migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

# Build Docker image
docker-build:
	docker build -t $(DOCKER_REGISTRY)/$(shell basename $(PWD)):latest .

# Run security scan
security:
	trivy fs . --severity CRITICAL,HIGH

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out

# Validate Helm chart
helm-lint:
	helm lint deploy/helm/

# Format code
fmt:
	gofmt -s -w .
	goimports -w .
