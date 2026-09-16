# CoffeeDex

A coffee tasting journal that transforms your brew experiences into Pokemon. Log your coffees, record brews with tasting notes, and after 5 brews a unique Pokemon is generated from your aggregated tasting data.

Built as a Pokemon-style desktop game (Ebitengine), not a web app.

## Overview

CoffeeDex separates coffee beans from individual brew sessions:

- **Coffee**: Bean information (name, origin, roaster, variety, roast level, processing method)
- **Brew**: Per-tasting evaluation (tasting notes, flavor traits, rating, dripper, brew time)

After logging 5 brews for a coffee, you can generate a Pokemon. Type and selection are based on averaged tasting traits and combined tasting notes across all brews. Each coffee gets exactly one Pokemon - no regeneration.

## Tech Stack

- **Language**: Go
- **UI**: Ebitengine 2D game (scene-based desktop app)
- **Storage**: SQLite (single file database, created automatically)
- **LLM**: OpenRouter (claude-sonnet-4-5) for Pokemon selection and descriptions, with rule-based fallback

## Quick Start

```bash
# Optional: enable LLM Pokemon selection
echo "OPENROUTER_API_KEY=your-key" > .env

# Run (database file is created automatically on first launch)
make run
# or: go run main.go
```

Command-line flags:

- `-db`: SQLite database path (default: `./coffee-dex.db`)
- `-enable-claude`: Enable LLM Pokemon selection (default: true; falls back to rules if no API key)

A `.env` file in the repo root is loaded automatically at startup.

## Project Structure

```
coffee-dex/
├── main.go                 # Entry point: wiring, .env loading, launches the game
├── game/                   # Ebitengine UI
│   ├── game.go             # Game loop and scene switching
│   ├── scene_menu.go       # Main menu
│   ├── scene_warehouse.go  # Coffee (bean) management
│   ├── scene_roastery.go   # Brew logging
│   ├── scene_pokemon_lab.go# Pokedex / Pokemon detail and compare
│   ├── scene_trophy_room.go# Stats and history
│   └── ui.go, textinput.go, autocomplete.go, dateinput.go  # Shared widgets
├── models/                 # Data models (coffee, brew, brewer, pokemon)
├── storage/                # SQLite persistence behind interfaces
│   └── (mysql*.go, memory.go are legacy from the old HTTP server era)
├── service/                # Business logic
│   ├── coffee.go           # Coffee CRUD
│   ├── pokemon.go          # Pokemon generation orchestration
│   ├── pokemon_mapper.go   # Trait-to-type calculation
│   ├── llm.go              # OpenRouter integration
│   └── statistics.go       # Analytics
├── handlers/               # Legacy HTTP handlers (unused, kept for reference)
├── cmd/migrate/            # MySQL-to-SQLite migration tool
├── sql/                    # Legacy MySQL schema and Gen 1 Pokemon data
└── docs/                   # Architecture and Pokemon mapping notes
```

## User Flow

1. **Add a Coffee** (Warehouse): enter bean information
2. **Log Brews** (Roastery): record tasting notes, flavor traits (0-10 scales), rating, dripper, and brew time
3. **Track Progress**: brew count progress (X/5) per coffee
4. **Generate Pokemon**: after 5+ brews, a Pokemon is chosen based on aggregated brew data, with an LLM-written description
5. **Pokemon Lab**: browse your collection, view coffee detail, compare Pokemon

## all docs by claude probably

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - System design and data flow
- [Pokemon Mapping](docs/POKEMON_MAPPING.md) - How coffee traits map to Pokemon types

## License

MIT
