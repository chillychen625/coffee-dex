# MySQL Database Setup Guide for Go

## Understanding MySQL with Go

### How Go Connects to MySQL

```
Your Go App ──> database/sql (Go standard library)
                     │
                     ├──> MySQL Driver (github.com/go-sql-driver/mysql)
                     │
                     └──> MySQL Server (running on your computer/cloud)
```

## Step-by-Step Setup

### 1. Install MySQL

**macOS:**

```bash
brew install mysql
brew services start mysql
```

**Ubuntu/Debian:**

```bash
sudo apt-get update
sudo apt-get install mysql-server
sudo systemctl start mysql
```

**Windows:**
Download from https://dev.mysql.com/downloads/mysql/

### 2. Create Database

```bash
# Log into MySQL
mysql -u root -p

# Create database
CREATE DATABASE coffee_db;

# Create user (optional but recommended)
CREATE USER 'coffee_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON coffee_db.* TO 'coffee_user'@'localhost';
FLUSH PRIVILEGES;

# Exit
exit;
```

### 3. Add MySQL Driver to Your Go Project

```bash
# Add the driver dependency
go get github.com/go-sql-driver/mysql

# Update go.mod
go mod tidy
```

This will update your `go.mod` file:

```go
module go-coffee-log

go 1.21

require github.com/go-sql-driver/mysql v1.7.1
```

### 4. Update main.go to Use MySQL

```go
package main

import (
    "fmt"
    "go-coffee-log/handlers"
    "go-coffee-log/service"
    "go-coffee-log/storage"
    "log"
    "net/http"
)

func main() {
    // Option 1: Use In-Memory Storage (for testing)
    // var store storage.CoffeeStorage = storage.NewMemoryStorage()

    // Option 2: Use MySQL Storage (for production)
    store, err := storage.NewMySQLStorage(
        "localhost:3306",    // host
        "coffee_user",       // user
        "your_password",     // password
        "coffee_db",         // database name
    )
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer store.Close()  // Close database when main exits

    // Rest of your setup...
    coffeeService := service.NewCoffeeService(store)
    coffeeHandler := handlers.NewCoffeeHandler(coffeeService)

    mux := http.NewServeMux()
    mux.HandleFunc("POST /coffees", coffeeHandler.CreateCoffee)
    mux.HandleFunc("GET /coffees", coffeeHandler.ListCoffees)
    mux.HandleFunc("GET /coffees/{id}", coffeeHandler.GetCoffee)
    mux.HandleFunc("PUT /coffees/{id}", coffeeHandler.UpdateCoffee)
    mux.HandleFunc("DELETE /coffees/{id}", coffeeHandler.DeleteCoffee)

    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", loggingMiddleware(mux)))
}
```

## Key Concepts Explained

### 1. Database Connection String (DSN)

```go
dsn := "user:password@tcp(host)/dbname?parseTime=true"
//      ^^^^  ^^^^^^^^      ^^^^  ^^^^^^  ^^^^^^^^^^^^^
//      |     |             |     |       |
//      user  password      host  db      parse times to time.Time
```

**parseTime=true** is crucial! It tells the driver to convert MySQL DATETIME to Go's `time.Time`.

### 2. Prepared Statements (SQL Injection Prevention)

**BAD (Vulnerable to SQL Injection):**

```go
// NEVER DO THIS!
query := "INSERT INTO coffees (name) VALUES ('" + coffee.Name + "')"
db.Exec(query)
```

**GOOD (Safe with placeholders):**

```go
// Use ? placeholders
query := "INSERT INTO coffees (name, origin) VALUES (?, ?)"
db.Exec(query, coffee.Name, coffee.Origin)
```

The `?` placeholders are replaced safely by the driver, preventing injection attacks.

### 3. Querying Data

**Single Row:**

```go
row := db.QueryRow("SELECT * FROM coffees WHERE id = ?", id)
var coffee models.Coffee
err := row.Scan(&coffee.ID, &coffee.Name, &coffee.Origin, ...)
```

**Multiple Rows:**

```go
rows, err := db.Query("SELECT * FROM coffees")
defer rows.Close()  // Always close!

for rows.Next() {
    var coffee models.Coffee
    rows.Scan(&coffee.ID, &coffee.Name, ...)
    coffees = append(coffees, coffee)
}
```

