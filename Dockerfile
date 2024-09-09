# Stage 1: Build the Go binary
FROM golang:alpine AS builder

# Install PostgreSQL client utilities, bash, aws-cli, and cron
RUN apk add --no-cache postgresql-client bash aws-cli dcron

# Print Go version for reference
RUN go version

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install Goose for database migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Stage 2: Run the binary
FROM alpine:3.18

# Install PostgreSQL client utilities, bash, aws-cli, cron, and sudo
RUN apk add --no-cache postgresql-client bash aws-cli dcron sudo

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the Goose binary from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy migration files
COPY --from=builder /app/sql/schema ./sql/schema

# Copy wait-for-it script and backup script
COPY --from=builder /app/wait-for-it.sh .
COPY --from=builder /app/backup.sh .

# Copy entrypoint script
COPY entrypoint.sh .

# Make scripts executable
RUN chmod +x wait-for-it.sh backup.sh entrypoint.sh

# Create a non-root user and give it permissions to use crontab and crond
RUN adduser -D appuser && \
    echo "appuser ALL=(ALL) NOPASSWD: /usr/sbin/crond, /usr/bin/crontab" >> /etc/sudoers

# Change ownership of the /app directory to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port 8080 to the outside world
EXPOSE 8080

# Use the entrypoint script
ENTRYPOINT ["./entrypoint.sh"]
