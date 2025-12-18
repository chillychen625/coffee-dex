# Coffee-Pokemon Integration Plan (Final)

## Overview

Transform coffee tasting entries into Pokemon with automatic trait mapping, creating a unique "CoffeeDex" desktop application that looks and feels like an authentic Nintendo DS Pokedex.

## Revised Scope

- **Generations**: Gen 1 only (151 total Pokemon)
- **Images**: Gen 1 Pokemon sprites only - local assets, minimal overhead (~150KB)
- **Platform**: Electron desktop app with TypeScript frontend calling existing Go backend
- **LLM Integration**: Qwen3:4b via Ollama for nuanced mappings with structured outputs
- **Uniqueness**: One Pokemon per coffee, no duplicates within Gen 1 pool

## Current System Analysis

### Coffee Storage Architecture

- **Layered Architecture**: Models → Storage → Service → Handlers
- **Coffee Model**: 12 tasting traits (0-10 scale), metadata, validation
- **Storage Options**: In-memory (testing) and MySQL (production)
- **API**: RESTful endpoints with JSON responses on localhost:8080
- **Logging**: Request/response logging middleware

### Tasting Traits (0-10 scale)

- `berry_intensity`, `stonefruit_intensity`, `roast_intensity`
- `citrus_fruits_intensity`, `bitterness`, `florality`
- `spice`, `sweetness`, `aromatic_intensity`
- `savory`, `body`, `cleanliness`

## Pokemon Mapping Strategy

### 1. Primary Type Mapping (Gen 1 Focus)

| Coffee Trait Dominance            | Pokemon Type  | Gen 1 Examples        |
| --------------------------------- | ------------- | --------------------- |
| High Sweetness + Low Bitterness   | Normal/Fairy  | Clefairy, Jigglypuff  |
| High Bitterness + High Roast      | Fire/Dark     | Charmander, Growlithe |
| High Citrus + High Florality      | Grass/Fairy   | Bulbasaur, Oddish     |
| High Berry + High Fruit           | Poison/Grass  | Bulbasaur, Bellsprout |
| High Spice + High Body            | Ground/Fire   | Charmander, Ponyta    |
| High Savory + High Roast          | Electric/Fire | Magnemite, Voltorb    |
| High Cleanliness + Low Bitterness | Water/Ice     | Squirtle, Psyduck     |
| High Aromatic + High Florality    | Psychic/Fairy | Jigglypuff, Mr. Mime  |

### 2. Generation Assignment (Coffee Complexity → Pokemon Rarity)

**Common (Pidgey, Caterpie, Rattata) - Basic coffees**

- Total trait variance < 15
- 1 dominant trait (>7)
- Standard processing methods

**Uncommon (Pikachu, Eevee, Jigglypuff) - Good coffees**

- Total trait variance 15-30
- 2-3 balanced traits
- Recognizable roasters/regions

**Rare (Charizard, Blastoise, Venusaur) - Excellent coffees**

- Total trait variance 30-45
- 3-4 significant traits
- Single origin, specialty processing

**Legendary (Mew, Mewtwo, Articuno) - Exceptional coffees**

- Total trait variance >45
- 4+ high traits (>8)
- Award-winning, limited edition

### 3. Qwen3:4b LLM Integration

#### LLM-Assigned Pokemon Selection

```go
type LLMMappingRequest struct {
    CoffeeName    string        `json:"coffee_name"`
    Origin        string        `json:"origin"`
    TastingTraits TastingTraits `json:"tasting_traits"`
    TastingNotes  []string      `json:"tasting_notes"`
    Candidates    []Pokemon     `json:"candidates"`
}

type LLMMappingResponse struct {
    SelectedPokemon  string   `json:"selected_pokemon"`
    Confidence       float64  `json:"confidence"`
    Description      string   `json:"description"`
    TraitMapping     []string `json:"trait_mapping"`
}
```

**Qwen3:4b Prompt Template:**

```
You are a Pokemon expert specializing in coffee-Pokemon mappings.
Given a coffee's characteristics, select the best Gen 1 Pokemon match and write a Pokedex-style description.

Coffee: {name} from {origin}
Tasting Notes: {tasting_notes}
Dominant Traits: {trait_description}

Available Pokemon: {candidate_list}

Respond with ONLY valid JSON:
{
  "selected_pokemon": "exact_pokemon_name",
  "confidence": 0.95,
  "description": "Pokedex-style description connecting coffee traits to Pokemon characteristics",
  "trait_mapping": ["sweetness -> Pokemon's sweet nature", "bitterness -> bold personality"]
}
```

