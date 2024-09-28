build:
	@go build -o bin/uploady cmd/main.go

run: build
	@./bin/uploady