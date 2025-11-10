# Use the official Go image as the base image
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/rag-service ./cmd/api

# Build the dataloader
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/dataloader ./cmd/dataloader

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install necessary dependencies
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the binaries from the builder stage
COPY --from=builder /app/rag-service .
COPY --from=builder /app/dataloader .

# Copy the data directory with samples
COPY --from=builder /app/data ./data

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./rag-service"]