#!/bin/sh
set -e

# Wait for the database to be ready
./wait-for-it.sh db ${DB_USER}

# Function to run migrations
run_migration() {
    local direction=$1
    echo "Running ${direction} migration..."
    /usr/local/bin/goose -dir ./sql/schema postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" $direction
}

# Check if a migration command was passed
if [ "$1" = "migrate" ]; then
    direction=${2:-up}  # Default to 'up' if no direction specified
    run_migration $direction
    exit 0
fi

# Run database migrations (up) by default
run_migration up

# Start the backup cron job
echo "0 2 * * * /app/backup.sh" | sudo crontab -

# Start the cron daemon
sudo crond

# Start the main application
exec ./main