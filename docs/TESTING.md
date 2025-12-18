# Coffee-Dex Database Testing Guide

This guide explains how to set up and test the MySQL database for the coffee-dex application.

## Prerequisites

- macOS
- Homebrew
- Go 1.21+
- MySQL (installed via the setup script)

## Quick Start

### 1. Install and Setup MySQL

MySQL has been installed via Homebrew. To set up the database:

```bash
./scripts/setup_database.sh
```

This script will:

- Check if MySQL is running
- Create the `coffee_log` database
- Create a user `coffee_user` with password `coffee_pass123`
- Grant necessary privileges

### 2. Run Database Tests

Run the comprehensive test suite:

```bash
./scripts/test_database.sh
```

Or run tests manually:

```bash
# Run all MySQL tests
go test -v ./storage -run TestMySQL

# Run with coverage
go test -v -cover ./storage -run TestMySQL

# Run a specific test
go test -v ./storage -run TestSaveAndGetByID
```

### 3. Populate Test Data

Generate sample coffee entries in the database:

```bash
# Create all 6 sample entries
go run scripts/populate_test_data.go

# Create only 3 entries
go run scripts/populate_test_data.go -count=3

# Use custom MySQL credentials
go run scripts/populate_test_data.go -mysql-user=root -mysql-password=mypass
```

## Available Test Cases

The test suite ([`storage/mysql_test.go`](storage/mysql_test.go)) includes:

1. **TestNewMySQLStorage** - Tests database connection
2. **TestSaveAndGetByID** - Tests saving and retrieving a single coffee
3. **TestGetAll** - Tests retrieving all coffees
4. **TestUpdate** - Tests updating a coffee entry
5. **TestDelete** - Tests deleting a coffee entry
6. **TestGetByIDNotFound** - Tests error handling for non-existent entries
7. **TestUpdateNotFound** - Tests error handling for updating non-existent entries
8. **TestDeleteNotFound** - Tests error handling for deleting non-existent entries
9. **TestJSONMarshaling** - Tests proper JSON field storage and retrieval

## Database Configuration

Default configuration:

- **Host:** `localhost:3306`
- **Database:** `coffee_log`
- **User:** `coffee_user`
- **Password:** `coffee_pass123`

You can customize these via command-line flags when running the application:

```bash
go run main.go \
  -storage=mysql \
  -mysql-host=localhost:3306 \
  -mysql-user=coffee_user \
  -mysql-password=coffee_pass123 \
  -mysql-db=coffee_log
```

## Manual Testing

### Connect to MySQL

```bash
mysql -ucoffee_user -pcoffee_pass123 coffee_log
```

### Useful SQL Commands

```sql
-- View all coffees
SELECT * FROM coffees;

-- Count entries
SELECT COUNT(*) FROM coffees;

-- View specific coffee
SELECT * FROM coffees WHERE name LIKE '%Ethiopian%';

-- Delete all entries (careful!)
DELETE FROM coffees;

-- Drop and recreate table
DROP TABLE coffees;
```

Then restart your application to recreate the table automatically.

## Testing with the HTTP API

Once you've populated test data, you can test the API:

```bash
# Start the server with MySQL
go run main.go -storage=mysql -mysql-user=coffee_user -mysql-password=coffee_pass123

# In another terminal, test the endpoints:

# Get all coffees
curl http://localhost:8080/coffees

# Get specific coffee (replace ID)
curl http://localhost:8080/coffees/{id}

# Create a new coffee
curl -X POST http://localhost:8080/coffees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Coffee",
    "origin": "Ethiopia",
    "roaster": "Local Roaster",
    "roast_level": "light",
    "processing_method": "washed",
    "tasting_notes": ["chocolate", "berry", "floral", "", ""],
    "tasting_traits": {
      "berry_intensity": 7,
      "stonefruit_intensity": 3,
      "roast_intensity": 2,
      "citrus_fruits_intensity": 6,
      "bitterness": 2,
      "florality": 8,
      "spice": 1,
      "sweetness": 7,
      "aromatic_intensity": 8,
      "savory": 1,
      "body": 5,
      "cleanliness": 9
    },
    "rating": 8,
    "recipe": ["20g coffee", "320ml water", "94°C"],
    "dripper": "V60",
    "end_time": {
      "minutes": 2,
      "seconds": 45
    }
  }'
```

## Troubleshooting

### MySQL Not Running

```bash
# Start MySQL
brew services start mysql

# Check status
brew services list | grep mysql

# Check if MySQL is responding
mysqladmin ping
```

### Database Connection Errors

1. Verify MySQL is running: `mysqladmin ping`
2. Check credentials: `mysql -ucoffee_user -pcoffee_pass123`
3. Verify database exists: `mysql -uroot -e "SHOW DATABASES;"`
4. Re-run setup script: `./scripts/setup_database.sh`

### Test Failures

- Ensure database is set up: `./scripts/setup_database.sh`
- Clear test data: `mysql -ucoffee_user -pcoffee_pass123 coffee_log -e "DELETE FROM coffees;"`
- Check MySQL logs: `tail -f /opt/homebrew/var/mysql/*.err`

### Permission Issues

If you get permission errors with scripts:

```bash
chmod +x scripts/*.sh
```

## Database Schema

The `coffees` table schema:

```sql
CREATE TABLE coffees (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    origin VARCHAR(255),
    roaster VARCHAR(255),
    roast_level VARCHAR(50),
    processing_method VARCHAR(100),
    tasting_notes JSON,
    tasting_traits JSON,
    rating INT,
    recipe JSON,
    dripper VARCHAR(100),
    end_time_minutes INT,
    end_time_seconds INT,
    created_at DATETIME,
    updated_at DATETIME
);
```

## Cleaning Up

To stop MySQL:

```bash
brew services stop mysql
```

To remove the database:

```bash
mysql -uroot -e "DROP DATABASE coffee_log;"
```

To uninstall MySQL completely:

```bash
brew services stop mysql
brew uninstall mysql
rm -rf /opt/homebrew/var/mysql
```
