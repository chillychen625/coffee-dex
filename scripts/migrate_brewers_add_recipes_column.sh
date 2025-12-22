#!/bin/bash

# Migration script to add the 'recipes' JSON column to the brewers table
# This migrates from the old brewer_recipes join table approach to the new standalone recipes approach

MYSQL_HOST="${MYSQL_HOST:-localhost:3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD}"
MYSQL_DB="${MYSQL_DB:-coffee_log}"

# Extract host and port
HOST=$(echo $MYSQL_HOST | cut -d: -f1)
PORT=$(echo $MYSQL_HOST | cut -d: -f2)

echo "=========================================="
echo "Migrating brewers table"
echo "=========================================="
echo "Database: $MYSQL_DB"
echo "Host: $HOST:$PORT"
echo "User: $MYSQL_USER"
echo ""

# Check if column exists
COLUMN_EXISTS=$(mysql -h "$HOST" -P "$PORT" -u "$MYSQL_USER" ${MYSQL_PASSWORD:+-p"$MYSQL_PASSWORD"} -D "$MYSQL_DB" -sse \
  "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='$MYSQL_DB' AND TABLE_NAME='brewers' AND COLUMN_NAME='recipes'")

if [ "$COLUMN_EXISTS" -eq "0" ]; then
  echo "Adding 'recipes' JSON column to brewers table..."
  mysql -h "$HOST" -P "$PORT" -u "$MYSQL_USER" ${MYSQL_PASSWORD:+-p"$MYSQL_PASSWORD"} -D "$MYSQL_DB" -e \
    "ALTER TABLE brewers ADD COLUMN recipes JSON AFTER pokeball_type;"
  
  if [ $? -eq 0 ]; then
    echo "✅ Successfully added 'recipes' column"
  else
    echo "❌ Failed to add 'recipes' column"
    exit 1
  fi
else
  echo "✅ 'recipes' column already exists"
fi

echo ""
echo "=========================================="
echo "Migration complete!"
echo "=========================================="
echo ""
echo "Note: The old 'brewer_recipes' table is kept for backward compatibility"
echo "but is no longer used. You can drop it manually if desired:"
echo "  DROP TABLE IF EXISTS brewer_recipes;"