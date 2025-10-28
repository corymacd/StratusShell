.PHONY: generate build install test integration-test clean help

help:
	@echo "StratusShell Build Commands:"
	@echo "  make generate        - Generate templ files (requires templ CLI)"
	@echo "  make build           - Build binary"
	@echo "  make install         - Install to /usr/local/bin (requires sudo)"
	@echo "  make test            - Run unit tests"
	@echo "  make integration-test - Run integration tests (requires sudo/docker)"
	@echo "  make clean           - Remove build artifacts"

generate:
	@if command -v templ >/dev/null 2>&1; then \
		templ generate; \
	else \
		echo "Warning: templ CLI not found. Skipping code generation."; \
		echo "Install with: go install github.com/a-h/templ/cmd/templ@latest"; \
	fi

build: generate
	go build -o stratusshell main.go

install: build
	sudo cp stratusshell /usr/local/bin/
	sudo mkdir -p /etc/stratusshell
	sudo cp configs/default.yaml /etc/stratusshell/

test:
	go test ./...

integration-test:
	@echo "Integration tests require root privileges"
	@if [ -d "./test/integration" ]; then \
		INTEGRATION_TESTS=1 go test ./test/integration/...; \
	else \
		echo "No integration tests found (./test/integration directory does not exist)"; \
	fi

clean:
	rm -f stratusshell
	find . -name "*_templ.go" -delete