#### Implementation with Qwen3:4b

```go
func mapCoffeeWithLLM(coffee models.Coffee) (string, error) {
    // Get rule-based candidates first
    candidates := getRuleBasedCandidates(coffee)

    // Call Ollama with qwen3:4b
    response, err := callOllamaQwen(coffee, candidates)
    if err != nil {
        log.Printf("LLM mapping failed: %v, falling back to rules", err)
        return getRuleBasedMapping(coffee)
    }

    // Validate and return
    return response.SelectedPokemon, nil
}

func callOllamaQwen(coffee models.Coffee, candidates []Pokemon) (*LLMMappingResponse, error) {
    client := &http.Client{Timeout: 30 * time.Second}

    payload := map[string]interface{}{
        "model":  "qwen3:4b",
        "prompt": buildPrompt(coffee, candidates),
        "stream": false,
        "format": "json",
    }

    // Send request and parse response
    // ... implementation details
}
```

### 4. Uniqueness Enforcement

```go
func ensureUniquePokemon(coffee models.Coffee, mappedPokemon string) (string, error) {
    usedPokemon, err := getUsedPokemon()
    if err != nil {
        return "", err
    }

    if usedPokemon[mappedPokemon] {
        // Find alternative with similar traits
        alternatives := findSimilarPokemon(coffee, mappedPokemon)
        if len(alternatives) > 0 {
            // Reserve the alternative
            reservePokemon(alternatives[0])
            return alternatives[0], nil
        }
        return "", fmt.Errorf("no unique Pokemon available - collection complete!")
    }

    reservePokemon(mappedPokemon)
    return mappedPokemon, nil
}
```

## Electron Desktop App Architecture

### Technology Stack

- **Backend**: Existing Go API on localhost:8080
- **Frontend**: TypeScript + React + Electron
- **Styling**: CSS with DS Pokedex authentic styling
- **LLM**: Qwen3:4b via Ollama local API
- **Assets**: Gen 1 Pokemon sprites (151 x 32x32px)

### Project Structure

```
coffee-dex-desktop/
├── main.js                 # Electron main process
├── preload.js              # Electron preload script
├── package.json
├── src/
│   ├── renderer/
│   │   ├── components/
│   │   │   ├── Pokedex/
│   │   │   │   ├── TopScreen.tsx
│   │   │   │   ├── BottomScreen.tsx
│   │   │   │   └── Navigation.tsx
│   │   │   ├── Coffee/
│   │   │   │   └── CoffeeForm.tsx
│   │   │   └── Layout/
│   │   │       └── PokedexFrame.tsx
│   │   ├── hooks/
│   │   │   ├── usePokemon.ts
│   │   │   └── useCoffee.ts
│   │   ├── types/
│   │   │   ├── pokemon.ts
│   │   │   └── coffee.ts
│   │   ├── services/
│   │   │   ├── api.ts          # Backend API calls
│   │   │   └── pokemonApi.ts   # Pokemon-specific API
│   │   └── styles/
│   │       └── pokedex.css
├── static/
│   └── pokemon-sprites/       # Gen 1 Pokemon images
└── dist/                      # Built Electron app
```

### Electron Configuration

```javascript
// main.js
const { app, BrowserWindow } = require("electron");
const path = require("path");

function createWindow() {
  const mainWindow = new BrowserWindow({
    width: 320, // Slightly wider for UI elements
    height: 240, // Taller for comfortable viewing
    minWidth: 320,
    minHeight: 240,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, "preload.js"),
    },
    frame: false, // Custom frame
    transparent: true, // Rounded corners effect
    backgroundColor: "#00000000",
  });

  mainWindow.loadFile("dist/index.html");
}

app.whenReady().then(createWindow);
```

### API Integration Layer

```typescript
// src/renderer/services/api.ts
const API_BASE = "http://localhost:8080";

export class CoffeeAPI {
  async getAllCoffees(): Promise<Coffee[]> {
    const response = await fetch(`${API_BASE}/coffees`);
    return response.json();
  }

  async createCoffee(coffee: Omit<Coffee, "id">): Promise<Coffee> {
    const response = await fetch(`${API_BASE}/coffees`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(coffee),
    });
    return response.json();
  }

  async generatePokemon(coffeeId: string): Promise<PokemonMapping> {
    const response = await fetch(`${API_BASE}/coffees/${coffeeId}/pokemon`, {
      method: "POST",
    });
    return response.json();
  }
}
```

## Authentic Pokedex UI Design

### Layout Specifications

