FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy static files and templates
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV DB_HOST=postgres
ENV DB_USER=jujudb
ENV DB_PASSWORD=your-secure-postgres-password
ENV DB_NAME=jujudb
ENV SESSION_KEY=your-super-secret-session-key-change-in-production
ENV APP_PASSWORD=your-secure-app-password

CMD ["./main"]
