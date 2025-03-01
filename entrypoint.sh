#!/bin/bash
set -e

# Printing helper functions with standardized formats and timestamps
log_info() {
  echo -e "\033[0;34m[INFO][$(date '+%Y-%m-%d %H:%M:%S')]\033[0m $1"
}

log_success() {
  echo -e "\033[0;32m[SUCCESS][$(date '+%Y-%m-%d %H:%M:%S')]\033[0m $1"
}

log_warn() {
  echo -e "\033[0;33m[WARNING][$(date '+%Y-%m-%d %H:%M:%S')]\033[0m $1"
}

log_error() {
  echo -e "\033[0;31m[ERROR][$(date '+%Y-%m-%d %H:%M:%S')]\033[0m $1"
}

# Function to check migration status with retry
check_migrations() {
  local retry_count=0
  local max_retries=3

  log_info "Checking migration status..."

  while [ $retry_count -lt $max_retries ]; do
    if /usr/local/bin/goose -dir ./sql/schema postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" status; then
      return 0
    else
      retry_count=$((retry_count+1))
      log_warn "Failed to check migration status (attempt $retry_count/$max_retries). Retrying in 5s..."
      sleep 5
    fi
  done

  log_error "Failed to check migration status after $max_retries attempts."
  return 1
}

# Function to run migrations with improved error handling and built-in retries
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

    local mode=${MIGRATION_MODE:-safe}
    local retry_count=0
    local max_retries=3
    local backup_path=""

    log_info "Running ${direction} migration in ${mode} mode..."

    # Ensure the migration directory exists
    if [ ! -d "./sql/schema" ]; then
        log_error "Migration directory not found!"
        return 1
    fi

    # Backup the database before migrations if mode is safe
    if [ "$mode" = "safe" ] && [ "$direction" = "up" ]; then
        log_info "Creating database backup before migrations..."
        TIMESTAMP=$(date +%Y%m%d_%H%M%S)
        mkdir -p /backups
        backup_path="/backups/pre_migration_${TIMESTAMP}.dump"

        PGPASSWORD=${DB_PASSWORD} pg_dump -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -F c -f "$backup_path" || {
            log_error "Failed to create database backup. Aborting migrations."
            return 1
        }
        log_success "Database backup created at $backup_path"
    fi

    # Check for pending migrations before proceeding
    check_migrations
    local pending_migrations=$(/usr/local/bin/goose -dir ./sql/schema postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" status | grep -c 'Pending' || echo "0")

    if [ "$pending_migrations" -eq 0 ] && [ "$direction" = "up" ]; then
        log_info "No pending migrations. Database is up to date."
        return 0
    fi

    # Run actual migration with retries
    while [ $retry_count -lt $max_retries ]; do
        if /usr/local/bin/goose -dir ./sql/schema postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" $direction; then
            log_success "Migration ${direction} completed successfully"
            return 0
        else
            local exit_code=$?
            retry_count=$((retry_count+1))

            if [ $retry_count -lt $max_retries ]; then
                log_warn "Migration ${direction} failed (attempt $retry_count/$max_retries). Retrying in 5s..."
                sleep 5
            else
                log_error "Migration ${direction} failed with exit code $exit_code after $max_retries attempts"

                # Attempt recovery if mode is safe and we were doing 'up' migrations
                if [ "$mode" = "safe" ] && [ "$direction" = "up" ] && [ -f "$backup_path" ]; then
                    log_warn "Attempting to restore database from backup..."
                    PGPASSWORD=${DB_PASSWORD} pg_restore -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} --clean --if-exists "$backup_path" || {
                        log_error "Failed to restore database backup. Manual intervention required."
                        return 2
                    }
                    log_success "Database restored from backup"
                fi

                return $exit_code
            fi
        fi
    done
}

# Setup health check function
health_check() {
    if [ -z "${HEALTH_CHECK_URL}" ]; then
        # No health check URL defined, considered healthy
        return 0
    fi

    curl --silent --fail "${HEALTH_CHECK_URL}" > /dev/null
    return $?
}

# Wait for PostgreSQL to be ready with improved health check
wait_for_postgres() {
    log_info "Waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}..."
    ./wait-for-it.sh ${DB_HOST}:${DB_PORT} --timeout=60 --strict || {
        log_error "PostgreSQL is not available after 60 seconds. Exiting."
        return 1
    }

    # Additional connection test
    for i in {1..5}; do
        if PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1" >/dev/null 2>&1; then
            log_success "PostgreSQL connection test successful"
            return 0
        fi
        log_warn "PostgreSQL connection test failed (attempt $i/5). Retrying in 5s..."
        sleep 5
    done

    log_error "Failed to connect to PostgreSQL after 5 attempts"
    return 1
}

