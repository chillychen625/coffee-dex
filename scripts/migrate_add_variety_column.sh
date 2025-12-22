#!/bin/bash

# Migration script to add variety column to coffees table
# Run this script to add the variety field to existing databases

echo "Adding variety column to coffees table..."

# Get MySQL credentials from environment or use defaults
MYSQL_USER=${MYSQL_USER:-coffee_user}
MYSQL_PASSWORD=${MYSQL_PASSWORD:-coffee_pass123}
MYSQL_HOST=${MYSQL_HOST:-localhost}
MYSQL_PORT=${MYSQL_PORT:-3306}
MYSQL_DATABASE=${MYSQL_DATABASE:-coffee_log}

# Add variety column
mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -h"$MYSQL_HOST" -P"$MYSQL_PORT" "$MYSQL_DATABASE" <<EOF
ALTER TABLE coffees ADD COLUMN variety VARCHAR(255) AFTER roaster;
EOF

if [ $? -eq 0 ]; then
    echo "✓ Successfully added variety column to coffees table"
else
    echo "✗ Failed to add variety column (it may already exist)"
    exit 1
fi

echo "Migration complete!"