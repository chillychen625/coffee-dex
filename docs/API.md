# API Reference

Base URL: `http://localhost:8080`

## Health

### GET /health

Check if the server is running.

**Response**
```json
{"status": "ok"}
```

## Coffees

### GET /coffees

List all coffees.

**Response**
```json
[
  {
    "id": "uuid",
    "name": "Ethiopia Yirgacheffe",
    "origin": "Ethiopia",
    "roaster": "Local Roasters",
    "variety": "Heirloom",
    "roast_level": "light",
    "processing_method": "washed",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

### GET /coffees/recent

Get recently created coffees (limit 10).

**Response**: Same as GET /coffees

### GET /coffees/{id}

Get a specific coffee by ID.

**Response**
```json
{
  "id": "uuid",
  "name": "Ethiopia Yirgacheffe",
  "origin": "Ethiopia",
  "roaster": "Local Roasters",
  "variety": "Heirloom",
  "roast_level": "light",
  "processing_method": "washed",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### POST /coffees

Create a new coffee.

**Request Body**
```json
{
  "name": "Ethiopia Yirgacheffe",
  "origin": "Ethiopia",
  "roaster": "Local Roasters",
  "variety": "Heirloom",
  "roast_level": "light",
  "processing_method": "washed"
}
```

**Roast Levels**: `light`, `medium`, `dark`, `light medium`, `medium dark`, `unclear`

**Processing Methods**: `washed`, `natural`, `honey`, `coferment`, `experimental`

**Response**: Created coffee object

### PUT /coffees/{id}

Update an existing coffee.

**Request Body**: Same as POST /coffees

**Response**: Updated coffee object

### DELETE /coffees/{id}

Delete a coffee. Also deletes associated brews and Pokemon.

**Response**: 204 No Content

## Brews

### GET /brews

List all brews.

**Response**
```json
[
  {
    "id": "uuid",
    "coffee_id": "uuid",
    "tasting_notes": ["blueberry", "floral", "honey", "", ""],
    "tasting_traits": {
      "berry_intensity": 8,
      "stonefruit_intensity": 5,
      "roast_intensity": 3,
      "citrus_fruits_intensity": 6,
      "bitterness": 2,
      "florality": 7,
      "spice": 3,
      "sweetness": 7,
      "aromatic_intensity": 6,
      "savory": 2,
      "body": 5,
      "cleanliness": 8
    },
    "rating": 8,
    "recipe": [],
    "dripper": "V60",
    "end_time": {
      "minutes": 3,
      "seconds": 15
    },
    "created_at": "2024-01-15T11:00:00Z"
  }
]
```

### GET /brews/recent

Get recently created brews (limit 10).

**Response**: Same as GET /brews

### GET /brews/{id}

Get a specific brew by ID.

**Response**: Single brew object

### POST /brews

Create a new brew.

**Request Body**
```json
{
  "coffee_id": "uuid",
  "tasting_notes": ["blueberry", "floral", "honey", "", ""],
  "tasting_traits": {
    "berry_intensity": 8,
    "stonefruit_intensity": 5,
    "roast_intensity": 3,
    "citrus_fruits_intensity": 6,
    "bitterness": 2,
    "florality": 7,
    "spice": 3,
    "sweetness": 7,
    "aromatic_intensity": 6,
    "savory": 2,
    "body": 5,
    "cleanliness": 8
  },
  "rating": 8,
  "recipe": [],
  "dripper": "V60",
  "end_time": {
    "minutes": 3,
    "seconds": 15
  }
}
```

All trait values are integers from 0-10.

**Response**: Created brew object

### DELETE /brews/{id}

Delete a brew.

**Response**: 204 No Content

### GET /coffees/{coffee_id}/brews

Get all brews for a specific coffee.

**Response**: Array of brew objects

### GET /coffees/{coffee_id}/brew-progress

Get brew count and Pokemon generation eligibility.

**Response**
```json
{
  "count": 3,
  "required": 5,
  "can_generate_pokemon": false,
  "has_pokemon": false
}
```

- `count`: Number of brews logged for this coffee
- `required`: Number of brews needed (always 5)
- `can_generate_pokemon`: True if count >= 5
- `has_pokemon`: True if Pokemon already generated

## Pokemon

### POST /pokemon/{coffee_id}

Generate a Pokemon for a coffee. Requires 5+ brews and no existing Pokemon.

**Response**
```json
{
  "id": "uuid",
  "coffee_id": "uuid",
  "pokemon_id": 12,
  "pokemon_name": "Butterfree",
  "nickname": "",
  "level": 42,
  "mapping_confidence": 0.85,
  "llm_description": "This coffee's delicate floral notes and light body remind me of Butterfree...",
  "trait_mapping": [
    {
      "trait": "florality",
      "pokemon_stat": "special",
      "reasoning": "High floral intensity maps to Butterfree's special attack"
    }
  ],
  "created_at": "2024-01-15T12:00:00Z"
}
```

**Error Responses**
- 400: Not enough brews (need 5+)
- 400: Pokemon already exists for this coffee
- 404: Coffee not found

### GET /pokemon/{coffee_id}

Get the Pokemon for a coffee.

**Response**: CoffeePokemon object

**Error Response**: 404 if no Pokemon exists

### PUT /pokemon/{coffee_id}/nickname

Update the Pokemon's nickname.

**Request Body**
```json
{
  "nickname": "My Favorite"
}
```

**Response**: Updated CoffeePokemon object

### GET /pokedex

Get all generated Pokemon.

**Response**
```json
[
  {
    "id": "uuid",
    "coffee_id": "uuid",
    "pokemon_id": 12,
    "pokemon_name": "Butterfree",
    "nickname": "",
    "level": 42,
    "mapping_confidence": 0.85,
    "llm_description": "...",
    "trait_mapping": [...],
    "created_at": "2024-01-15T12:00:00Z"
  }
]
```

### GET /pokedex/stats

Get Pokedex completion statistics.

**Response**
```json
{
  "total_pokemon": 5,
  "unique_pokemon": 4,
  "completion_percentage": 3.3
}
```

## Statistics

### GET /statistics

Get aggregated statistics from all brews.

**Response**
```json
{
  "total_coffees": 10,
  "total_pokemon": 5,
  "completion_percentage": 3.3,
  "average_rating": 7.5,
  "highest_rated_coffee": {
    "name": "Ethiopia Yirgacheffe",
    "rating": 9
  },
  "lowest_rated_coffee": {
    "name": "Brazil Santos",
    "rating": 5
  },
  "type_distribution": {
    "fire": 2,
    "grass": 3
  },
  "top_origins": [
    {
      "origin": "Ethiopia",
      "count": 5,
      "avg_rating": 8.2
    }
  ],
  "processing_methods": {
    "washed": 6,
    "natural": 4
  },
  "roast_levels": {
    "light": 5,
    "medium": 3,
    "dark": 2
  },
  "trait_averages": {
    "berry_intensity": 6.5,
    "sweetness": 7.0
  },
  "brewer_stats": [
    {
      "brewer": "V60",
      "count": 15,
      "avg_rating": 7.8,
      "avg_brew_time": 195
    }
  ],
  "confidence_metrics": {
    "average_confidence": 0.78,
    "high_confidence_count": 3,
    "medium_confidence_count": 2,
    "low_confidence_count": 0
  }
}
```

## Brewers

### GET /brewers

List all brewers with their recipes.

**Response**
```json
[
  {
    "id": "uuid",
    "name": "V60",
    "pokeball_type": "poke-ball",
    "recipes": [
      {
        "id": "uuid",
        "name": "4:6 Method",
        "steps": ["Bloom 50g", "Pour to 150g", "..."]
      }
    ],
    "created_at": "2024-01-15T10:00:00Z"
  }
]
```

### POST /brewers

Create a new brewer.

**Request Body**
```json
{
  "name": "V60",
  "pokeball_type": "poke-ball"
}
```

**Pokeball Types**: `poke-ball`, `great-ball`, `ultra-ball`, `master-ball`

**Response**: Created brewer object

### DELETE /brewers/{id}

Delete a brewer.

**Response**: 204 No Content

### GET /brewers/pokeball-types

Get available pokeball types.

**Response**
```json
["poke-ball", "great-ball", "ultra-ball", "master-ball"]
```

### POST /brewers/{id}/standalone-recipes

Add a recipe to a brewer.

**Request Body**
```json
{
  "name": "4:6 Method",
  "steps": ["Bloom 50g for 45s", "Pour to 150g", "Pour to 250g"]
}
```

**Response**: 201 Created

### DELETE /brewers/{id}/standalone-recipes/{recipe_id}

Remove a recipe from a brewer.

**Response**: 204 No Content

## Error Responses

All endpoints return errors in this format:

```json
{
  "error": "Error message here"
}
```

Common HTTP status codes:
- 400: Bad Request (validation error, business rule violation)
- 404: Not Found
- 500: Internal Server Error
