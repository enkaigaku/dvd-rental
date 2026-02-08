.PHONY: infra-up infra-down infra-logs infra-ps \
       proto-gen sqlc-gen generate \
       run-store test lint fmt

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

# Testing
test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	goimports -w .
