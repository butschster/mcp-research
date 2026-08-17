# Pinned to the builder's own platform: the frontend build is architecture-neutral,
# and running npm/Nuxt under QEMU for arm64 costs many minutes for nothing.
FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend

WORKDIR /app/frontend
COPY frontend/package.json ./
# npm 10, which node:22 bundles, crashes resolving this tree without a lockfile
# ("Cannot read properties of null (reading 'edgesOut')"), and the lockfile is
# deliberately not in the repository. Without this the release image stops
# building — which is how a deploy discovers the problem.
RUN npm install -g npm@11 && npm install
COPY frontend/ ./
RUN NUXT_PUBLIC_API_BASE= npm run generate

# Same reasoning: build on the native platform and let Go cross-compile.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /app/frontend/.output/public ./internal/api/static/

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
# Supplied by buildx for the image being produced.
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o /mcp-research ./cmd/mcp-research

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /mcp-research /usr/local/bin/mcp-research

EXPOSE 8088 8081

ENTRYPOINT ["mcp-research"]
CMD ["--config", "/etc/mcp-research/config.yaml"]
