# Pokemon Type Mapping Algorithm

## Overview

The Coffee Dex uses a sophisticated multi-stage algorithm to map coffees to Pokemon. This algorithm combines rule-based type scoring, database filtering, and optional LLM enhancement to create meaningful, accurate coffee-Pokemon pairings.

## Architecture

```
Coffee Input
    ↓
[Stage 1] Type Mapper - Calculate type scores
    ↓
[Stage 2] Database Filter - Get Pokemon candidates
    ↓
[Stage 3] LLM/Rule Selection - Pick best Pokemon
    ↓
[Stage 4] Uniqueness Check - Ensure no duplicates
    ↓
Final Pokemon Mapping
```

---

## Stage 1: Type Score Calculation

**File**: `service/pokemon_mapper.go`

The mapper calculates a score (0-1) for each of the 13 Pokemon types based on coffee characteristics.

### Supported Types

1. **Normal** - Generic, balanced coffee
2. **Fire** - Roasty, savory, peppery
3. **Water** - Clean, light-bodied
4. **Grass** - Floral, vegetal, aromatic
5. **Electric** - Sharp acidity, citrus
6. **Ice** - Minty, cooling, crisp
7. **Poison** - Funky, spicy, experimental
8. **Ground** - Earthy, grainy, full-bodied
9. **Rock** - Stonefruit-forward
10. **Dark** - Heavily roasted, bitter
11. **Fairy** - Sweet, dessert-like
12. **Psychic** - Complex, highly specific notes
13. **Bug** - Spicy vibes

### Scoring Components

For each type, the algorithm evaluates:

#### 1. **Primary Traits** (weighted 2.0-3.0x)

Core characteristics that define the type.

Example (Fire type):

- `roast_intensity`: Weight 2.5, needs 7-10
- `savory`: Weight 2.0, needs 6-10
- `spice`: Weight 2.2, needs 7-10 (peppery)

#### 2. **Secondary Traits** (weighted 1.0-1.5x)

Supporting characteristics.

Example (Fire type):

- `bitterness`: Weight 1.2, needs 6-9
- `body`: Weight 1.0, needs 7-10

#### 3. **Keyword Matches** (+20 points possible)

Searches tasting notes for type-specific keywords.

Example (Fire type keywords):

- "pepper", "roast", "smoke", "char", "burnt", "toast", "caramel"

Matching algorithm:

```go
matches = 0
for each tasting_note in coffee.tasting_notes:
    for keyword in type_keywords:
        if keyword in tasting_note.lowercase():
            matches++
            break  // Count each note only once
keyword_score = matches / 5.0  // Normalize to 0-1
```

#### 4. **Processing Method Bonus** (multiplier)

Boosts score based on coffee processing.

Example (Fire type):

- Dark roast: 1.8x multiplier
- Medium dark: 1.5x multiplier

Example (Poison type):

- Natural: 1.5x
- Experimental: 1.8x
- Coferment: 1.7x

#### 5. **Roast Level Bonus** (multiplier)

Boosts score based on roast darkness.

Example (Electric type):

- Light: 1.6x
- Light medium: 1.3x

### Scoring Formula

```
For each type:
    score = 0
    max_possible = 0

    // Primary traits
    for trait in primary_traits:
        value = coffee.tasting_traits[trait.name]
        if value >= trait.min:
            normalized = min(value, trait.max) / 10.0
            contribution = normalized * trait.weight * 10.0
            score += contribution
        max_possible += trait.weight * 10.0

    // Secondary traits (same formula)

    // Keywords
    if has_keywords:
        keyword_score = count_matches(coffee.notes, keywords) / 5.0
        score += keyword_score * 20.0
        max_possible += 20.0

    // Processing bonus
    if processing_bonus[coffee.processing_method]:
        score *= processing_bonus[coffee.processing_method]

    // Roast bonus
    if roast_bonus[coffee.roast_level]:
        score *= roast_bonus[coffee.roast_level]

    // Normalize to 0-1
    final_score = min(score / max_possible, 1.0)
```

