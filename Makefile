# Load environment variables from .env
ifneq (,$(wildcard ./.env))
	include .env
	export $(shell sed 's/=.*//' .env)
endif

build:
	@go build -o bin/uploady cmd/main.go

run: build
	@./bin/uploady

test:
	@go test -v ./...
# Use 'MIGRATION_NAME' to pass the name of the migration you want to create
migration:
	@if [ -z "$(MIGRATION_NAME)" ]; then \
		echo "Error: Please provide a migration name using MIGRATION_NAME=<name>"; \
		exit 1; \
	else \
		migrate create -ext sql -dir cmd/migrate/migrations "$(MIGRATION_NAME)"; \
	fi

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down

# Create the database using a raw SQL command in the Makefile
create-database:
	@echo "Creating database: $(DB_NAME) on host: $(DB_HOST)..."
	@if [ -z "$(DB_USER)" ] || [ -z "$(DB_PASSWORD)" ] || [ -z "$(DB_HOST)" ] || [ -z "$(DB_NAME)" ]; then \
		echo "Error: Please ensure DB_USER, DB_PASSWORD, DB_HOST, and DB_NAME are defined in the .env file"; \
		exit 1; \
	else \
		mysql -u $(DB_USER) -p$(DB_PASSWORD) -h $(DB_HOST) -e "CREATE DATABASE IF NOT EXISTS \`$(DB_NAME)\`"; \
		echo "Database '$(DB_NAME)' created successfully"; \
	fi
# Run migrations
migrate-up-docker:
	docker-compose run --rm api migrate -path /app/migrations -database "mysql://root:mypassword@tcp(db:3306)/uploady" up
