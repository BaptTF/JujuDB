FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Final stage
FROM scratch

WORKDIR /

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy static files and templates
COPY --from=builder /app/web ./web

# Expose port
EXPOSE 8080

CMD ["./main"]
