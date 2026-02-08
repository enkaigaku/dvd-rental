.PHONY: infra-up infra-down infra-logs infra-ps

# Infrastructure
infra-up:
	docker compose -f deployments/docker-compose.yml up -d

infra-down:
	docker compose -f deployments/docker-compose.yml down

infra-logs:
	docker compose -f deployments/docker-compose.yml logs -f

infra-ps:
	docker compose -f deployments/docker-compose.yml ps
