FROM node:22-alpine AS frontend

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN NUXT_PUBLIC_API_BASE= npm run generate

FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /app/frontend/.output/public ./internal/api/static/

ARG VERSION=dev
ARG COMMIT=none
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /mcp-research ./cmd/mcp-research

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /mcp-research /usr/local/bin/mcp-research

EXPOSE 8088 8081

ENTRYPOINT ["mcp-research"]
CMD ["--config", "/etc/mcp-research/config.yaml"]
