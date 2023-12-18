# Format all files
fmt:
	@echo "==> Formatting source"
	@gofmt -s -w $(shell find . -type f -name '*.go' -not -path "./vendor/*")
	@echo "==> Done"
.PHONY: fmt

# Tidy the go.mod file
tidy:
	@echo "==> Cleaning go.mod"
	@go mod tidy
	@echo "==> Done"
.PHONY: tidy

# Run all tests
test:
	@go test -cover -race ./...
.PHONY: test

# Lint the project
lint:
	@golangci-lint run ./...
.PHONY: lint

# Build the commands
build:
	@goreleaser release --clean --snapshot
.PHONY: build
