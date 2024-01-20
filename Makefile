up:
	docker compose up -d --build

down:
	docker compose down --remove-orphans

check-frontend-test:
	docker compose exec frontend pnpm test

check-frontend-format:
	docker compose exec frontend pnpm format

check-frontend-lint:
	docker compose exec frontend pnpm lint

check-frontend: check-frontend-format check-frontend-test check-frontend-lint

restart: down up