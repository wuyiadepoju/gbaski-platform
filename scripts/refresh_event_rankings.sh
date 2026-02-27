#!/bin/bash

# Log file for the cron job
LOG_FILE="/var/log/gbaski/refresh_event_rankings.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# Ensure log directory exists
mkdir -p "$(dirname "$LOG_FILE")"

# Database connection parameters
DB_NAME="gbaski"
DB_USER="postgres"
DB_HOST="localhost"
DB_PASSWORD="t6ePyUlVHSAPSc3VO2iS2lKL44Gy"  # Replace with actual password

# Log start of refresh
echo "[$TIMESTAMP] Starting event rankings refresh" >> "$LOG_FILE"

# Execute the refresh command with password
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "REFRESH MATERIALIZED VIEW CONCURRENTLY event_rankings;" >> "$LOG_FILE" 2>&1

# Log completion and status
if [ $? -eq 0 ]; then
    echo "[$TIMESTAMP] Successfully refreshed event rankings" >> "$LOG_FILE"
else
    echo "[$TIMESTAMP] Failed to refresh event rankings" >> "$LOG_FILE"
fi 