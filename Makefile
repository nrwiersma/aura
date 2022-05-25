## Format all files
fmt:
	@echo "==> Formatting source"
	@gofmt -s -w $(shell find . -type f -name '*.go' -not -path "./vendor/*")
	@echo "==> Done"
.PHONY: fmt

## Run the code tests
test:
	@go test -cover -race ./...
.PHONY: test

## Lint the code
lint:
	@golangci-lint run ./...
.PHONY: lint

## Build forward auth
build:
	@go build -o aura ./cmd/aura
.PHONY: build
