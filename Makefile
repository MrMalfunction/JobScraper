up-dev:
	docker compose -f infra/dev/docker-compose.yml up --build
up-prod:
	docker compose -f infra/prod/docker-compose.yml up --build -d
