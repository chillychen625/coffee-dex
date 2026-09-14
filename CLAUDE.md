# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working in this repository.

## Build Commands

```bash
# Run the app (Ebitengine game, opens a window)
go run main.go

# Run with a specific database file (default: ./coffee-dex.db)
go run main.go -db=/path/to/coffee-dex.db

# Run without LLM Pokemon selection
go run main.go -enable-claude=false

# Build binary
make build-server   # outputs bin/coffee-dex

# Tests and lint
make test
make lint
```

There is no HTTP server anymore. `main.go` wires storage and services directly
into the game and calls `game.Run()`. A `.env` file in the repo root is loaded
automatically at startup; set `OPENROUTER_API_KEY` there to enable LLM Pokemon
selection.

## Architecture

```
main.go -> game/ (Ebitengine UI) -> service/ (business logic) -> storage/ (SQLite)
```

- **game/**: Ebitengine desktop app. Each screen is a scene file
  (`scene_warehouse.go`, `scene_roastery.go`, `scene_pokemon_lab.go`,
  `scene_trophy_room.go`, `scene_menu.go`) that handles its own input and
  drawing; `game.go` owns the scene-switching loop. Shared widgets live in
  `ui.go`, `textinput.go`, `autocomplete.go`, `dateinput.go`.
- **service/**: Business logic. Coffee/brew CRUD, statistics, Pokemon
  generation, and LLM integration.
- **storage/**: SQLite implementations behind interfaces. MySQL and in-memory
  implementations also exist here but are legacy from the old HTTP-server era
  and are not wired into `main.go`.
- **handlers/**: Legacy HTTP handlers from the removed Electron app. Not
  referenced by any code; safe to ignore.

### Data Model

- **Coffee**: Bean information (name, origin, roaster, variety, roast_level, processing_method)
- **Brew**: Per-tasting data (tasting_notes, tasting_traits, rating, dripper, end_time) linked to Coffee via coffee_id
- **Brewer**: Dripper/setup presets (with recipes)
- **CoffeePokemon**: Generated Pokemon (one per coffee, after 5+ brews)

Relationships:
- Coffee 1:N Brew (one coffee has many brews)
- Coffee 0..1 CoffeePokemon (one coffee has at most one Pokemon)

### Pokemon Generation Flow

1. Coffee must have 5+ brews logged
2. Brew data is aggregated (averaged traits, combined notes)
3. Pokemon types calculated from averaged traits (`service/pokemon_mapper.go`)
4. LLM (OpenRouter, `anthropic/claude-sonnet-4-5`) selects the best Pokemon
   from unassigned candidates and writes a lore-based description
   (`service/llm.go`)
5. Falls back to rule-based selection if the LLM is unavailable or the API key
   is not set

## Key Files

- `main.go` - Entry point, flag parsing, dependency wiring, .env loading
- `game/game.go` - Game loop and scene switching
- `game/services.go` - Service struct passed from main into the game
- `service/pokemon_mapper.go` - Trait-to-type calculation logic
- `service/pokemon.go` - Pokemon generation orchestration
- `service/llm.go` - OpenRouter integration
- `storage/storage.go` - Storage interfaces
- `storage/sqlite_db.go` - SQLite connection and schema setup
- `cmd/migrate/` - MySQL-to-SQLite migration tool
