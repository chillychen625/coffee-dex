# CoffeeDex

A coffee tasting journal that transforms your brew experiences into Pokemon. Log your coffees, record multiple brews with tasting notes, and after 5 brews, generate a unique Pokemon based on your aggregated tasting data.

## Overview

CoffeeDex separates coffee beans from individual brew sessions:

- **Coffee**: Bean information (name, origin, roaster, variety, roast level, processing method)
- **Brew**: Per-tasting evaluation (tasting notes, flavor traits, rating, dripper, brew time)

After logging 5 brews for a coffee, you can generate a Pokemon. The Pokemon type and selection is based on the averaged tasting traits and combined tasting notes from all your brews. Each coffee can only have one Pokemon - no regeneration.

## Tech Stack

- **Backend**: Go with MySQL storage
- **Frontend**: Electron + React + TypeScript
- **LLM Integration**: Ollama for Pokemon selection reasoning

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Ollama (optional, for LLM-powered Pokemon mapping)

### Database Setup

```bash
mysql -u root -p < sql/schema.sql
```

### Start the Backend

```bash
go run main.go -storage=mysql -mysql-host=localhost:3306 -mysql-user=root -mysql-password=yourpassword -mysql-db=coffee_log
```

Backend flags:
- `-storage`: `memory` or `mysql` (default: memory)
- `-mysql-host`: MySQL host and port (default: localhost:3306)
- `-mysql-user`: MySQL username (default: root)
- `-mysql-password`: MySQL password
- `-mysql-db`: Database name (default: coffee_log)
- `-ollama-url`: Ollama URL (default: http://localhost:11434)
- `-ollama-model`: Ollama model (default: qwen3:4b)
- `-enable-llm`: Enable LLM mapping (default: true)

### Start the Frontend

```bash
cd coffee-dex-desktop
npm install
npm start
```

The Electron app will automatically spawn the Go backend when launched.

## Project Structure

```
coffee-dex/
├── main.go                 # Application entry point and routing
├── models/                 # Data models
│   ├── coffee.go          # Coffee bean model
│   ├── brew.go            # Brew session model
│   └── pokemon.go         # Pokemon and mapping models
├── storage/               # Data persistence
│   ├── interface.go       # Storage interfaces
│   ├── mysql.go           # MySQL coffee storage
│   ├── mysql_brew.go      # MySQL brew storage
│   ├── mysql_pokemon.go   # MySQL Pokemon storage
│   └── mysql_brewer.go    # MySQL brewer storage
├── service/               # Business logic
│   ├── coffee.go          # Coffee CRUD operations
│   ├── brew.go            # Brew management and aggregation
│   ├── pokemon.go         # Pokemon generation and mapping
│   ├── llm.go             # LLM integration for Pokemon selection
│   ├── pokemon_mapper.go  # Type calculation from traits
│   └── statistics.go      # Analytics and statistics
├── handlers/              # HTTP handlers
│   ├── coffee.go          # Coffee endpoints
│   ├── brew.go            # Brew endpoints
│   ├── pokemon.go         # Pokemon endpoints
│   ├── brewer.go          # Brewer management endpoints
│   └── statistics.go      # Statistics endpoints
├── sql/                   # Database schemas
│   └── schema.sql         # MySQL table definitions
├── static/                # Static assets
│   └── pokemon-sprites/   # Pokemon sprite images
├── coffee-dex-desktop/    # Electron frontend
│   ├── src/
│   │   ├── main/          # Electron main process
│   │   ├── renderer/      # React components
│   │   ├── services/      # API client
│   │   ├── types/         # TypeScript definitions
│   │   └── styles/        # CSS styles
│   └── package.json
└── docs/                  # Documentation
```

## User Flow

1. **Add a Coffee**: Enter bean information (name, origin, roaster, variety, roast level, processing method)
2. **Log Brews**: For each brew session, record tasting notes, flavor traits (0-10 scales), rating, dripper, and brew time
3. **Track Progress**: View brew count progress (X/5) on the coffee detail page
4. **Generate Pokemon**: After 5+ brews, click "Generate Pokemon" to create a unique Pokemon based on aggregated brew data
5. **View Pokedex**: Browse your Pokemon collection with coffee details and LLM analysis

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - System design and data flow
- [API Reference](docs/API.md) - REST endpoint documentation
- [Development](docs/DEVELOPMENT.md) - Setup and development workflow
- [Pokemon Mapping](docs/POKEMON_MAPPING.md) - How coffee traits map to Pokemon types

## License

MIT
