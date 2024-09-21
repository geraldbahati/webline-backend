#!/bin/bash

# Set the backup directory
BACKUP_DIR=/backups

# Get the current date
DATE=$(date +"%Y%m%d_%H%M%S")

# Create backup directories
mkdir -p $BACKUP_DIR/{database,env}

# Database backup
DB_BACKUP_FILE=$BACKUP_DIR/database/${DB_NAME}_$DATE.sql.gz
PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -F c $DB_NAME | gzip > $DB_BACKUP_FILE

# Environment variables backup
ENV_BACKUP_FILE=$BACKUP_DIR/env/env_vars_$DATE.txt
env > $ENV_BACKUP_FILE

# Cleanup old backups (keep last 7 days)
find $BACKUP_DIR -type f -mtime +7 -delete

# Backup to AWS S3 if credentials are available
if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ] && [ -n "$AWS_BUCKET_NAME" ]; then
    aws s3 sync $BACKUP_DIR s3://$AWS_BUCKET_NAME/backups
    echo "Backup synced to S3 bucket: $AWS_BUCKET_NAME"
else
    echo "AWS credentials not found. Skipping S3 backup."
fi

echo "Backup completed:"
echo "Database: $DB_BACKUP_FILE"
echo "Environment: $ENV_BACKUP_FILE"
