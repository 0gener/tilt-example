# syntax=docker/dockerfile:1
# Use the official Golang image with Debian Bookworm for the build stage
FROM golang:1.22-bookworm AS builder

# Set the working directory inside the builder container
WORKDIR /app

# Copy go.mod and go.sum files
COPY ../../go.mod ./
COPY ../../go.sum ./

# Copy the vendor directory
COPY ../../vendor ./vendor/

# Copy the necessary folders for eventconsumerservice
COPY ../../../cmd/eventconsumerservice ./cmd/eventconsumerservice
COPY ../../../internal/app/eventconsumerservice ./internal/app/eventconsumerservice
COPY ../../../internal/app/common ./internal/app/common

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/eventconsumerservice ./cmd/eventconsumerservice

# Use the distroless base image for the final stage
FROM gcr.io/distroless/base-debian12

# Set environment variable for Gin mode
ENV GIN_MODE=release

# Set the working directory inside the final container
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /go/bin/eventconsumerservice /app/

# Run as non-root user
USER nonroot:nonroot

# Define the entry point of the container
ENTRYPOINT ["./eventconsumerservice"]