### Type Selection

After scoring all 13 types:

1. **Sort types by score** (highest first)
2. **Primary type**: Highest score ≥ minimum threshold
   - If no type meets threshold → `Normal` type
3. **Secondary type**: Second highest score ≥ 80% of its threshold
   - Optional, only if significantly different from primary

Example output:

```
Coffee: Ethiopian Natural, Sweetness=9, Berry=8, Floral=7

Type Scores:
- Fairy: 0.85 (≥ 0.65 threshold) ✓ PRIMARY
- Grass: 0.72 (≥ 0.55 * 0.8 = 0.44) ✓ SECONDARY
- Normal: 0.45 (< 0.40 threshold) ✗
```

---

## Stage 2: Database Candidate Filtering

**File**: `service/pokemon.go` → `getTypedCandidates()`

Once types are determined, we fetch Pokemon from the database:

```sql
-- Get Pokemon matching primary type
SELECT * FROM pokemons WHERE type LIKE '%Fire%'

-- Get Pokemon matching secondary type (if exists)
SELECT * FROM pokemons WHERE type LIKE '%Grass%'
```

The `GetPokemonByType()` function uses a `LIKE` query to match both:

- Single-type Pokemon (e.g., `"Fire"`)
- Dual-type Pokemon (e.g., `"Fire/Flying"`)

**Candidate limit**: Maximum 10 Pokemon sent to LLM (performance optimization)

If no matches found → Fallback to `Normal` type Pokemon

---

## Stage 3: Pokemon Selection

**File**: `service/pokemon.go` → `MapCoffeeToPokemon()`

### Option A: LLM-Enhanced Selection (Preferred)

If LLM service is configured:

1. **Send to LLM**:

   - Coffee details (name, origin, tasting notes, traits)
   - Candidate Pokemon list (max 10)
   - Prompt requesting JSON response

2. **LLM Response**:

   ```json
   {
     "selected_pokemon": "Charizard",
     "confidence": 0.85,
     "description": "Pokedex-style description",
     "trait_mapping": [
       {
         "trait": "roast_intensity",
         "pokemon_stat": "Attack",
         "reasoning": "High roast matches fierce attacking power"
       }
     ]
   }
   ```

3. **Validation**:
   - Verify selected Pokemon is in candidate list
   - If invalid → Fallback to Option B

### Option B: Rule-Based Selection (Fallback)

If LLM unavailable or fails:

1. **Select first candidate** from type-matched list
2. **Calculate confidence**: `type_score * 0.9`
3. **Build trait mapping**:

   ```go
   if sweetness >= 7 → maps to HP
   if bitterness >= 7 → maps to Attack
   if body >= 7 → maps to Defense
   if citrus >= 7 → maps to Speed
   if aroma >= 7 → maps to Special
   ```

4. **Generate description**:
   ```
   "Type-based mapping: Charizard (Fire-type) matches coffee's
   fire characteristics with 85% confidence"
   ```

---

## Stage 4: Uniqueness Enforcement

**File**: `service/pokemon.go` → `ensureUniquePokemon()`

Each Pokemon can only be used once across all coffees.

```
1. Check: Is pokemon.ID already in coffee_pokemon table?

2. If NOT used:
   → Return this Pokemon ✓

3. If USED:
   → Get all Pokemon of same type from database
   → Filter to unused Pokemon
   → Return first unused alternative

4. If NO alternatives:
   → Return error (will fail on database UNIQUE constraint)
```

Database constraint:

```sql
CREATE UNIQUE INDEX idx_unique_pokemon
ON coffee_pokemon(pokemon_id)
```

---

## Stage 5: Final Mapping Creation

**File**: `service/pokemon.go`

Combine all data into `CoffeePokemon` model:

