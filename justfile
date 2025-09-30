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

format:
    @echo "Formatting Go code..."
    go tool golang.org/x/tools/cmd/goimports -w .
    go tool mvdan.cc/gofumpt -w .
    @echo "All code formatted!"
