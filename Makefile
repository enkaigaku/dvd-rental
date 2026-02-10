.PHONY: infra-up infra-down infra-logs infra-ps \
       up down logs ps \
       proto-gen sqlc-gen generate \
       run-store run-film run-customer run-rental run-payment \
       run-customer-bff run-admin-bff \
       build-all test lint fmt

# Infrastructure (local dev: postgres + redis only)
infra-up:
	docker compose -f deployments/docker-compose.yml up -d postgres redis

infra-down:
	docker compose -f deployments/docker-compose.yml down

infra-logs:
	docker compose -f deployments/docker-compose.yml logs -f postgres redis

infra-ps:
	docker compose -f deployments/docker-compose.yml ps

# All services (docker)
up:
	docker compose -f deployments/docker-compose.yml up -d

down:
	docker compose -f deployments/docker-compose.yml down

logs:
	docker compose -f deployments/docker-compose.yml logs -f

ps:
	docker compose -f deployments/docker-compose.yml ps

# Code generation
proto-gen:
	buf generate

sqlc-gen:
	sqlc generate

generate: proto-gen sqlc-gen

# Run gRPC services (local development)
run-store:
	go run ./cmd/store-service

run-film:
	go run ./cmd/film-service

run-customer:
	go run ./cmd/customer-service

run-rental:
	go run ./cmd/rental-service

run-payment:
	go run ./cmd/payment-service

# Run BFF services (local development)
run-customer-bff:
	go run ./cmd/customer-bff

run-admin-bff:
	go run ./cmd/admin-bff

# Build all services
build-all:
	go build ./cmd/store-service
	go build ./cmd/film-service
	go build ./cmd/customer-service
	go build ./cmd/rental-service
	go build ./cmd/payment-service
	go build ./cmd/customer-bff
	go build ./cmd/admin-bff

# Testing
test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	goimports -w .