# Wait for Redis to be ready with improved health check
wait_for_redis() {
    log_info "Waiting for Redis at ${REDIS_HOST}:${REDIS_PORT}..."
    ./wait-for-it.sh ${REDIS_HOST}:${REDIS_PORT} --timeout=30 --strict || {
        log_error "Redis is not available after 30 seconds. Exiting."
        return 1
    }

    # Additional connection test
    for i in {1..5}; do
        if echo "AUTH ${REDIS_PASSWORD}" | redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} | grep -q "OK" && \
           echo "PING" | redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} -a "${REDIS_PASSWORD}" | grep -q "PONG"; then
            log_success "Redis connection test successful"
            return 0
        fi
        log_warn "Redis connection test failed (attempt $i/5). Retrying in 5s..."
        sleep 5
    done

    log_error "Failed to connect to Redis after 5 attempts"
    return 1
}

# Ensure clean shutdown
cleanup() {
    log_info "Performing cleanup..."

    # Create a flag file to signal shutdown to child processes
    touch /tmp/shutdown_in_progress

    # Stop the application
    if [ -n "$APP_PID" ] && kill -0 $APP_PID 2>/dev/null; then
        log_info "Sending SIGTERM to application (PID: $APP_PID)"
        kill -TERM $APP_PID 2>/dev/null || true

        # Wait for application to terminate gracefully
        for i in {1..30}; do
            if ! kill -0 $APP_PID 2>/dev/null; then
                break
            fi
            sleep 1
        done

        # Force kill if still running
        if kill -0 $APP_PID 2>/dev/null; then
            log_warn "Application did not shut down gracefully. Sending SIGKILL."
            kill -KILL $APP_PID 2>/dev/null || true
        fi
    fi

    # Stop cron daemon
    if [ -n "$CRON_PID" ] && kill -0 $CRON_PID 2>/dev/null; then
        log_info "Stopping cron daemon (PID: $CRON_PID)"
        kill -TERM $CRON_PID 2>/dev/null || true
    fi

    log_success "Cleanup completed"
}

# Initialize counters for retry logic
failed_starts=0
max_failed_starts=3

# Main execution
main() {
    # Handle signals for graceful shutdown
    trap cleanup SIGTERM SIGINT

    # Wait for dependencies
    wait_for_postgres || exit 1
    wait_for_redis || exit 1

    # Handle commands
    if [ "$1" = "migrate" ]; then
        direction=${2:-up}
        run_migration $direction
        exit $?
    elif [ "$1" = "backup" ]; then
        log_info "Running database backup..."
        ./backup.sh
        exit $?
    elif [ "$1" = "shell" ]; then
        log_info "Starting shell..."
        /bin/bash
        exit 0
    else
        # Run migrations if AUTO_MIGRATE is enabled (default: true)
        if [ "${AUTO_MIGRATE:-true}" = "true" ]; then
            run_migration up || {
                log_error "Migration failed. Exiting."
                exit 1
            }
        else
            log_info "Automatic migrations disabled"
        fi
    fi

    # Start cron in the background with proper error handling
    log_info "Starting cron daemon..."
    crond -f &
    CRON_PID=$!

    # Check if cron started properly
    if ! kill -0 $CRON_PID 2>/dev/null; then
        log_error "Failed to start cron daemon"
        CRON_PID=""
    else
        log_success "Cron daemon started with PID $CRON_PID"
    fi

    # Start the application with restart capability
    while [ $failed_starts -lt $max_failed_starts ]; do
        log_info "Starting application..."

        # Start the application
        ./main &
        APP_PID=$!

        # Wait for application to start
        log_info "Waiting for application to initialize (PID: $APP_PID)..."
        sleep 5

        # If we're shutting down, break the loop
        if [ -f /tmp/shutdown_in_progress ]; then
            break
        fi

        # Wait for the application to exit
        wait $APP_PID || true
        app_exit_code=$?

        # Check if we're shutting down
        if [ -f /tmp/shutdown_in_progress ]; then
            break
        fi

        # Handle application exit
        if [ $app_exit_code -ne 0 ]; then
            failed_starts=$((failed_starts+1))
            log_error "Application exited with error code $app_exit_code (attempt $failed_starts/$max_failed_starts)"

            if [ $failed_starts -lt $max_failed_starts ]; then
                log_info "Restarting application in 5 seconds..."
                sleep 5
            else
                log_error "Application failed to start after $max_failed_starts attempts. Giving up."
                cleanup
                exit 1
            fi
        else
            log_info "Application exited normally with code $app_exit_code"
            break
        fi
    done

    # Wait for cron to finish (should only happen on container shutdown)
    if [ -n "$CRON_PID" ]; then
        wait $CRON_PID || true
    fi

    log_info "Container stopping..."
}

# Execute main function
main "$@"
