# Development Guide

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- MySQL 8.0 or higher
- Ollama (optional, for LLM Pokemon mapping)

## Initial Setup

### 1. Clone and Install Dependencies

```bash
git clone <repository-url>
cd coffee-dex

# Install Go dependencies
go mod download

# Install frontend dependencies
cd coffee-dex-desktop
npm install
```

### 2. Database Setup

Create the MySQL database and tables:

```bash
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS coffee_log"
mysql -u root -p coffee_log < sql/schema.sql
```

The schema creates:
- `coffees` - Bean information
- `brews` - Tasting sessions
- `coffee_pokemon` - Generated Pokemon mappings
- `pokemon` - Static Gen 1 Pokemon data
- `brewers` - Brewing devices
- `brewer_recipes` - Standalone recipes

### 3. Ollama Setup (Optional)

Install Ollama and pull a model:

```bash
# Install Ollama (macOS)
brew install ollama

# Start Ollama
ollama serve

# Pull a model (in another terminal)
ollama pull qwen3:4b
```

## Running the Application

### Backend Only

```bash
# With MySQL storage
go run main.go -storage=mysql -mysql-password=yourpassword

# With in-memory storage (no persistence)
go run main.go -storage=memory

# Without LLM (uses fallback mapping)
go run main.go -storage=mysql -mysql-password=yourpassword -enable-llm=false
```

### Frontend Only (Development)

```bash
cd coffee-dex-desktop
npm run dev
```

This starts the renderer process with hot reload. You need to run the backend separately.

### Full Application (Electron)

```bash
cd coffee-dex-desktop
npm start
```

The Electron app spawns the Go backend automatically on startup.

## Project Structure

```
coffee-dex/
├── main.go              # Entry point, routing, dependency wiring
├── go.mod               # Go module definition
├── models/              # Data structures
├── storage/             # Database implementations
├── service/             # Business logic
├── handlers/            # HTTP request handlers
├── sql/                 # Database schemas
├── scripts/             # Utility scripts
├── static/              # Static files (sprites)
└── coffee-dex-desktop/  # Electron frontend
    ├── src/
    │   ├── main/        # Electron main process
    │   ├── renderer/    # React components
    │   ├── services/    # API client
    │   ├── types/       # TypeScript types
    │   ├── styles/      # CSS
    │   └── components/  # Shared components
    ├── build.js         # Build script
    └── package.json
```

## Development Workflow

### Backend Changes

1. Make changes to Go files
2. Restart the server: `go run main.go -storage=mysql -mysql-password=yourpassword`
3. Test with curl or the frontend

### Frontend Changes

1. Make changes to TypeScript/React files
2. If using `npm run dev`, changes auto-reload
3. If using `npm start`, restart Electron

### Database Schema Changes

1. Update `sql/schema.sql`
2. Either:
   - Drop and recreate: `mysql -u root -p coffee_log < sql/schema.sql`
   - Or write a migration script

### Adding a New Endpoint

1. Define the route in `main.go`
2. Create or update handler in `handlers/`
3. Create or update service in `service/`
4. Create or update storage interface/implementation in `storage/`
5. Add frontend API method in `src/services/api.ts`
6. Update TypeScript types in `src/types/`

## Building for Production

### Backend Binary

```bash
go build -o coffee-dex-server .
```

### Electron Application

```bash
cd coffee-dex-desktop
npm run build
npm run dist
```

This creates platform-specific distributables in `dist/`.

## Testing

### Populate Test Data

```bash
go run scripts/populate_test_data.go -storage=mysql -mysql-password=yourpassword
```

This creates sample coffees and brews for testing.

### Manual API Testing

```bash
# Health check
curl http://localhost:8080/health

# Create coffee
curl -X POST http://localhost:8080/coffees \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Coffee","origin":"Test Origin","roast_level":"medium","processing_method":"washed"}'

# Create brew
curl -X POST http://localhost:8080/brews \
  -H "Content-Type: application/json" \
  -d '{"coffee_id":"<coffee-id>","rating":7,"dripper":"V60","tasting_notes":["note1","","","",""],"tasting_traits":{"berry_intensity":5,"stonefruit_intensity":5,"roast_intensity":5,"citrus_fruits_intensity":5,"bitterness":5,"florality":5,"spice":5,"sweetness":5,"aromatic_intensity":5,"savory":5,"body":5,"cleanliness":5},"end_time":{"minutes":3,"seconds":0}}'

# Check brew progress
curl http://localhost:8080/coffees/<coffee-id>/brew-progress

# Generate Pokemon (after 5+ brews)
curl -X POST http://localhost:8080/pokemon/<coffee-id>
```

## Troubleshooting

### Backend won't start

- Check MySQL is running: `mysql -u root -p -e "SELECT 1"`
- Verify database exists: `mysql -u root -p -e "SHOW DATABASES LIKE 'coffee_log'"`
- Check port 8080 is free: `lsof -i :8080`

### LLM not working

- Check Ollama is running: `curl http://localhost:11434/api/tags`
- Verify model is installed: `ollama list`
- Check model name matches `-ollama-model` flag

### Frontend can't connect to backend

- Verify backend is running on port 8080
- Check for CORS issues in browser console
- Ensure the Electron main process spawned the backend

### Database errors

- Check MySQL connection parameters
- Verify schema is up to date
- Look for foreign key constraint violations

## Code Style

### Go

- Use `gofmt` for formatting
- Follow standard Go project layout
- Error messages should be lowercase, no punctuation

### TypeScript

- Use TypeScript strict mode
- Define interfaces for all API responses
- Use functional React components with hooks

### CSS

- GameBoy-inspired styling in `pokemon-gameboy.css`
- Use existing CSS classes when possible
- Prefix custom classes with `pokemon-`
