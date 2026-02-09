.PHONY: infra-up infra-down infra-logs infra-ps \
       proto-gen sqlc-gen generate \
       run-store run-film run-customer run-rental run-payment test lint fmt

# Infrastructure
infra-up:
	docker compose -f deployments/docker-compose.yml up -d

infra-down:
	docker compose -f deployments/docker-compose.yml down

infra-logs:
	docker compose -f deployments/docker-compose.yml logs -f

infra-ps:
	docker compose -f deployments/docker-compose.yml ps

# Code generation
proto-gen:
	buf generate

sqlc-gen:
	sqlc generate

generate: proto-gen sqlc-gen

# Run services
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

# Testing
test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	goimports -w .
