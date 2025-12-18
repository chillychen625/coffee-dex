# CoffeeDex Implementation Guide

## Overview

This guide provides step-by-step instructions for completing and running the CoffeeDex Pokemon integration system.

## Current Implementation Status

### ✅ Completed Components

**Backend (Go)**

- Pokemon models (`models/pokemon.go`)
- Pokemon storage layer (`storage/pokemon_storage.go`)
- Pokemon service with rule-based mapping (`service/pokemon.go`)
- LLM integration service (`service/llm.go`)
- Pokemon API handlers (`handlers/pokemon.go`)
- Updated main.go with Pokemon routes
- Database schema for Pokemon tables

**Frontend (Electron + TypeScript)**

- Project structure setup (`coffee-dex-desktop/`)
- TypeScript configuration (`tsconfig.json`)
- Package configuration (`package.json`)
- Main Electron process (`main.ts`)
- Pokedex styling (`src/styles/pokedex.css`)
- Type definitions (`src/types/pokemon.ts`)
- Basic React component structure

### 🔄 In Progress/Remaining

**Backend Tasks**

1. Create Pokemon data seeding script
2. Database initialization for Pokemon tables
3. Complete error handling and validation

**Frontend Tasks**

1. Complete React components for Pokedex interface
2. Download Gen 1 Pokemon sprites
3. API integration layer
4. Main application component

## Setup Instructions

### 1. Backend Setup

#### Install Dependencies

```bash
go mod download
```

#### Database Setup (MySQL)

```bash
# Start MySQL if not running
mysql.server start

# Run database setup
./scripts/setup_database.sh
```

#### Run Backend Server

```bash
# With MySQL and Pokemon features
go run main.go -storage=mysql -enable-llm=true

# With in-memory storage (Pokemon features disabled)
go run main.go -storage=memory
```

#### LLM Setup (Optional)

```bash
# Install and start Ollama
curl -fsSL https://ollama.ai/install.sh | sh
ollama run qwen3:4b
```

### 2. Frontend Setup

#### Install Dependencies

```bash
cd coffee-dex-desktop
npm install
```

#### Download Pokemon Sprites

```bash
# Create sprites directory
mkdir -p static/pokemon-sprites

# Download Gen 1 Pokemon sprites from PokemonDB
# Format: https://img.pokemondb.net/sprites/red-blue/normal/{pokemon}.png
# Example for Bulbasaur: https://img.pokemondb.net/sprites/red-blue/normal/bulbasaur.png

# Create sprites index file
cat > static/pokemon-sprites/index.json << 'EOF'
{
  "1": {
    "name": "bulbasaur",
    "display_name": "Bulbasaur",
    "type": "Grass/Poison"
  },
  "4": {
    "name": "charmander",
    "display_name": "Charmander",
    "type": "Fire"
  },
  "7": {
    "name": "squirtle",
    "display_name": "Squirtle",
    "type": "Water"
  }
  // Add all 151 Gen 1 Pokemon...
}
EOF
```

#### Build and Run Desktop App

```bash
# Development mode
npm run dev

# Production build
npm run build
npm run package
```

## API Endpoints

### Coffee Operations (Existing)

- `GET /coffees` - List all coffees
- `POST /coffees` - Create new coffee
- `GET /coffees/{id}` - Get specific coffee
- `PUT /coffees/{id}` - Update coffee
- `DELETE /coffees/{id}` - Delete coffee

### Pokemon Operations (New)

- `POST /coffees/{id}/pokemon` - Generate Pokemon for coffee
- `GET /coffees/{id}/pokemon` - Get Pokemon for coffee
- `PUT /coffees/{coffee_id}/pokemon/nickname` - Update Pokemon nickname
- `GET /pokedex` - Get complete CoffeeDex collection
- `GET /pokedex/stats` - Get collection statistics

## Pokemon Mapping Algorithm

### Rule-Based Type Determination

```go
// Determine primary Pokemon type based on coffee traits
if traits.Sweetness >= 7 && traits.Bitterness <= 3 {
    return "Normal" // Sweet, pleasant nature
}
if traits.Bitterness >= 7 && traits.RoastIntensity >= 6 {
    return "Fire" // Bold, intense characteristics
}
// ... additional rules
```

### LLM Enhancement (Qwen3:4b)

- **Prompt**: Structured JSON input with coffee characteristics
- **Response**: Pokemon selection with confidence and description
- **Fallback**: Rule-based mapping if LLM unavailable
- **Structured Output**: JSON response with Pokemon name, confidence, description, trait mapping

