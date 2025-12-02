# Stage 1: The Builder
# Use the official, full Go image for compiling. It has the Go compiler and all build tools.
FROM golang:1.22-alpine AS builder

# Set CGO_ENABLED=0 to ensure a fully static binary (no external C dependencies needed at runtime).
# This is usually required if using 'scratch' or 'distroless' as the final stage.
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your application source code
COPY . .

# Build the application. Use -ldflags to strip debugging symbols for a smaller binary.
RUN go build -ldflags="-w -s" -o /go-api ./main.go

# --- End of Build Stage ---

# Stage 2: The Final Production Image
# Use a minimal image to run the binary. 'scratch' is the absolute minimum.
FROM scratch

# If your Go code makes HTTPS requests (which it will for object storage) or uses timezones, 
# you might need certificates and timezone data. We copy them from the build stage.
# NOTE: If you use gcr.io/distroless/static, these files are usually included, simplifying this step.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary from the builder stage
COPY --from=builder /go-api /go-api

# The object storage will likely need a directory for uploads/data
# You can create this directory if needed, or mount a volume to it at runtime
# NOTE: If using 'scratch', you cannot use RUN/mkdir. You might need to use 'alpine' instead
# if you absolutely must create runtime directories within the image. 
# For true object storage, you'll mount external storage (a volume) anyway.


# Create uploads directory
RUN mkdir -p /app/store

# Expose port
EXPOSE 8280

# Set the entrypoint to run your compiled binary
ENTRYPOINT ["/go-api"]