.PHONY: build run test clean lint

# Build the server binary
build:
	go build -o bin/flashesale ./cmd/server/main.go

# Run the server
run:
	go run ./cmd/server/main.go

# Run all tests
test:
	go test -v -race -count=1 ./...

# Run tests with coverage
test-cover:
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Lint code (requires golangci-lint)
lint:
	golangci-lint run ./...

# Download dependencies
deps:
	go mod tidy
	go mod download

# Build frontend
frontend-build:
	cd frontend && npm install && npm run build
