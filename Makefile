.PHONY: build build-all run run-memory run-sse test clean frontend-install frontend-dev frontend-build frontend-embed storybook storybook-build

# A locally built binary should say which build it is, since the footer now
# shows it. `git describe` gives `v1.4.0` on a tag, `v1.4.0-3-gabc1234` three
# commits later, and `-dirty` with uncommitted work — which is exactly what you
# want to read back when somebody says "it does this on my machine".
# `?=` so CI, which passes VERSION explicitly, still wins.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
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

# Storybook
storybook:
	cd frontend && npm run storybook

storybook-build:
	cd frontend && npm run build-storybook

# Build frontend and copy to Go embed directory
frontend-embed: frontend-build
	rm -rf $(STATIC_DIR)
	cp -r frontend/.output/public $(STATIC_DIR)
	@echo "Frontend embedded into $(STATIC_DIR)"
