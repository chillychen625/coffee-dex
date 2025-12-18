# CoffeeDex 🗃️☕

A Pokemon-themed coffee logging system that automatically maps coffee tasting notes to Pokemon based on flavor profiles, origins, and processing methods.

![CoffeeDex](static/pokemon-sprites/025.png)

## 🎯 Quick Start

```bash
# Complete setup (database + data + dependencies)
make full-setup

# Start the backend server
make start-server

# Run the desktop app (in another terminal)
make run-desktop
```

## 📁 Project Structure

```
coffee-dex/
├── 📁 docs/                 # Documentation and guides
│   ├── README.md           # Main documentation
│   ├── IMPLEMENTATION_GUIDE.md
│   ├── IMPLEMENTATION_STATUS.md
│   ├── POKEMON_INTEGRATION_PLAN.md
│   └── TESTING.md
├── 📁 sql/                 # Database scripts and data
│   ├── setup_pokemon_database.sql
│   └── pokemon_gen1_data.sql
├── 📁 scripts/             # Utility scripts
│   └── download_pokemon_sprites.sh
├── 📁 coffee-dex-desktop/  # Electron desktop app
├── 📁 handlers/            # Go HTTP handlers
├── 📁 models/              # Go data models
├── 📁 service/             # Business logic services
├── 📁 storage/             # Database storage layer
├── 📁 static/              # Static assets
│   └── pokemon-sprites/    # Pokemon images (151 Gen 1)
├── 📄 Makefile             # Project management
├── 📄 .gitignore           # Git ignore rules
└── 📄 main.go              # Go application entry point
```

## 🚀 Features

### Backend (Go + MySQL)

- **Pokemon Database**: Complete Gen 1 Pokemon with stats and descriptions
- **Intelligent Mapping**: Rule-based + LLM-powered coffee-to-Pokemon assignment
- **RESTful API**: Complete endpoints for Pokemon generation and management
- **MySQL Integration**: Persistent storage with proper relationships

### Desktop App (Electron + TypeScript)

- **Authentic Pokedex UI**: Nintendo DS-style interface
- **Coffee Upload**: Tasting notes input with trait analysis
- **Pokemon Display**: Visual Pokemon entries with sprite integration
- **Real-time Mapping**: Instant coffee-to-Pokemon generation

### Mapping Algorithm

- **Coffee Type Analysis**: Water/Fire/Grass types for roast levels
- **Flavor Profiles**: Sweet→Fairy, Bitter→Dark, Acidic→Poison, etc.
- **Origin Mapping**: Geographic distribution matching
- **Processing Methods**: Coffee techniques mapped to Pokemon traits
- **Uniqueness**: Each Pokemon assigned to only one coffee

## 🔧 Database Setup

### Using Makefile (Recommended)

```bash
# Set up database schema
make setup-db

# Load Pokemon data
make load-pokemon-data

# Check database status
make db-info
```

### Manual Setup

```bash
# Create database and tables
mysql -u root < sql/setup_pokemon_database.sql

# Load Pokemon data
mysql -u root coffee_log < sql/pokemon_gen1_data.sql

# Verify setup
mysql -u root coffee_log -e "SELECT COUNT(*) FROM pokemons;"
```

## 🛠️ Development

### Prerequisites

- Go 1.19+
- Node.js 16+ (for desktop app)
- MySQL 8.0+
- Qwen3:4b LLM (optional, for enhanced mapping)

### Available Commands

```bash
make help           # Show all available commands
make install-deps   # Install Go and Node.js dependencies
make build-server   # Build Go server binary
make build-desktop  # Build Electron app
make test           # Run Go tests
make clean          # Clean build artifacts
make check-mysql    # Verify MySQL connection
```

### API Endpoints

- `GET /api/pokemon` - List all Pokemon
- `GET /api/pokemon/:id` - Get specific Pokemon
- `POST /api/pokemon/generate` - Generate Pokemon from coffee
- `POST /api/coffee` - Create coffee entry
- `GET /api/coffee/:id/pokemon` - Get Pokemon for coffee

## 🎮 Usage

1. **Start Backend**: `make start-server` (runs on http://localhost:8080)
2. **Launch Desktop App**: `make run-desktop`
3. **Upload Coffee**: Enter tasting notes in the Pokedex interface
4. **View Pokemon**: Automatically generated Pokemon based on coffee traits
5. **Browse Collection**: Navigate through your coffee-Pokemon mappings

## 📊 Pokemon Data

- **151 Gen 1 Pokemon** with authentic stats and descriptions
- **Sprite Integration**: All Pokemon have proper sprite files
- **Type Mapping**: Coffee characteristics mapped to Pokemon types
- **Unique Assignments**: Each Pokemon can only be assigned to one coffee

## 🤖 LLM Integration

The system supports Qwen3:4b for enhanced Pokemon mapping:

```bash
# Set LLM API key (optional)
export QWEN_API_KEY="your-api-key"

# Enable LLM mapping in configuration
```

## 📚 Documentation

- **[Implementation Guide](docs/IMPLEMENTATION_GUIDE.md)** - Detailed setup and architecture
- **[Implementation Status](docs/IMPLEMENTATION_STATUS.md)** - Current progress and remaining tasks
- **[Pokemon Integration Plan](docs/POKEMON_INTEGRATION_PLAN.md)** - Mapping algorithm details
- **[Testing Guide](docs/TESTING.md)** - Testing procedures and examples

## 🎯 Current Status

- ✅ **Backend**: 100% Complete
- ✅ **Database**: 100% Complete (151 Pokemon loaded)
- ✅ **Assets**: 100% Complete (sprites downloaded)
- 🔄 **Desktop App**: 80% Complete (UI ready, needs compilation fixes)
- 📋 **Documentation**: 100% Complete

See [Implementation Status](docs/IMPLEMENTATION_STATUS.md) for detailed progress.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🔮 Future Enhancements

- Gen 2-9 Pokemon expansion
- Advanced LLM models integration
- Web interface option
- Cloud synchronization
- Social features (share Pokemon mappings)
- Mobile app support

---

**Coffee + Pokemon = CoffeeDex** ☕🎮
