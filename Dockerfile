# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -o str ./cmd/str

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Tokyo

# Copy binary from builder
COPY --from=builder /build/str .

# Copy static files (if exists)
COPY --from=builder /build/web ./web

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./str"]
