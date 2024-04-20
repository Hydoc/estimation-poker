up:
	docker compose up -d --build

up-prod:
	docker compose -f docker-compose.prod.yml up -d --build

down-prod:
	docker compose -f docker-compose.prod.yml down

down:
	docker compose down --remove-orphans

check-backend-test:
	docker compose exec backend go test ./...

check-frontend-test:
	docker compose exec frontend pnpm test

check-frontend-format:
	docker compose exec frontend pnpm format

check-frontend-lint:
	docker compose exec frontend pnpm lint

# make version=X.X.X build-and-push-image
build-and-push-image:
	docker build -t hydoc/estimation:$(version) -f deployment.Dockerfile .
	docker push hydoc/estimation:$(version)

check-frontend: check-frontend-format check-frontend-test check-frontend-lint

logs-backend:
	docker compose logs -tf backend

restart: down up
