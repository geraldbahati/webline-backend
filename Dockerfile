# Stage 1: Build the Go binary
FROM golang:latest AS builder

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

# Set the working directory to /app/cmd
WORKDIR /app/cmd

# Build the Go app
RUN go build -o /app/main .

# Stage 2: Run the binary
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the Goose binary from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy migration files
COPY --from=builder /app/sql/schema /app/sql/schema

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable and migrate the database
CMD ["sh", "-c", "/usr/local/bin/goose -dir /app/sql/schema postgres \"postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable\" up && ./main"]