```go
mapping := &CoffeePokemon{
    ID: uuid.New(),
    CoffeeID: coffee.ID,
    PokemonID: selected_pokemon.ID,
    PokemonName: selected_pokemon.Name,
    Level: rating * 5,  // 0-10 rating → 0-50 level
    MappingConfidence: confidence_score,
    LLMDescription: combined_description,
    TraitMapping: trait_mappings,
    CreatedAt: time.Now()
}
```

**Description Format**:

```
[LLM or rule-based description]

Type Analysis: This coffee exhibits fire-type characteristics
with strong roast_intensity, savory
```

---

## Complete Example

### Input Coffee

```json
{
  "name": "Ethiopia Guji Natural",
  "origin": "Ethiopia",
  "roast_level": "light",
  "processing_method": "natural",
  "rating": 9,
  "tasting_notes": ["blueberry", "strawberry", "floral", "honey", "tea"],
  "tasting_traits": {
    "sweetness": 9,
    "berry_intensity": 9,
    "florality": 8,
    "aromatic_intensity": 8,
    "citrus_fruits_intensity": 4,
    "body": 6,
    "cleanliness": 7,
    "bitterness": 3,
    "roast_intensity": 3,
    "spice": 2,
    "savory": 2,
    "stonefruit_intensity": 5
  }
}
```

### Stage 1: Type Scoring

**Fairy Type**:

- Primary: sweetness(9) × 3.0 = 27.0 points
- Primary: aromatic(8) × 2.0 = 16.0 points
- Secondary: florality(8) × 1.5 = 12.0 points
- Secondary: berry(9) × 1.5 = 13.5 points
- Keywords: "honey" match = 0.2 × 20 = 4.0 points
- Processing bonus: natural × 1.4 = multiply by 1.4
- **Final: 0.87** ✓ PRIMARY

**Grass Type**:

- Primary: florality(8) × 2.5 = 20.0 points
- Primary: aromatic(8) × 2.0 = 16.0 points
- Keywords: "floral", "tea" = 0.4 × 20 = 8.0 points
- Roast bonus: light × 1.5 = multiply by 1.5
- **Final: 0.76** ✓ SECONDARY

### Stage 2: Candidates

```sql
-- Fairy types: Clefairy, Clefable, Jigglypuff, Wigglytuff, Mr. Mime
-- Grass types: Bulbasaur, Ivysaur, Venusaur, Oddish, Gloom, Vileplume, ...
```

### Stage 3: LLM Selection

**LLM picks**: Clefairy (Fairy type, matches sweet/berry profile)

### Stage 4: Uniqueness

Check: Clefairy not used → ✓ Approved

### Final Mapping

```json
{
  "pokemon_id": 35,
  "pokemon_name": "Clefairy",
  "level": 45,
  "mapping_confidence": 0.87,
  "llm_description": "This enchanting Fairy-type perfectly captures
  the sweet, berry-forward character of this natural Ethiopian coffee...\n\n
  Type Analysis: This coffee exhibits fairy-type characteristics with
  strong sweetness, aromatic_intensity and grass-type characteristics
  with strong florality, aromatic_intensity",
  "trait_mapping": [
    {"trait": "sweetness", "pokemon_stat": "HP", "reasoning": "..."},
    {"trait": "aroma", "pokemon_stat": "Special", "reasoning": "..."}
  ]
}
```

---

## Performance Characteristics

- **Type Calculation**: O(13) - constant, all types evaluated
- **Database Query**: O(log n) - indexed type lookup
- **LLM Call**: ~2-5 seconds (optional, can be skipped)
- **Uniqueness Check**: O(log n) - indexed pokemon_id lookup

**Total time**: 2-6 seconds with LLM, <1 second without

---

## Future Enhancements

1. **Machine Learning**: Train ML model on user preferences to refine type weights
2. **Regional Variations**: Different type mappings for different coffee origins
3. **Dynamic Thresholds**: Adjust minimum thresholds based on coffee complexity
4. **Multi-Type Pokemon**: Prefer dual-type Pokemon when both types score high
5. **Evolution Chains**: Map coffee development (multiple brews) to Pokemon evolution
