build:
	@go build -o bin/uploady cmd/main.go

run: build
	@./bin/uploady

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
