#!/bin/bash

# Coffee Log Database Setup Script
# This script creates the MySQL database and user for the coffee-dex application

set -e  # Exit on error

echo "🚀 Setting up MySQL database for Coffee Log..."

# Database configuration
DB_NAME="coffee_log"
DB_USER="coffee_user"
DB_PASSWORD="coffee_pass123"
DB_HOST="localhost"
DB_PORT="3306"

# Check if MySQL is running
if ! mysqladmin ping -h"$DB_HOST" --silent; then
    echo "❌ MySQL is not running. Please start MySQL first:"
    echo "   brew services start mysql"
    exit 1
fi

echo "✅ MySQL is running"

# Create database and user
echo "📦 Creating database and user..."
mysql -h"$DB_HOST" -uroot <<EOF
-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS $DB_NAME;

-- Create user if it doesn't exist (MySQL 8.0+ syntax)
CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASSWORD';

-- Grant privileges
GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'localhost';
FLUSH PRIVILEGES;

-- Show databases
SHOW DATABASES;
EOF

echo ""
echo "✅ Database and user created!"
echo ""

# Create tables from schema
echo "📦 Creating database tables..."
if [ -f "../sql/schema.sql" ]; then
    mysql -h"$DB_HOST" -uroot "$DB_NAME" < ../sql/schema.sql
    echo "✅ Tables created from schema.sql"
elif [ -f "sql/schema.sql" ]; then
    mysql -h"$DB_HOST" -uroot "$DB_NAME" < sql/schema.sql
    echo "✅ Tables created from schema.sql"
else
    echo "⚠️  Warning: sql/schema.sql not found. Tables will be created by the application."
fi

echo ""
echo "✅ Database setup complete!"
echo ""
echo "📋 Connection Details:"
echo "   Database: $DB_NAME"
echo "   User:     $DB_USER"
echo "   Password: $DB_PASSWORD"
echo "   Host:     $DB_HOST:$DB_PORT"
echo ""
echo "🔧 To connect manually:"
echo "   mysql -u$DB_USER -p$DB_PASSWORD $DB_NAME"
echo ""
echo "🚀 To run the application with MySQL:"
echo "   go run main.go -storage=mysql -mysql-user=$DB_USER -mysql-password=$DB_PASSWORD -mysql-db=$DB_NAME"
echo ""