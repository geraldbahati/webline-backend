#!/bin/sh
set -e

# Wait for the database to be ready
./wait-for-it.sh db ${DB_USER}

# Run database migrations
/usr/local/bin/goose -dir ./sql/schema postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" up

# Start the backup cron job
echo "0 2 * * * /app/backup.sh" | sudo crontab -

# Start the cron daemon
sudo crond

# Start the main application
exec ./main