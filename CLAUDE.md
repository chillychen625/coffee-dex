# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working in this repository.

## Build Commands

### Backend (Go)

```bash
# Run the server with MySQL storage
go run main.go -storage=mysql -mysql-password=yourpassword

# Run with in-memory storage (no persistence)
go run main.go -storage=memory

# Build binary
go build -o coffee-dex-server .
```

### Frontend (Electron + React)

```bash
cd coffee-dex-desktop

# Install dependencies
npm install

# Development mode (hot reload for renderer)
npm run dev

# Start full Electron app (spawns backend automatically)
npm start

# Build for production
npm run build
npm run dist
```

### Database

```bash
# Create/reset database
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS coffee_log"
mysql -u root -p coffee_log < sql/schema.sql
```

## Architecture

### Layered Backend

```
HTTP Handlers (handlers/) -> Services (service/) -> Storage (storage/)
```

- **Handlers**: Request parsing, response formatting, route handling
- **Services**: Business logic, validation, orchestration
- **Storage**: Database operations via interfaces (MySQL or memory implementations)

### Data Model

- **Coffee**: Bean information (name, origin, roaster, variety, roast_level, processing_method)
- **Brew**: Per-tasting data (tasting_notes, tasting_traits, rating, dripper, end_time) linked to Coffee via coffee_id
- **CoffeePokemon**: Generated Pokemon (one per coffee, after 5+ brews)

Relationships:
- Coffee 1:N Brew (one coffee has many brews)
- Coffee 0..1 CoffeePokemon (one coffee has at most one Pokemon)

### Frontend Structure

Electron app with React renderer:
- `src/main/` - Electron main process (spawns Go backend)
- `src/renderer/` - React components and views
- `src/services/api.ts` - API client for backend communication
- `src/types/` - TypeScript type definitions

### Pokemon Generation Flow

1. Coffee must have 5+ brews logged
2. Brew data is aggregated (averaged traits, combined notes)
3. Pokemon types calculated from averaged traits
4. LLM (Ollama) selects best Pokemon from candidates
5. Falls back to rule-based selection if LLM unavailable

## Key Files

- `main.go` - Entry point, routing, dependency wiring
- `sql/schema.sql` - Database schema
- `service/pokemon_mapper.go` - Trait-to-type calculation logic
- `service/llm.go` - Ollama integration
- `coffee-dex-desktop/src/renderer/App.tsx` - Main React component with view routing