### 4. Error Handling

```go
// Check if row was not found
if err == sql.ErrNoRows {
    return fmt.Errorf("coffee not found")
}

// Check if update affected any rows
result, err := db.Exec("UPDATE coffees SET name=? WHERE id=?", name, id)
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    return fmt.Errorf("no coffee with that ID")
}
```

## MySQL Data Types vs Go Types

| Go Type       | MySQL Type   | Example                         |
| ------------- | ------------ | ------------------------------- |
| string        | VARCHAR(255) | Name, Origin                    |
| string (long) | TEXT         | TastingNotes                    |
| int           | INT          | Rating                          |
| time.Time     | DATETIME     | CreatedAt, UpdatedAt            |
| bool          | BOOLEAN      | IsDecaf (if you add this field) |

## Common Patterns

### Pattern 1: Execute with No Return

```go
// INSERT, UPDATE, DELETE
_, err := db.Exec("DELETE FROM coffees WHERE id = ?", id)
```

### Pattern 2: Query Single Row

```go
// SELECT one row
row := db.QueryRow("SELECT * FROM coffees WHERE id = ?", id)
err := row.Scan(&coffee.ID, &coffee.Name, ...)
```

### Pattern 3: Query Multiple Rows

```go
// SELECT multiple rows
rows, err := db.Query("SELECT * FROM coffees")
defer rows.Close()
for rows.Next() {
    // Scan each row
}
```

## Testing Your MySQL Implementation

```bash
# 1. Start your server
go run main.go

# 2. Create a coffee
curl -X POST http://localhost:8080/coffees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Colombian Supremo",
    "origin": "Colombia",
    "roast_level": "Medium",
    "tasting_notes": "Nutty, caramel, smooth",
    "rating": 8,
    "brew_method": "Drip"
  }'

# 3. Check MySQL directly
mysql -u coffee_user -p coffee_db
SELECT * FROM coffees;
```

## Debugging Tips

### Enable MySQL Query Logging

```go
import _ "github.com/go-sql-driver/mysql"

// In NewMySQLStorage:
db, err := sql.Open("mysql", dsn)
db.SetMaxOpenConns(10)
db.SetMaxIdleConns(5)
```

### Check Connection

```go
err = db.Ping()
if err != nil {
    log.Fatal("Cannot reach database:", err)
}
log.Println("Successfully connected to MySQL!")
```

### View MySQL Errors

```bash
# Check MySQL error log
tail -f /usr/local/var/mysql/$(hostname).err  # macOS
tail -f /var/log/mysql/error.log              # Linux
```

## Environment Variables (Best Practice)

Instead of hardcoding credentials:

```go
import "os"

func main() {
    store, err := storage.NewMySQLStorage(
        os.Getenv("MYSQL_HOST"),
        os.Getenv("MYSQL_USER"),
        os.Getenv("MYSQL_PASSWORD"),
        os.Getenv("MYSQL_DATABASE"),
    )
}
```

Then run:

```bash
export MYSQL_HOST="localhost:3306"
export MYSQL_USER="coffee_user"
export MYSQL_PASSWORD="your_password"
export MYSQL_DATABASE="coffee_db"
go run main.go
```

## Next Steps

1. Implement the TODOs in `storage/mysql.go`
2. Test with in-memory storage first
3. Switch to MySQL once in-memory works
4. Add indexes for better performance:
   ```sql
   CREATE INDEX idx_rating ON coffees(rating);
   CREATE INDEX idx_origin ON coffees(origin);
   ```

## Comparison: Memory vs MySQL

| Feature           | Memory Storage      | MySQL Storage          |
| ----------------- | ------------------- | ---------------------- |
| Persistence       | ❌ Lost on restart  | ✅ Permanent           |
| Speed             | ✅ Very fast        | ⚡ Fast (with indexes) |
| Concurrent Access | ⚠️ Needs mutex      | ✅ Built-in            |
| Scalability       | ❌ Limited by RAM   | ✅ Can handle millions |
| Good for          | Development/Testing | Production             |

Good luck! 🚀
