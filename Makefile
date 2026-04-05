.PHONY: build build-all run run-memory run-sse test clean frontend-install frontend-dev frontend-build frontend-embed

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
STATIC_DIR := internal/api/static

# Build Go binary only (uses whatever is in static/)
build:
	go build $(LDFLAGS) -o bin/mcp-research ./cmd/mcp-research

# Full build: frontend + Go binary with embedded UI
build-all: frontend-embed build

# Run with file DB
run: build
	./bin/mcp-research --db research.db

run-memory: build
	./bin/mcp-research

run-sse: build
	./bin/mcp-research --transport sse --mcp-port 8081 --db research.db

test:
	go test ./cmd/... ./internal/...

clean:
	rm -rf bin/
	rm -rf frontend/.nuxt frontend/.output frontend/dist

# Frontend
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && NUXT_PUBLIC_API_BASE=http://localhost:8088 npm run dev

frontend-build:
	cd frontend && NUXT_PUBLIC_API_BASE= npm run generate

# Build frontend and copy to Go embed directory
frontend-embed: frontend-build
	rm -rf $(STATIC_DIR)
	cp -r frontend/.output/public $(STATIC_DIR)
	@echo "Frontend embedded into $(STATIC_DIR)"
