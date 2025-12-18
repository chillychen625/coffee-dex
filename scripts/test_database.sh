#!/bin/bash

# Coffee Log Database Test Runner
# This script runs all database tests and provides detailed output

set -e

echo "☕ Coffee Log Database Test Suite"
echo "=================================="
echo ""

# Check if MySQL is running
if ! mysqladmin ping -h"localhost" --silent; then
    echo "❌ MySQL is not running. Please start MySQL first:"
    echo "   brew services start mysql"
    exit 1
fi

echo "✅ MySQL is running"

# Check if database exists
if ! mysql -uroot -e "USE coffee_log" 2>/dev/null; then
    echo "⚠️  Database 'coffee_log' not found. Running setup script..."
    ./scripts/setup_database.sh
fi

echo ""
echo "🧪 Running Go tests..."
echo ""

# Run tests with verbose output
go test -v ./storage -run TestMySQL

echo ""
echo "📊 Running tests with coverage..."
echo ""

# Run tests with coverage
go test -v -cover ./storage -run TestMySQL

echo ""
echo "✅ All tests completed!"
echo ""