# Quick MySQL Reference for Go

## Essential Patterns You'll Use

### 1. Opening a Connection

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

// Connection string format
dsn := "user:password@tcp(host:port)/database?parseTime=true"

// Open connection
db, err := sql.Open("mysql", dsn)
if err != nil {
    return err
}

// Test connection
if err := db.Ping(); err != nil {
    return err
}
```

### 2. Creating a Table

```go
query := `
CREATE TABLE IF NOT EXISTS coffees (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    origin VARCHAR(255),
    roast_level VARCHAR(50),
    tasting_notes TEXT,
    rating INT,
    brew_method VARCHAR(100),
    created_at DATETIME,
    updated_at DATETIME
)`

_, err := db.Exec(query)
```

### 3. INSERT (Save)

```go
query := `
INSERT INTO coffees (id, name, origin, roast_level, tasting_notes, rating, brew_method, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

_, err := db.Exec(
    query,
    coffee.ID,
    coffee.Name,
    coffee.Origin,
    coffee.RoastLevel,
    coffee.TastingNotes,
    coffee.Rating,
    coffee.BrewMethod,
    coffee.CreatedAt,
    coffee.UpdatedAt,
)
```

### 4. SELECT Single Row (GetByID)

```go
query := `
SELECT id, name, origin, roast_level, tasting_notes, rating, brew_method, created_at, updated_at
FROM coffees
WHERE id = ?`

var coffee models.Coffee
row := db.QueryRow(query, id)
err := row.Scan(
    &coffee.ID,
    &coffee.Name,
    &coffee.Origin,
    &coffee.RoastLevel,
    &coffee.TastingNotes,
    &coffee.Rating,
    &coffee.BrewMethod,
    &coffee.CreatedAt,
    &coffee.UpdatedAt,
)

if err == sql.ErrNoRows {
    return models.Coffee{}, fmt.Errorf("coffee not found")
}
```

### 5. SELECT Multiple Rows (GetAll)

```go
query := `
SELECT id, name, origin, roast_level, tasting_notes, rating, brew_method, created_at, updated_at
FROM coffees`

rows, err := db.Query(query)
if err != nil {
    return nil, err
}
defer rows.Close()  // IMPORTANT: Always close!

var coffees []models.Coffee
for rows.Next() {
    var coffee models.Coffee
    err := rows.Scan(
        &coffee.ID,
        &coffee.Name,
        &coffee.Origin,
        &coffee.RoastLevel,
        &coffee.TastingNotes,
        &coffee.Rating,
        &coffee.BrewMethod,
        &coffee.CreatedAt,
        &coffee.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    coffees = append(coffees, coffee)
}

// Check for errors during iteration
if err = rows.Err(); err != nil {
    return nil, err
}

return coffees, nil
```

### 6. UPDATE

```go
query := `
UPDATE coffees
SET name=?, origin=?, roast_level=?, tasting_notes=?, rating=?, brew_method=?, updated_at=?
WHERE id=?`

result, err := db.Exec(
    query,
    coffee.Name,
    coffee.Origin,
    coffee.RoastLevel,
    coffee.TastingNotes,
    coffee.Rating,
    coffee.BrewMethod,
    coffee.UpdatedAt,
    id,
)

if err != nil {
    return err
}

// Check if any row was actually updated
rowsAffected, err := result.RowsAffected()
if err != nil {
    return err
}
if rowsAffected == 0 {
    return fmt.Errorf("coffee not found")
}
```

### 7. DELETE

```go
query := "DELETE FROM coffees WHERE id = ?"
result, err := db.Exec(query, id)

if err != nil {
    return err
}

rowsAffected, err := result.RowsAffected()
if err != nil {
    return err
}
if rowsAffected == 0 {
    return fmt.Errorf("coffee not found")
}
```

### 8. Closing Connection

```go
func (m *MySQLStorage) Close() error {
    return m.db.Close()
}
```

## Common Mistakes to Avoid

### ❌ Forgetting to Close Rows

```go
rows, _ := db.Query("SELECT * FROM coffees")
// Missing: defer rows.Close()
```

### ❌ Not Checking rows.Err()

```go
for rows.Next() {
    // scan...
}
// Missing: if err := rows.Err(); err != nil { return err }
```

### ❌ SQL Injection

```go
// NEVER DO THIS:
query := "SELECT * FROM coffees WHERE id = '" + id + "'"
```

### ✅ Use Placeholders Instead

```go
query := "SELECT * FROM coffees WHERE id = ?"
db.Query(query, id)
```

## Testing Your Implementation

```bash
# Install the driver
go get github.com/go-sql-driver/mysql

# Setup MySQL database
mysql -u root -p
CREATE DATABASE coffee_db;
USE coffee_db;

# Run your server
go run main.go

# Test with curl
curl -X POST http://localhost:8080/coffees \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Coffee","rating":8}'

# Check database directly
mysql -u root -p coffee_db
SELECT * FROM coffees;
```

## Pro Tips

1. **Always use `parseTime=true` in DSN**: Converts MySQL DATETIME to `time.Time`
2. **Use prepared statements**: The `?` placeholders prevent SQL injection
3. **Check `RowsAffected()`**: Know if UPDATE/DELETE actually did something
4. **Use `defer rows.Close()`**: Prevent connection leaks
5. **Handle `sql.ErrNoRows`**: Distinguish "not found" from other errors

## Order of Implementation

1. Start with `initTable()` - creates the table structure
2. Implement `Save()` - test creating coffees
3. Implement `GetByID()` - test retrieving coffees
4. Implement `GetAll()` - test listing coffees
5. Implement `Update()` - test modifying coffees
6. Implement `Delete()` - test removing coffees

Good luck! Remember: Start with in-memory storage, get it working, then switch to MySQL! 🚀
