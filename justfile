# justfile

# Build the application binary
build:
    go build -o ./bin/copilot-proxy ./cmd/copilot-proxy

# Run the application directly
run:
    go run ./cmd/copilot-proxy/main.go

# Format all Go source files
fmt:
    go fmt ./...

# Run all tests
test:
    go test -v ./...