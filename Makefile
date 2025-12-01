.PHONY: help install generate run dev build clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "Installing templ..."
	go install github.com/a-h/templ/cmd/templ@latest
	@echo "Installing Go dependencies..."
	go mod tidy
	@echo "Done!"

generate: ## Generate templ templates
	@echo "Generating templates..."
	templ generate

run: generate ## Generate templates and run the server
	@echo "Starting server..."
	go run main.go

dev: generate ## Run in development mode with auto-reload (requires air)
	@echo "Starting development server..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not found. Install with: go install github.com/air-verse/air@latest"; \
		echo "Falling back to regular run..."; \
		make run; \
	fi

build: generate ## Build the application
	@echo "Building application..."
	@GITSHA=$$(git rev-parse HEAD 2>/dev/null || echo "unknown"); \
	go build -ldflags "-X jordanmurray.xyz/site/version.GitSHA=$$GITSHA" -o bin/site main.go
	@echo "Binary created at bin/site"

clean: ## Clean generated files and build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	find . -name "*_templ.go" -type f -delete
	@echo "Done!"

watch: ## Watch templ files for changes
	@echo "Watching templates..."
	templ generate --watch
