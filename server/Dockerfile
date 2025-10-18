FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go server file
COPY server.go .

# Build the binary
RUN go clean -cache
RUN go build -o server server.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Create uploads directory
RUN mkdir -p /app/uploads

# Expose port
EXPOSE 8280

# Run the server
CMD ["./server"]