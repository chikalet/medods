.PHONY: build up down migrate logs

build:
	docker compose build --no-cache

up:
	docker compose up -d

down:
	docker compose down

migrate:
	docker compose exec app migrate -path /app/migrations -database "$$DB_URL" up

logs:
	docker compose logs -f app