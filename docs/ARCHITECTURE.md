# Architecture

## System Overview

CoffeeDex is a single Go binary: an Ebitengine desktop game wired directly to
services and SQLite storage. There is no HTTP layer.

```
+---------------------------------------------------+
|                game/ (Ebitengine)                 |
|  scene_menu / scene_warehouse / scene_roastery    |
|  scene_pokemon_lab / scene_trophy_room            |
|  shared widgets: ui.go, textinput.go, etc.        |
+-----------------------+---------------------------+
                        |
                        |  game.Services struct
                        v
+---------------------------------------------------+
|                    service/                       |
|  CoffeeService    BrewService    BrewerService    |
|  PokemonService   StatisticsService               |
|          |                                        |
|          v                                        |
|  ClaudeService (OpenRouter LLM, optional)         |
+-----------------------+---------------------------+
                        |
                        v
+---------------------------------------------------+
|                    storage/                       |
|  SQLite implementations behind interfaces         |
|  (coffee, brew, brewer, pokemon stores)           |
+-----------------------+---------------------------+
                        |
                        v
                 coffee-dex.db (SQLite file)
```

## Startup Flow (`main.go`)

1. Load `.env` from the repo root (sets any unset environment variables,
   notably `OPENROUTER_API_KEY`)
2. Parse flags: `-db` (SQLite path), `-enable-claude` (LLM on/off)
3. Open/create the SQLite database and apply schema
4. Construct storage stores, then services, then the optional LLM service
5. Initialize Pokemon reference data (Gen 1)
6. Hand the services to `game.Run()` and start the game loop

## Game Layer

Each screen is a scene implementing Ebitengine's update/draw pattern. Scenes
own their own input handling and layout. `game/game.go` holds the scene stack
and switching; `game/services.go` defines the `Services` struct that main uses
to inject dependencies. Scenes never touch storage directly - they call
services.

## Data Model

### Entities

**Coffee** - Bean information only
```go
type Coffee struct {
    ID               string    `json:"id"`
    Name             string    `json:"name"`
    Origin           string    `json:"origin"`
    Roaster          string    `json:"roaster"`
    Variety          string    `json:"variety"`
    RoastLevel       string    `json:"roast_level"`
    ProcessingMethod string    `json:"processing_method"`
    RoastDate        *DateOnly `json:"roast_date,omitempty"`
    IsFinished       bool      `json:"is_finished"`
    FinishedAt       *time.Time `json:"finished_at,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

**Brew** - Per-tasting evaluation data
```go
type Brew struct {
    ID            string        `json:"id"`
    CoffeeID      string        `json:"coffee_id"`
    TastingNotes  [5]string     `json:"tasting_notes"`
    TastingTraits TastingTraits `json:"tasting_traits"`
    Rating        int           `json:"rating"`
    Recipe        []string      `json:"recipe"`
    Dripper       string        `json:"dripper"`
    EndTime       DrawDownTime  `json:"end_time"`
    CreatedAt     time.Time     `json:"created_at"`
}
```

**TastingTraits** - Flavor profile (all 0-10 scale)
```go
type TastingTraits struct {
    BerryIntensity        int `json:"berry_intensity"`
    StoneFruitIntensity   int `json:"stonefruit_intensity"`
    RoastIntensity        int `json:"roast_intensity"`
    CitrusFruitsIntensity int `json:"citrus_fruits_intensity"`
    Bitterness            int `json:"bitterness"`
    Florality             int `json:"florality"`
    Spice                 int `json:"spice"`
    Sweetness             int `json:"sweetness"`
    AromaticIntensity     int `json:"aromatic_intensity"`
    Savory                int `json:"savory"`
    Body                  int `json:"body"`
    Cleanliness           int `json:"cleanliness"`
}
```

**CoffeePokemon** - Generated Pokemon mapping
```go
type CoffeePokemon struct {
    ID                string         `json:"id"`
    CoffeeID          string         `json:"coffee_id"`
    PokemonID         int            `json:"pokemon_id"`
    PokemonName       string         `json:"pokemon_name"`
    Nickname          string         `json:"nickname,omitempty"`
    Level             int            `json:"level"`
    MappingConfidence float64        `json:"mapping_confidence"`
    LLMDescription    string         `json:"llm_description"`
    TraitMapping      []TraitMapping `json:"trait_mapping"`
    CreatedAt         time.Time      `json:"created_at"`
}
```

### Relationships

```
Coffee (1) ──────────< Brew (many)
   │
   └── (0..1) ──────── CoffeePokemon (generated after 5+ brews)

Brewer (1) ──────────< Recipe (many)
```

- One Coffee can have many Brews
- One Coffee can have at most one CoffeePokemon (generated after 5+ brews)
- Brewers hold reusable recipes referenced by brews

## Storage Layer

`storage/storage.go` defines the store interfaces. The SQLite implementations
(`sqlite_*.go`) are the live ones:

Tables:
- `coffees` - Bean information
- `brews` - Tasting sessions with JSON columns for complex types
- `coffee_pokemon` - Generated Pokemon with UNIQUE constraint on coffee_id
- `pokemons` - Static Pokemon reference data (Gen 1, types, stats)
- `brewers` - Brewing devices
- `brewer_recipes` - Standalone recipes for brewers

The MySQL (`mysql*.go`) and in-memory (`memory.go`) implementations remain
from the old HTTP-server era and are not wired into `main.go`. The `handlers/`
directory is legacy from the same era.

## Service Layer

### BrewService

Brew CRUD and aggregation. **AggregatedBrewData** combines all brews for a
coffee: averaged tasting traits, combined unique tasting notes, average
rating, and brew count.

### PokemonService

Pokemon generation:
```go
func (s *PokemonService) MapCoffeeToPokemon(coffeeID string) (*CoffeePokemon, error)
func (s *PokemonService) CanGeneratePokemon(coffeeID string) (bool, error)
```

Flow:
1. Verify coffee exists and has 5+ brews
2. Verify no existing Pokemon for this coffee
3. Get aggregated brew data (averaged traits, combined notes)
4. Calculate candidate Pokemon types from traits (`pokemon_mapper.go`)
5. Filter unassigned Pokemon by matching types
6. LLM selects the best match and writes a description (rule-based fallback)
7. Save and return CoffeePokemon

### ClaudeService

OpenRouter client (`anthropic/claude-sonnet-4-5` via
`openrouter.ai/api/v1/chat/completions`) used for Pokemon selection. Receives
coffee info, averaged traits, combined notes, and the candidate list; returns
the selected Pokemon, confidence, and a lore-based description.

### StatisticsService

Analytics over coffees, brews, and Pokemon (roaster stats, ratings history,
collection totals).

## Data Flow Example: Generating a Pokemon

```
User triggers generation in Pokemon Lab
        |
        v
PokemonService.MapCoffeeToPokemon
        |
        +-> Verify 5+ brews exist
        +-> Verify no existing Pokemon
        +-> BrewService.GetAggregatedData (average traits, combine notes)
        +-> PokemonMapper.CalculatePokemonTypesFromTraits
        +-> PokemonStorage filter (unassigned, matching types)
        +-> ClaudeService (OpenRouter) or rule-based fallback
        +-> PokemonStorage.SaveCoffeePokemon
                |
                v
        Scene shows the new Pokemon with its description
```

See [POKEMON_MAPPING.md](POKEMON_MAPPING.md) for the mapping algorithm.
