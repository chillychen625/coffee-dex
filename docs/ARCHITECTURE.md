# Architecture

## System Overview

CoffeeDex follows a layered architecture with clear separation between data access, business logic, and presentation.

```
┌─────────────────────────────────────────────────────┐
│                  Electron Frontend                   │
│         (React + TypeScript + GameBoy CSS)          │
└─────────────────────────────────────────────────────┘
                          │
                          │ HTTP/REST
                          ▼
┌─────────────────────────────────────────────────────┐
│                   HTTP Handlers                      │
│     (coffee.go, brew.go, pokemon.go, etc.)          │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│                    Services                          │
│  (CoffeeService, BrewService, PokemonService, etc.) │
└─────────────────────────────────────────────────────┘
                          │
              ┌───────────┴───────────┐
              ▼                       ▼
┌─────────────────────┐   ┌─────────────────────┐
│      Storage        │   │    LLM Service      │
│  (MySQL/Memory)     │   │     (Ollama)        │
└─────────────────────┘   └─────────────────────┘
```

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
    BerryIntensity       int `json:"berry_intensity"`
    StoneFruitIntensity  int `json:"stonefruit_intensity"`
    RoastIntensity       int `json:"roast_intensity"`
    CitrusFruitsIntensity int `json:"citrus_fruits_intensity"`
    Bitterness           int `json:"bitterness"`
    Florality            int `json:"florality"`
    Spice                int `json:"spice"`
    Sweetness            int `json:"sweetness"`
    AromaticIntensity    int `json:"aromatic_intensity"`
    Savory               int `json:"savory"`
    Body                 int `json:"body"`
    Cleanliness          int `json:"cleanliness"`
}
```

**CoffeePokemon** - Generated Pokemon mapping
```go
type CoffeePokemon struct {
    ID                string        `json:"id"`
    CoffeeID          string        `json:"coffee_id"`
    PokemonID         int           `json:"pokemon_id"`
    PokemonName       string        `json:"pokemon_name"`
    Nickname          string        `json:"nickname,omitempty"`
    Level             int           `json:"level"`
    MappingConfidence float64       `json:"mapping_confidence"`
    LLMDescription    string        `json:"llm_description"`
    TraitMapping      []TraitMapping `json:"trait_mapping"`
    CreatedAt         time.Time     `json:"created_at"`
}
```

### Relationships

```
Coffee (1) ──────────< Brew (many)
   │
   └── (1) ──────────< CoffeePokemon (0..1)
```

- One Coffee can have many Brews
- One Coffee can have at most one CoffeePokemon (generated after 5+ brews)

## Storage Layer

### Interfaces

```go
type CoffeeStorage interface {
    Save(coffee Coffee) error
    GetByID(id string) (*Coffee, error)
    GetAll() ([]Coffee, error)
    GetRecent(limit int) ([]Coffee, error)
    Update(coffee Coffee) error
    Delete(id string) error
}

type BrewStorage interface {
    Save(brew Brew) error
    GetByID(id string) (*Brew, error)
    GetByCoffeeID(coffeeID string) ([]Brew, error)
    GetBrewCount(coffeeID string) (int, error)
    GetAggregatedData(coffeeID string) (*AggregatedBrewData, error)
    GetAll() ([]Brew, error)
    GetRecent(limit int) ([]Brew, error)
    Delete(id string) error
}