- **Dimensions**: 320px x 240px (comfortable desktop size)
- **Screen Ratio**: 4:3 aspect ratio like DS
- **Colors**: Classic Pokedex blue/gray with yellow accents
- **Fonts**: Monospace for authentic terminal feel

### Top Screen (Information Display)

```
┌─────────────────────────────────────────┐
│  ☰ CoffeeDex                    [⚙️]   │
│                                         │
│     [Pokemon Sprite - 64x64px]          │
│                                         │
│  Pokemon: {name}                       │
│  Coffee: {coffee_name}                 │
│                                         │
│     [◀ Previous]     [Next ▶]          │
└─────────────────────────────────────────┘
```

### Bottom Screen (Details Display)

```
┌─────────────────────────────────────────┐
│ Type: {type}           Gen: 1           │
│ Level: {level} ★★★★★     Rating: {rating} │
│                                         │
│ {llm_generated_pokedex_description}    │
│                                         │
│ STATS:                                  │
│ HP:    ████████████████████ 255         │
│ ATK:   ████████████████████ 255         │
│ DEF:   ████████████████████ 255         │
│ SPD:   ████████████████████ 255         │
│ SPC:   ████████████████████ 255         │
│                                         │
│ [View Coffee]  [Regenerate]  [★]        │
└─────────────────────────────────────────┘
```

### CSS Styling (pokedex.css)

```css
.pokedex-container {
  width: 320px;
  height: 240px;
  background: linear-gradient(145deg, #4a90e2, #357abd);
  border-radius: 20px;
  border: 4px solid #2c5282;
  box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.3), 0 15px 35px rgba(0, 0, 0, 0.5);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-family: "Courier New", monospace;
  color: #00ff00;
}

.screen {
  background: #1a3d17;
  border: 3px solid #0f2910;
  border-radius: 8px;
  padding: 8px;
  color: #00ff00;
  font-family: "Courier New", monospace;
  font-size: 11px;
  line-height: 1.3;
  box-shadow: inset 0 0 15px rgba(0, 255, 0, 0.2);
}

.top-screen {
  height: 70px;
  text-align: center;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.bottom-screen {
  height: 140px;
  overflow-y: auto;
}

.pokemon-sprite {
  width: 64px;
  height: 64px;
  image-rendering: pixelated;
  margin: 0 auto;
  border: 2px solid #00ff00;
  border-radius: 4px;
}

.nav-button {
  background: #2d5a27;
  border: 1px solid #00ff00;
  color: #00ff00;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 10px;
}

.nav-button:hover {
  background: #3d7a37;
}

.stat-bar {
  background: #0f2910;
  height: 6px;
  border-radius: 3px;
  overflow: hidden;
  margin: 2px 0;
  border: 1px solid #00ff00;
}

.stat-fill {
  background: linear-gradient(90deg, #ff0000, #ffff00, #00ff00);
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}

.rating-stars {
  color: #ffff00;
  font-size: 12px;
}

.screen-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  font-weight: bold;
}

.action-button {
  background: #2d5a27;
  border: 1px solid #00ff00;
  color: #00ff00;
  padding: 3px 6px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 9px;
  margin: 0 2px;
}

.action-button:hover {
  background: #3d7a37;
}
```

## Database Schema Updates

### Extend Existing MySQL Schema

```sql
-- Add Pokemon mapping to existing coffees table
ALTER TABLE coffees ADD COLUMN pokemon_mapping JSON;

-- Pokemon reference table (Gen 1 only)
CREATE TABLE pokemons (
    id INT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    sprite_path VARCHAR(255) NOT NULL,
    base_stats JSON NOT NULL,
    description TEXT
);

-- Coffee-Pokemon mapping
CREATE TABLE coffee_pokemon (
    id VARCHAR(36) PRIMARY KEY,
    coffee_id VARCHAR(36) NOT NULL,
    pokemon_id INT NOT NULL,
    nickname VARCHAR(100),
    level INT DEFAULT 1,
    mapping_confidence REAL,
    llm_description TEXT,
    trait_mapping JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (coffee_id) REFERENCES coffees(id),
    FOREIGN KEY (pokemon_id) REFERENCES pokemons(id)
);

-- Ensure Pokemon uniqueness
CREATE UNIQUE INDEX idx_unique_pokemon ON coffee_pokemon(pokemon_id);

-- Seed Gen 1 Pokemon data
INSERT INTO pokemons (id, name, type, sprite_path, base_stats, description) VALUES
(1, 'Bulbasaur', 'Grass/Poison', '/sprites/001-bulbasaur.png', '{"hp":45,"attack":49,"defense":49,"speed":45,"special":65}', 'A strange seed was planted on its back at birth. The plant sprouts and grows with this Pokemon.'),
-- ... continue for all 151 Gen 1 Pokemon
```

