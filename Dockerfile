# ── Build stage ──────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build a fully static binary (CGO disabled)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w -extldflags "-static"' \
    -o /build/api ./cmd/api/

# ── Runtime stage ────────────────────────────────────────────
FROM scratch

# TLS certs (needed for neo4j+s:// or any HTTPS calls)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Binary
COPY --from=builder /build/api /api

# Film seed data for auto-population on first boot
COPY --from=builder /build/filtered_films.csv /filtered_films.csv

EXPOSE 4000

ENTRYPOINT ["/api"]
