# Coffee Log Backend

A RESTful API backend for logging and managing coffee tasting notes, built with Go.

## Features

- ✨ Create, read, update, and delete coffee tasting entries
- 📝 Track detailed tasting notes with validation
- 🎯 Support for multiple storage backends (in-memory, MySQL)
- 🔄 RESTful API design with proper HTTP methods
- 🛡️ Input validation and error handling
- 📊 Structured logging with request tracking

## Project Structure

```
go-coffee-log/
├── models/         # Data models and validation
├── storage/        # Storage layer interfaces and implementations
├── service/        # Business logic layer
├── handlers/       # HTTP request handlers
└── main.go         # Application entry point
```

## API Endpoints

### Coffee Entries

| Method   | Endpoint        | Description                 |
| -------- | --------------- | --------------------------- |
| `POST`   | `/coffees`      | Create a new coffee entry   |
| `GET`    | `/coffees`      | List all coffee entries     |
| `GET`    | `/coffees/{id}` | Get a specific coffee entry |
| `PUT`    | `/coffees/{id}` | Update a coffee entry       |
| `DELETE` | `/coffees/{id}` | Delete a coffee entry       |

## Coffee Entry Schema

```json
{
  "id": "uuid",
  "name": "Ethiopian Yirgacheffe",
  "origin": "Ethiopia",
  "roaster": "Local Roasters Co.",
  "roast_level": "light",
  "processing_method": "washed",
  "tasting_notes": ["floral", "citrus", "tea-like", "bright", "clean"],
  "tasting_traits": {
    "berry_intensity": 7,
    "stonefruit_intensity": 4,
    "roast_intensity": 2,
    "citrus_fruits_intensity": 9,
    "bitterness": 1,
    "florality": 8,
    "spice": 3,
    "sweetness": 7,
    "aromatic_intensity": 8,
    "savory": 2,
    "body": 5,
    "cleanliness": 9
  },
  "rating": 9,
  "recipe": ["20g coffee", "320g water", "95°C", "2:30 brew time"],
  "dripper": "V60",
  "end_time": {
    "minutes": 2,
    "seconds": 30
  },
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Validation Rules

- **name**: Required, cannot be empty
- **rating**: Must be between 0-10
- **roast_level**: One of: `light`, `medium`, `dark`, `light medium`, `medium dark`, `unclear`
- **processing_method**: One of: `washed`, `natural`, `honey`, `coferment`, `experimental`
- **tasting_notes**: Array of 1-5 strings
- **tasting_traits**: Object with 12 intensity measurements (0-10):
  - `berry_intensity`: Berry flavor presence
  - `stonefruit_intensity`: Stonefruit flavor presence
  - `roast_intensity`: Roast flavor intensity
  - `citrus_fruits_intensity`: Citrus flavor presence
  - `bitterness`: Bitterness level
  - `florality`: Floral notes intensity
  - `spice`: Spice notes intensity
  - `sweetness`: Sweetness level
  - `aromatic_intensity`: Aroma strength
  - `savory`: Savory notes presence
  - `body`: Body/mouthfeel weight
  - `cleanliness`: Clarity and cleanness of flavors
- **end_time**: Valid minutes/seconds (seconds < 60)

## Getting Started

### Prerequisites

- Go 1.22 or higher
- MySQL (optional, for MySQL storage backend)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/go-coffee-log.git
cd go-coffee-log
```

2. Install dependencies:

```bash
go mod download
```

### Running the Server

The server supports two storage backends that can be selected via command-line flags:

#### With In-Memory Storage (Default)

```bash
go run main.go
# or explicitly
go run main.go -storage=memory
```

#### With MySQL Storage

First, set up MySQL following [MYSQL_GUIDE.md](MYSQL_GUIDE.md), then:

```bash
go run main.go -storage=mysql \
  -mysql-host=localhost:3306 \
  -mysql-user=root \
  -mysql-password=yourpassword \
  -mysql-db=coffee_log
```

**Command-line Flags:**

- `-storage`: Storage type (`memory` or `mysql`, default: `memory`)
- `-mysql-host`: MySQL server address (default: `localhost:3306`)
- `-mysql-user`: MySQL username (default: `root`)
- `-mysql-password`: MySQL password (default: empty)
- `-mysql-db`: MySQL database name (default: `coffee_log`)

The server will start on `http://localhost:8080`

## Usage Examples

### Create a Coffee Entry

```bash
curl -X POST http://localhost:8080/coffees \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Ethiopian Yirgacheffe",
    "origin": "Ethiopia",
    "roaster": "Local Roasters",
    "roast_level": "light",
    "processing_method": "washed",
    "tasting_notes": ["floral", "citrus", "tea-like", "bright", "clean"],
    "rating": 9,
    "recipe": ["20g coffee", "320g water"],
    "dripper": "V60",
    "end_time": {"minutes": 2, "seconds": 30}
  }'
```

### List All Coffees

```bash
curl http://localhost:8080/coffees
```

### Get a Specific Coffee

```bash
curl http://localhost:8080/coffees/{id}
```

### Update a Coffee

```bash
curl -X PUT http://localhost:8080/coffees/{id} \
  -H 'Content-Type: application/json' \
  -d '{"name": "Updated Name", "rating": 10}'
```

### Delete a Coffee

```bash
curl -X DELETE http://localhost:8080/coffees/{id}
```

## Architecture

The application follows a layered architecture pattern:

1. **Models Layer** (`models/`): Defines data structures and validation logic
2. **Storage Layer** (`storage/`): Provides interfaces and implementations for data persistence
3. **Service Layer** (`service/`): Contains business logic and orchestrates storage operations
4. **Handlers Layer** (`handlers/`): Processes HTTP requests and responses
5. **Main** (`main.go`): Initializes components and configures routing

### Storage Implementations

- **MemoryStorage**: In-memory storage using Go maps (thread-safe with `sync.RWMutex`)
- **MySQLStorage**: Persistent storage using MySQL database

## Development

### Adding a New Storage Backend

1. Implement the `CoffeeStorage` interface in `storage/storage.go`
2. Create a new file in the `storage/` directory
3. Initialize your storage in `main.go`

### Running Tests

```bash
go test ./...
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
