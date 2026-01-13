# Pokemon Mapping Algorithm

This document explains how CoffeeDex maps coffee tasting data to Pokemon types and selects specific Pokemon.

## Overview

Pokemon generation happens after a coffee has 5 or more brews logged. The system:

1. Aggregates all brew data (averages traits, combines notes)
2. Calculates primary and secondary Pokemon types from traits
3. Filters candidate Pokemon by matching types
4. Uses an LLM to select the best Pokemon and generate a description
5. Falls back to rule-based selection if LLM is unavailable

## Data Aggregation

Before mapping, all brews for a coffee are aggregated:

**Tasting Traits**: Each trait is averaged across all brews
```
AveragedSweetness = Sum(AllBrews.Sweetness) / BrewCount
```

**Tasting Notes**: All unique, non-empty notes are combined
```
CombinedNotes = Union(AllBrews.TastingNotes) - EmptyStrings
```

**Rating**: Averaged to calculate Pokemon level
```
Level = (AverageRating / 10) * 100
```

## Type Calculation

Pokemon types are determined by evaluating traits against weighted rules.

### Type Rules

Each Pokemon type has rules based on trait thresholds:

| Type | Primary Traits | Secondary Traits |
|------|----------------|------------------|
| Fire | roast_intensity >= 7 | bitterness >= 5 |
| Grass | florality >= 6 | cleanliness >= 5 |
| Water | cleanliness >= 7 | body <= 4 |
| Electric | citrus >= 7 | aromatic_intensity >= 5 |
| Psychic | florality >= 7 | aromatic_intensity >= 6 |
| Dark | roast_intensity >= 8 | bitterness >= 6 |
| Fairy | sweetness >= 7 | florality >= 5 |
| Bug | florality >= 5 AND citrus >= 5 | cleanliness >= 4 |
| Flying | aromatic_intensity >= 6 | body <= 5 |
| Poison | fermented/funky notes | spice >= 5 |
| Ground | body >= 7 | savory >= 4 |
| Rock | body >= 8 | bitterness >= 5 |
| Ice | cleanliness >= 8 | body <= 3 |
| Fighting | body >= 6 | bitterness >= 4 |
| Ghost | unusual processing | spice >= 6 |
| Dragon | all traits high | aromatic_intensity >= 7 |
| Normal | balanced traits | no strong characteristics |

### Processing Method Influence

Processing methods add type bonuses:

| Processing | Type Bonus |
|------------|------------|
| Natural | +Fire, +Fairy |
| Honey | +Bug, +Fairy |
| Washed | +Water, +Ice |
| Coferment | +Poison, +Ghost |
| Experimental | +Psychic, +Ghost |

### Roast Level Influence

Roast levels modify type scores:

| Roast | Type Modification |
|-------|-------------------|
| Light | +Grass, +Electric, -Fire |
| Medium | Neutral |
| Dark | +Fire, +Dark, -Grass |

### Keyword Matching

Tasting notes are scanned for type-associated keywords:

| Keywords | Type |
|----------|------|
| berry, blueberry, strawberry | Fairy |
| chocolate, cocoa, caramel | Fire |
| citrus, lemon, orange, grapefruit | Electric |
| floral, jasmine, rose, lavender | Grass, Psychic |
| honey, maple, sugar | Fairy, Bug |
| wine, fermented, funky | Poison |
| nutty, almond, hazelnut | Ground |
| spice, cinnamon, pepper | Fire, Fighting |
| tea, herbal | Grass |
| tropical, mango, pineapple | Flying |

### Score Calculation

Each type receives a score based on:
```
TypeScore = (PrimaryTraitMatches * 2) + SecondaryTraitMatches + ProcessingBonus + RoastBonus + KeywordMatches
```

The two highest-scoring types become the primary and secondary type.

## Pokemon Selection

### Candidate Filtering

After determining types, candidates are filtered:

1. Get all Gen 1 Pokemon matching primary type
2. If secondary type exists, prefer Pokemon with matching secondary type
3. If no matches, fall back to primary type only

### LLM Selection

When Ollama is available, an LLM prompt is constructed:

```
Given this coffee:
- Name: Ethiopia Yirgacheffe
- Origin: Ethiopia
- Roast: Light
- Process: Washed
- Averaged Traits: berry=8, floral=7, citrus=6, sweetness=7, body=4, ...
- Combined Notes: blueberry, jasmine, honey

Select from these Pokemon: [Butterfree, Venomoth, ...]

Respond with JSON containing pokemon_id, confidence, description, trait_mapping.
```

The LLM provides:
- Selected Pokemon and why
- Confidence score (0.0-1.0)
- Descriptive text explaining the match
- Trait-to-Pokemon-stat mappings

### Fallback Selection

Without LLM, a rule-based selection occurs:

1. Score each candidate Pokemon based on trait alignment
2. Prefer Pokemon whose stats match the coffee profile:
   - High sweetness -> High special stat
   - High body -> High defense/HP
   - High aromatic -> High speed
3. Select highest-scoring Pokemon
4. Generate basic description from template

## Confidence Calculation

Mapping confidence reflects how well the coffee matches the Pokemon:

```
Confidence = (TraitMatchScore + TypeMatchScore + NoteMatchScore) / MaxPossibleScore
```

Factors:
- How well traits align with type rules
- How many tasting notes match type keywords
- LLM's self-reported confidence (when available)

Confidence ranges:
- High (> 70%): Strong match
- Medium (40-70%): Moderate match
- Low (< 40%): Weak match, primarily type-based

## Examples

### Light Fruity Ethiopian

Traits: berry=9, floral=8, citrus=7, sweetness=8, body=3
Notes: blueberry, jasmine, bergamot, honey

Type Calculation:
- Fairy: berry >= 7 (yes), sweetness >= 7 (yes) = strong match
- Grass: floral >= 6 (yes), cleanliness >= 5 (yes) = good match
- Electric: citrus >= 7 (yes) = moderate match

Result: Fairy/Grass type -> Candidates: Clefairy, Jigglypuff, Bulbasaur, etc.

### Dark Roasted Indonesian

Traits: roast=9, body=8, bitterness=7, spice=6, sweetness=3
Notes: dark chocolate, tobacco, earthy, cedar

Type Calculation:
- Dark: roast >= 8 (yes), bitterness >= 6 (yes) = strong match
- Fire: roast >= 7 (yes), bitterness >= 5 (yes) = good match
- Ground: body >= 7 (yes), savory >= 4 (yes) = good match

Result: Dark/Fire type -> Candidates: Arcanine, Growlithe, etc.

### Experimental Cofermented

Traits: spice=8, aromatic=9, florality=7, citrus=6, cleanliness=4
Notes: wine, tropical fruit, jasmine, fermented

Type Calculation:
- Poison: fermented note (yes), spice >= 5 (yes) = strong match
- Ghost: experimental processing bonus + spice >= 6 (yes) = good match
- Psychic: florality >= 7 (yes), aromatic >= 6 (yes) = good match

Result: Poison/Psychic type -> Candidates: Gengar, Haunter, etc.

## Limitations

1. Only Gen 1 Pokemon (151) are available
2. Some type combinations have no matching Pokemon
3. LLM quality varies by model and prompt
4. Rule-based fallback is less nuanced than LLM selection
5. Aggregation smooths out brew-to-brew variation