type PokemonStorage interface {
    SaveCoffeePokemon(cp CoffeePokemon) error
    GetCoffeePokemon(coffeeID string) (*CoffeePokemon, error)
    HasPokemon(coffeeID string) bool
    GetAllCoffeePokemon() ([]CoffeePokemon, error)
    GetPokemonByID(id int) (*Pokemon, error)
    GetPokemonByType(pokemonType string) ([]Pokemon, error)
    GetAllPokemon() ([]Pokemon, error)
}
```

### MySQL Implementation

Tables:
- `coffees` - Bean information
- `brews` - Tasting sessions with JSON columns for complex types
- `coffee_pokemon` - Generated Pokemon with UNIQUE constraint on coffee_id
- `pokemon` - Static Pokemon data (Gen 1, types, stats)
- `brewers` - Brewing devices
- `brewer_recipes` - Standalone recipes for brewers

## Service Layer

### BrewService

Handles brew CRUD and aggregation:

```go
func (s *BrewService) CreateBrew(brew Brew) (*Brew, error)
func (s *BrewService) GetBrewsForCoffee(coffeeID string) ([]Brew, error)
func (s *BrewService) GetBrewProgress(coffeeID string, hasPokemon bool) (*BrewProgress, error)
func (s *BrewService) GetAggregatedData(coffeeID string) (*AggregatedBrewData, error)
```

**AggregatedBrewData** combines all brews:
- Averaged tasting traits
- Combined unique tasting notes
- Average rating
- Brew count

### PokemonService

Handles Pokemon generation:

```go
func (s *PokemonService) MapCoffeeToPokemon(coffeeID string) (*CoffeePokemon, error)
func (s *PokemonService) CanGeneratePokemon(coffeeID string) (bool, error)
func (s *PokemonService) HasPokemon(coffeeID string) bool
```

Pokemon generation flow:
1. Verify coffee exists and has 5+ brews
2. Verify no existing Pokemon for this coffee
3. Get aggregated brew data (averaged traits, combined notes)
4. Calculate candidate Pokemon types from traits
5. Filter Pokemon by matching types
6. Use LLM to select best Pokemon and generate description
7. Save and return CoffeePokemon

### LLMService

Integrates with Ollama for intelligent Pokemon selection:

```go
func (s *LLMService) MapCoffeeToPokemonWithTraits(
    coffee Coffee,
    traits TastingTraits,
    combinedNotes []string,
    candidates []Pokemon,
) (*LLMPokemonMapping, error)
```

The LLM receives:
- Coffee bean information
- Averaged tasting traits from all brews
- Combined tasting notes
- Candidate Pokemon list

Returns:
- Selected Pokemon ID
- Confidence score
- Description explaining the match
- Trait-to-stat mappings

## Frontend Architecture

### Views

- `start` - Title screen
- `home` - Main menu
- `coffee-form` - Add new coffee (2 steps)
- `coffee-list` - Browse all coffees
- `coffee-detail` - Coffee info, brew progress, brew list
- `brew-form` - Log a brew (4 steps)
- `pokedex` - Pokemon collection viewer
- `statistics` - Analytics dashboard
- `special-items` - Brewer management
- `settings` - App settings

### State Management

Single AppState object with React useState:

```typescript
interface AppState {
  view: ViewType;
  coffees: Coffee[];
  currentCoffee: Coffee | null;
  currentBrews: Brew[];
  brewProgress: BrewProgress | null;
  currentPokemon: CoffeePokemon | null;
  pokedex: CoffeePokemon[];
  loading: boolean;
  error: string | null;
  // ... form state
}
```

### API Client

Centralized API service (`src/services/api.ts`) wrapping all HTTP calls:

```typescript
class CoffeeDexAPI {
  // Coffee endpoints
  getCoffees(): Promise<Coffee[]>
  getCoffee(id: string): Promise<Coffee>
  createCoffee(coffee: Partial<Coffee>): Promise<Coffee>

  // Brew endpoints
  createBrew(brew: Partial<Brew>): Promise<Brew>
  getBrewsForCoffee(coffeeId: string): Promise<Brew[]>
  getBrewProgress(coffeeId: string): Promise<BrewProgress>

  // Pokemon endpoints
  generatePokemon(coffeeId: string): Promise<CoffeePokemon>
  getCoffeePokemon(coffeeId: string): Promise<CoffeePokemon>
  getPokedex(): Promise<CoffeePokemon[]>
}
```

## Data Flow Examples

### Creating a Coffee and First Brew

```
User -> CoffeeForm -> POST /coffees -> CoffeeService.Create -> CoffeeStorage.Save
                                                                      |
User <- coffee-detail <- Coffee created                               |
                                                                      v
User -> BrewForm -> POST /brews -> BrewService.Create -> BrewStorage.Save
                                                                |
User <- coffee-detail <- Brew saved, progress updated           |
                                                                v
                                               GET /coffees/{id}/brew-progress
```

### Generating a Pokemon

```
User clicks "Generate Pokemon"
        |
        v
POST /pokemon/{coffee_id}
        |
        v
PokemonService.MapCoffeeToPokemon
        |
        ├─> Verify 5+ brews exist
        ├─> Verify no existing Pokemon
        ├─> BrewService.GetAggregatedData (average traits, combine notes)
        ├─> PokemonMapper.CalculatePokemonTypesFromTraits
        ├─> PokemonStorage.GetPokemonByType (filter candidates)
        ├─> LLMService.MapCoffeeToPokemonWithTraits
        └─> PokemonStorage.SaveCoffeePokemon
                |
                v
        CoffeePokemon returned
                |
                v
        User sees Pokemon in Pokedex
```
