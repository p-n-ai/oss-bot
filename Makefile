.PHONY: test lint build-cli build-bot docker setup

# Testing
test:
	go test ./...

test-v:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting
lint:
	golangci-lint run ./...

# Building
build-cli:
	CGO_ENABLED=0 go build -o bin/oss ./cmd/oss

build-bot:
	CGO_ENABLED=0 go build -o bin/oss-bot ./cmd/bot

build: build-cli build-bot

# Docker
docker:
	docker build -f deploy/docker/Dockerfile -t oss-bot .

# Setup
setup:
	cp -n .env.example .env 2>/dev/null || true
	go mod download
	@echo "Setup complete. Edit .env with your configuration."
