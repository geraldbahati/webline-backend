# Stage 1: Build the Go binary
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache --update \
    build-base \
    postgresql-client \
    bash \
    aws-cli \
    dcron

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install Goose for database migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy the source code to the Working Directory
COPY . .

# Build the Go app with platform-specific settings
ARG TARGETPLATFORM
RUN case "${TARGETPLATFORM}" in \
      "linux/amd64") GOARCH=amd64 ;; \
      "linux/arm64") GOARCH=arm64 ;; \
      *) GOARCH=amd64 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -a -installsuffix cgo -o main ./cmd

# Stage 2: Run the binary
FROM --platform=$TARGETPLATFORM alpine:latest

# Install necessary runtime dependencies, including redis-cli and dcron
RUN apk add --no-cache --update \
    postgresql-client \
    bash \
    aws-cli \
    dcron \
    redis \
    shadow  # Install shadow to get 'su'

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the Goose binary from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy migration files
COPY --from=builder /app/sql/schema ./sql/schema

# Copy necessary scripts
COPY --from=builder /app/wait-for-it.sh .
COPY --from=builder /app/backup.sh .

# Copy the crontab file
COPY crontab-appuser /etc/crontabs/appuser

# Create the 'crontab' group, set permissions, and change ownership
RUN addgroup -S crontab && \
    chmod 600 /etc/crontabs/appuser && \
    chown root:crontab /etc/crontabs/appuser

# Copy entrypoint script
COPY entrypoint.sh .

# Make scripts executable
RUN chmod +x wait-for-it.sh backup.sh entrypoint.sh

# Create a non-root user and set ownership
RUN adduser -D appuser && \
    chown -R appuser:appuser /app

# # Switch to the non-root user
# USER appuser

# Expose port 8080 to the outside world
EXPOSE 8080

# Use the entrypoint script
ENTRYPOINT ["./entrypoint.sh"]
