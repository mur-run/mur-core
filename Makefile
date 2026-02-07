.PHONY: build test lint clean install release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/mur-run/mur-core/cmd/mur/cmd.Version=$(VERSION) \
           -X github.com/mur-run/mur-core/cmd/mur/cmd.Commit=$(COMMIT) \
           -X github.com/mur-run/mur-core/cmd/mur/cmd.BuildDate=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o mur ./cmd/mur

test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

clean:
	rm -f mur mur-* coverage.out coverage.html

install: build
	cp mur $(GOPATH)/bin/mur

# Build for all platforms
release:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mur-linux-amd64 ./cmd/mur
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o mur-linux-arm64 ./cmd/mur
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mur-darwin-amd64 ./cmd/mur
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o mur-darwin-arm64 ./cmd/mur
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o mur-windows-amd64.exe ./cmd/mur

# Quick check before commit
check: lint test
	@echo "✅ All checks passed"

help:
	@echo "Available targets:"
	@echo "  build      - Build mur binary"
	@echo "  test       - Run tests"
	@echo "  test-cover - Run tests with coverage"
	@echo "  lint       - Run linter"
	@echo "  clean      - Remove build artifacts"
	@echo "  install    - Build and install to GOPATH/bin"
	@echo "  release    - Build for all platforms"
	@echo "  check      - Run lint and test"
