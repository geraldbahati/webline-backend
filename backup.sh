#!/bin/bash
set -e

# Logging helpers
log_info() {
  echo -e "\033[0;34m[INFO]\033[0m $1"
}

log_success() {
  echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

log_warn() {
  echo -e "\033[0;33m[WARNING]\033[0m $1"
}

log_error() {
  echo -e "\033[0;31m[ERROR]\033[0m $1"
}

# Set the backup directory
BACKUP_DIR="/backups"
BACKUP_RETENTION_DAYS=7
MAX_S3_BACKUPS=30

# Ensure backup directories exist
mkdir -p "$BACKUP_DIR"/{database,env,migrations}

# Get the current date with timezone information
DATE=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="webline_backup_${DATE}"

# Database connection validation
if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
  log_error "Database connection parameters are not properly set"
  exit 1
fi

# Check database connection
log_info "Checking database connection..."
if ! PGPASSWORD=$DB_PASSWORD pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t 5; then
  log_error "Could not connect to the database. Please check your connection parameters."
  exit 1
fi

# Database backup
log_info "Creating database backup..."
DB_BACKUP_FILE="${BACKUP_DIR}/database/${DB_NAME}_${DATE}.dump"
if ! PGPASSWORD=$DB_PASSWORD pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -F c -f "$DB_BACKUP_FILE" "$DB_NAME"; then
  log_error "Database backup failed!"
  exit 1
fi
log_success "Database backup completed: $DB_BACKUP_FILE"

# Create a compressed version for S3 upload
log_info "Compressing backup for S3..."
S3_BACKUP_FILE="${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
tar -czf "$S3_BACKUP_FILE" -C "$BACKUP_DIR/database" "$(basename "$DB_BACKUP_FILE")" -C "$BACKUP_DIR/env" "$(basename "$ENV_BACKUP_FILE")" || {
  log_warn "Compression partially failed but continuing..."
}

# Environment variables backup
log_info "Creating environment variables backup..."
ENV_BACKUP_FILE="${BACKUP_DIR}/env/env_vars_${DATE}.txt"
env | grep -v "PASSWORD\|SECRET\|KEY" > "$ENV_BACKUP_FILE" # Exclude sensitive info
echo "# Sensitive variables have been redacted for security" >> "$ENV_BACKUP_FILE"
log_success "Environment backup completed: $ENV_BACKUP_FILE"

# Save migration version status
log_info "Saving migration status..."
MIGRATION_BACKUP_FILE="${BACKUP_DIR}/migrations/migration_status_${DATE}.txt"
if [ -x "$(command -v goose)" ]; then
  goose -dir ./sql/schema postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" status > "$MIGRATION_BACKUP_FILE" || {
    log_warn "Could not get migration status"
  }
  log_success "Migration status saved: $MIGRATION_BACKUP_FILE"
else
  log_warn "Goose not available, skipping migration status"
fi

# Cleanup old backups
log_info "Cleaning up old backups (keeping last $BACKUP_RETENTION_DAYS days)..."
find "$BACKUP_DIR/database" -type f -mtime +$BACKUP_RETENTION_DAYS -delete
find "$BACKUP_DIR/env" -type f -mtime +$BACKUP_RETENTION_DAYS -delete
find "$BACKUP_DIR/migrations" -type f -mtime +$BACKUP_RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "webline_backup_*.tar.gz" -type f -mtime +$BACKUP_RETENTION_DAYS -delete

# Backup to AWS S3 if credentials are available
if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ] && [ -n "$AWS_BUCKET_NAME" ]; then
  log_info "Uploading backup to S3 bucket: $AWS_BUCKET_NAME..."

  # Upload the compressed backup
  if aws s3 cp "$S3_BACKUP_FILE" "s3://$AWS_BUCKET_NAME/backups/$(basename "$S3_BACKUP_FILE")"; then
    log_success "Backup uploaded to S3 successfully"

    # Clean up old S3 backups
    log_info "Cleaning up old S3 backups..."
    S3_BACKUP_LIST=$(aws s3 ls "s3://$AWS_BUCKET_NAME/backups/" | grep "webline_backup_" | sort -r)
    S3_BACKUP_COUNT=$(echo "$S3_BACKUP_LIST" | wc -l)

    if [ "$S3_BACKUP_COUNT" -gt "$MAX_S3_BACKUPS" ]; then
      log_info "S3 has $S3_BACKUP_COUNT backups, keeping only $MAX_S3_BACKUPS"
      TO_DELETE=$(echo "$S3_BACKUP_LIST" | tail -n $(($S3_BACKUP_COUNT - $MAX_S3_BACKUPS)) | awk '{print $4}')

      for file in $TO_DELETE; do
        log_info "Removing old S3 backup: $file"
        aws s3 rm "s3://$AWS_BUCKET_NAME/backups/$file"
      done
    fi
  else
    log_error "Failed to upload backup to S3"
  fi
else
  log_warn "AWS credentials not found. Skipping S3 backup."
fi

# Print backup summary
log_success "Backup process completed successfully:"
log_info "Database: $DB_BACKUP_FILE"
log_info "Environment: $ENV_BACKUP_FILE"
log_info "Migration Status: $MIGRATION_BACKUP_FILE"
log_info "Compressed Backup: $S3_BACKUP_FILE"
