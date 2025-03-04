# Stage 1: Build the Go binary
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

# Install build dependencies - only what we need for building
RUN apk add --no-cache --update \
    build-base \
    postgresql-client \
    bash

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files first for better layer caching
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install Goose for database migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy only the necessary source code to the Working Directory
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./sql ./sql

# Build the Go app with platform-specific settings
ARG TARGETPLATFORM
ARG VERSION
ARG BUILD_DATE

# Add build info to binary
RUN echo "Building for $TARGETPLATFORM with version $VERSION on $BUILD_DATE" && \
    case "${TARGETPLATFORM}" in \
      "linux/amd64") GOARCH=amd64 ;; \
      "linux/arm64") GOARCH=arm64 ;; \
      *) GOARCH=amd64 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} \
    go build -ldflags="-w -s -X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE" \
    -a -installsuffix cgo -o main ./cmd

# Stage 2: Create a minimal runtime image
FROM --platform=$TARGETPLATFORM alpine:latest

# Add security labels
LABEL org.opencontainers.image.vendor="Webline Backend" \
      org.opencontainers.image.title="Webline Backend" \
      org.opencontainers.image.description="Webline Backend Service" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}"

# Install necessary runtime dependencies only
RUN apk add --no-cache --update \
    postgresql-client \
    bash \
    aws-cli \
    dcron \
    redis \
    shadow \
    curl && \
    rm -rf /var/cache/apk/*

# Set the Current Working Directory inside the container
WORKDIR /app

# Create a non-root user to run our application
RUN addgroup -S appgroup && \
    adduser -S -G appgroup appuser && \
    mkdir -p /app/data /backups && \
    chown -R appuser:appgroup /app /backups

# Copy the Pre-built binary from the previous stage
COPY --from=builder --chown=appuser:appgroup /app/main .

# Copy the Goose binary from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy migration files
COPY --from=builder --chown=appuser:appgroup /app/sql/schema ./sql/schema

# Copy necessary configuration files and scripts
COPY --chown=appuser:appgroup ./wait-for-it.sh .
COPY --chown=appuser:appgroup ./backup.sh .
COPY --chown=appuser:appgroup ./entrypoint.sh .
COPY --chown=root:crontab ./crontab-appuser /etc/crontabs/appuser

# Make scripts executable
RUN chmod +x wait-for-it.sh backup.sh entrypoint.sh && \
    # Create the 'crontab' group if it doesn't exist
    getent group crontab || addgroup -S crontab && \
    # Set correct crontab permissions
    chmod 600 /etc/crontabs/appuser && \
    chown root:crontab /etc/crontabs/appuser

# Create volume mount points
VOLUME ["/app/data", "/backups"]

# Expose port 8080 to the outside world
EXPOSE 8080

# Switch to the non-root user for the actual container execution
USER appuser

# Use the entrypoint script
ENTRYPOINT ["/bin/bash", "./entrypoint.sh"]

# Default command if nothing is provided
CMD []