### Uniqueness Enforcement

- Each Pokemon can only be mapped to one coffee
- Automatic alternative selection for duplicate Pokemon
- Collection completion tracking

## Database Schema

### Pokemon Reference Table

```sql
CREATE TABLE pokemons (
    id INT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    sprite_path VARCHAR(255) NOT NULL,
    base_stats JSON NOT NULL,
    description TEXT
);
```

### Coffee-Pokemon Mapping Table

```sql
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

CREATE UNIQUE INDEX idx_unique_pokemon ON coffee_pokemon(pokemon_id);
```

## UI Design

### Pokedex Interface Specifications

- **Dimensions**: 320px x 240px (authentic DS size)
- **Top Screen**: Pokemon sprite, name, navigation
- **Bottom Screen**: Type, stats, description, actions
- **Styling**: Classic Pokedex blue/gray with green CRT screen effect
- **Font**: Courier New monospace for authentic feel

### Key UI Components

1. **PokedexFrame** - Main container with authentic styling
2. **TopScreen** - Pokemon information display
3. **BottomScreen** - Detailed stats and description
4. **Navigation** - Previous/Next Pokemon browsing
5. **ActionButtons** - View coffee, regenerate, favorite

## Testing the Integration

### 1. Test Backend API

```bash
# Create a test coffee
curl -X POST http://localhost:8080/coffees \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ethiopian Yirgacheffe",
    "origin": "Ethiopia",
    "roaster": "Test Roaster",
    "roast_level": "light",
    "processing_method": "washed",
    "tasting_notes": ["floral", "citrus", "tea-like"],
    "rating": 8,
    "tasting_traits": {
      "sweetness": 7,
      "bitterness": 2,
      "citrus_fruits_intensity": 8,
      "florality": 9,
      "body": 5,
      "aromatic_intensity": 8
    }
  }'

# Generate Pokemon for the coffee
curl -X POST http://localhost:8080/coffees/{coffee_id}/pokemon

# View CoffeeDex
curl http://localhost:8080/pokedex
```

### 2. Test Desktop App

1. Ensure backend is running on localhost:8080
2. Start desktop app: `npm run dev`
3. Create new coffee or use existing one
4. Click "Generate Pokemon" button
5. Browse Pokemon collection in Pokedex interface

## Configuration Options

### Backend Flags

- `-storage=memory|mysql` - Storage backend
- `-enable-llm=true|false` - Enable LLM mapping
- `-ollama-url=http://localhost:11434` - Ollama API URL
- `-ollama-model=qwen3:4b` - LLM model name

### Frontend Configuration

- **API Base URL**: `http://localhost:8080` (configurable)
- **Pokemon Sprites**: Local static files in `static/pokemon-sprites/`
- **Window Size**: Fixed 320x240px (DS Pokedex size)

## Troubleshooting

### Common Issues

1. **MySQL Connection Failed**

   - Check MySQL is running: `mysql.server status`
   - Verify database exists: `mysql -u root -e "SHOW DATABASES;"`

2. **Pokemon Features Disabled**

   - Requires MySQL storage backend
   - Check server logs for Pokemon initialization messages

3. **LLM Mapping Failed**

   - Verify Ollama is running: `ollama list`
   - Check model is available: `ollama run qwen3:4b`
   - Fallback to rule-based mapping is automatic

4. **Desktop App Won't Start**
   - Check backend is running on localhost:8080
   - Verify all dependencies installed: `npm install`
   - Check Electron installation: `npx electron --version`

### Debug Mode

```bash
# Backend with detailed logging
LOG_LEVEL=debug go run main.go -storage=mysql -enable-llm=true

# Desktop app with DevTools
# Press F12 or use menu: View > Developer Tools
```

## Next Steps

### Immediate Tasks

1. Complete React components for Pokedex interface
2. Download and organize Gen 1 Pokemon sprites
3. Create Pokemon data seeding script
4. Test end-to-end integration

### Future Enhancements

1. Add Gen 2-4 Pokemon support
2. Implement Pokemon battle system
3. Add collection sharing features
4. Create mobile companion app
5. Add coffee recipe recommendations based on Pokemon

## Support

For issues or questions:

1. Check server logs for error messages
2. Verify all services are running (MySQL, Ollama, backend API)
3. Test API endpoints directly with curl
4. Check browser console for frontend errors

## License and Attribution

- Pokemon sprites: PokemonDB (https://pokemondb.net)
- Pokemon is a trademark of Nintendo, Game Freak, and Creatures Inc.
- This is a fan project for educational/entertainment purposes