### Go Backend Extensions

#### New Models (`models/pokemon.go`)

```go
type Pokemon struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    Type        string `json:"type"`
    SpritePath  string `json:"sprite_path"`
    BaseStats   Stats  `json:"base_stats"`
    Description string `json:"description"`
}

type Stats struct {
    HP      int `json:"hp"`
    Attack  int `json:"attack"`
    Defense int `json:"defense"`
    Speed   int `json:"speed"`
    Special int `json:"special"`
}

type CoffeePokemon struct {
    ID                string  `json:"id"`
    CoffeeID          string  `json:"coffee_id"`
    PokemonID         int     `json:"pokemon_id"`
    PokemonName       string  `json:"pokemon_name"`
    Nickname          string  `json:"nickname"`
    Level             int     `json:"level"`
    MappingConfidence float64 `json:"mapping_confidence"`
    LLMDescription    string  `json:"llm_description"`
    TraitMapping      string  `json:"trait_mapping"`
    CreatedAt         time.Time `json:"created_at"`
}
```

#### New Service (`service/pokemon.go`)

```go
type PokemonService struct {
    coffeeService *service.CoffeeService
    storage       PokemonStorage
    llmService    *LLMService
}

type PokemonStorage interface {
    GetAllPokemon() ([]models.Pokemon, error)
    GetPokemonByTypeAndGeneration(pokemonType string, generation int) ([]models.Pokemon, error)
    ReservePokemon(pokemonID int) error
    IsPokemonUsed(pokemonID int) (bool, error)
    CreateCoffeePokemon(mapping models.CoffeePokemon) error
    GetCoffeePokemon(coffeeID string) (*models.CoffeePokemon, error)
}

func (s *PokemonService) MapCoffeeToPokemon(coffee models.Coffee) (*models.CoffeePokemon, error) {
    // 1. Rule-based candidate selection
    candidates := s.getRuleBasedCandidates(coffee)

    // 2. LLM refinement with Qwen3:4b
    llmResponse, err := s.llmService.MapCoffeeToPokemon(coffee, candidates)
    if err != nil {
        log.Printf("LLM mapping failed, using rules: %v", err)
        return s.getRuleBasedMapping(coffee)
    }

    // 3. Ensure uniqueness
    finalPokemon, err := s.ensureUniquePokemon(coffee, llmResponse.SelectedPokemon)
    if err != nil {
        return nil, err
    }

    // 4. Create mapping
    mapping := &models.CoffeePokemon{
        ID:                uuid.New().String(),
        CoffeeID:          coffee.ID,
        PokemonID:         finalPokemon.ID,
        PokemonName:       finalPokemon.Name,
        MappingConfidence: llmResponse.Confidence,
        LLMDescription:    llmResponse.Description,
        TraitMapping:      s.formatTraitMapping(llmResponse.TraitMapping),
        CreatedAt:         time.Now(),
    }

    return s.storage.CreateCoffeePokemon(*mapping)
}
```

#### New Handlers (`handlers/pokemon.go`)

```go
type PokemonHandler struct {
    service *service.PokemonService
}

func (h *PokemonHandler) GeneratePokemon(w http.ResponseWriter, r *http.Request) {
    coffeeID := r.PathValue("coffee_id")

    // Get coffee
    coffee, err := h.service.GetCoffee(coffeeID)
    if err != nil {
        respondError(w, http.StatusNotFound, "Coffee not found")
        return
    }

    // Generate Pokemon mapping
    mapping, err := h.service.MapCoffeeToPokemon(coffee)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondJSON(w, http.StatusCreated, mapping)
}

func (h *PokemonHandler) GetCoffeeDex(w http.ResponseWriter, r *http.Request) {
    mappings, err := h.service.GetAllCoffeePokemon()
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to fetch CoffeeDex")
        return
    }

    respondJSON(w, http.StatusOK, mappings)
}
```

#### API Route Extensions

```go
// Add to main.go mux
mux.HandleFunc("/coffees/{id}/pokemon", pokemonHandler.GeneratePokemon)
mux.HandleFunc("/pokedex", pokemonHandler.GetCoffeeDex)
```

## Implementation Steps

### Phase 1: Electron App Setup (Week 1)

1. **Initialize Electron Project**

   ```bash
   npm init -y
   npm install electron electron-builder typescript @types/node
   npm install react react-dom @types/react @types/react-dom
   npx tsc --init
   ```

