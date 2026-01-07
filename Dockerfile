# Stage 1: Build the Go binary
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .
# Copy the .env file (optional, better to use environment variables in prod)
COPY --from=builder /app/.env . 

EXPOSE 8080

CMD ["./main"]