#!/bin/sh
set -e

# Function to run migrations
run_migration() {
    local direction=$1
    echo "Running ${direction} migration..."
    /usr/local/bin/goose -dir ./sql/schema postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" $direction
}

# Wait for PostgreSQL to be ready
./wait-for-it.sh db:5432 --timeout=30 --strict -- echo "PostgreSQL is up"

# Wait for Redis to be ready
./wait-for-it.sh redis:6379 --timeout=30 --strict -- echo "Redis is up"

echo "Redis is up. Starting application..."

# Run migrations (up) by default or based on command
if [ "$1" = "migrate" ]; then
    direction=${2:-up}
    run_migration $direction
    exit 0
else
    run_migration up
fi

# Start the cron daemon in the background as root
crond -f &
CRON_PID=$!

# Start the main application as 'appuser' in the background
su appuser -c "./main" &
APP_PID=$!

# Function to handle termination signals
terminate() {
    echo "Stopping cron daemon..."
    kill $CRON_PID
    echo "Stopping main application..."
    kill $APP_PID
    exit 0
}

# Trap termination signals to gracefully shut down
trap terminate SIGTERM SIGINT

# Wait for both processes
wait $CRON_PID
wait $APP_PID