2. **Basic Pokedex UI**

   - Create PokedexFrame component
   - Implement top/bottom screens
   - Add Pokemon sprite display

3. **Backend API Integration**
   - Create API service layer
   - Test connection to existing Go backend
   - Implement coffee CRUD operations

### Phase 2: Pokemon Mapping (Week 2)

1. **Database Extensions**

   - Add Pokemon tables to MySQL
   - Seed Gen 1 Pokemon data
   - Create Pokemon service layer

2. **Rule-Based Mapping**

   - Implement type determination
   - Add generation assignment
   - Create uniqueness enforcement

3. **LLM Integration**
   - Set up Ollama connection with Qwen3:4b
   - Create structured prompt system
   - Add confidence scoring

### Phase 3: Enhanced Features (Week 3)

1. **Navigation System**

   - Previous/Next Pokemon browsing
   - Search and filter functionality
   - Collection statistics

2. **Coffee Integration**

   - Add "Generate Pokemon" button to coffee forms
   - Show mapping progress
   - Display mapping results with explanations

3. **Data Persistence**
   - Save Pokemon nicknames
   - Track mapping confidence scores
   - Export CoffeeDex data

### Phase 4: Polish and Distribution (Week 4)

1. **UI Polish**

   - Refine Pokedex styling
   - Add animations and transitions
   - Implement responsive design

2. **Performance Optimization**

   - Lazy load Pokemon sprites
   - Cache API responses
   - Optimize rendering

3. **Build and Distribution**
   ```bash
   npm run build
   npm run package  # Creates platform-specific installers
   ```

## Gen 1 Pokemon Sprite Assets

### Asset Specifications

- **Format**: PNG with transparency
- **Size**: 64x64 pixels (larger for desktop clarity)
- **Style**: Classic pixel art from original games
- **Total Size**: ~300KB for all 151 sprites
- **Source**: Official Pokemon sprites or high-quality fan art

### File Structure

```
static/pokemon-sprites/
├── 001-bulbasaur.png
├── 002-ivysaur.png
├── 003-venusaur.png
├── ...
├── 151-mew.png
└── pokemon-index.json  # Metadata
```

### Pokemon Index Metadata

```json
{
  "1": {
    "name": "Bulbasaur",
    "type": "Grass/Poison",
    "description": "A strange seed was planted on its back at birth.",
    "stats": {
      "hp": 45,
      "attack": 49,
      "defense": 49,
      "speed": 45,
      "special": 65
    }
  }
}
```

## Performance Considerations

### Desktop App Performance

- **Startup Time**: < 3 seconds on modern hardware
- **Memory Usage**: < 100MB RAM
- **API Response Time**: < 2 seconds for Pokemon mapping
- **Sprite Loading**: Lazy load with cache

### LLM Performance

- **Model**: Qwen3:4b (4GB VRAM requirement)
- **Response Time**: < 10 seconds per mapping
- **Fallback**: Rule-based mapping if LLM unavailable
- **Caching**: Cache responses for similar coffee profiles

## Distribution and Installation

### Build Commands

```bash
# Development
npm run dev

# Production build
npm run build

# Create installers
npm run package

# Creates:
# - macOS: CoffeeDex-darwin-x64.app
# - Windows: CoffeeDex-win32-x64.exe
# - Linux: CoffeeDex-linux-x64.AppImage
```

### Installation Requirements

- **Operating System**: macOS 10.14+, Windows 10+, Ubuntu 18.04+
- **Memory**: 4GB RAM (2GB for app, 2GB for Qwen3:4b)
- **Storage**: 500MB (100MB app + 300MB sprites + 100MB Ollama model)
- **Network**: localhost access to existing Go backend

### User Experience Flow

1. **Start Go Backend**: `go run main.go -storage=mysql`
2. **Start Ollama**: `ollama run qwen3:4b`
3. **Launch CoffeeDex Desktop App**
4. **Import existing coffees or create new ones**
5. **Generate Pokemon mappings automatically**
6. **Browse collection in authentic Pokedex interface**

## Success Metrics

1. **Mapping Accuracy**: >90% user satisfaction with Pokemon assignments
2. **Uniqueness Achievement**: 100% unique Pokemon mapping until collection complete
3. **LLM Utilization**: >80% of mappings using Qwen3:4b over rule-based
4. **User Engagement**: Average session >10 minutes browsing CoffeeDex
5. **Performance**: Average mapping time <5 seconds total

This refined plan provides a focused, implementable approach to creating an authentic Pokedex desktop experience for coffee enthusiasts using Electron, Qwen3:4b, and Gen 1 Pokemon while leveraging your existing robust Go backend.
