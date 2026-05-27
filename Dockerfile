# Build stage
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source files
COPY . .

# Build the outlived Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o outlived ./cmd/outlived

# Run stage
FROM debian:bookworm-slim

# Install ca-certificates (required for scraping Wikipedia/HTTPS requests)
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/outlived /app/outlived

# Expose port 8080 by default (Cloud Run will override this with PORT env var)
EXPOSE 8080

ENTRYPOINT ["/app/outlived"]
