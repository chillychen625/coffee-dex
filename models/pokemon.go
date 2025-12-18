package models

import "time"

// Pokemon represents a Gen 1 Pokemon with its characteristics
type Pokemon struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	SpritePath  string `json:"sprite_path"`
	BaseStats   Stats  `json:"base_stats"`
	Description string `json:"description"`
}

// Stats represents Pokemon base statistics
type Stats struct {
	HP      int `json:"hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
	Special int `json:"special"`
}

// CoffeePokemon represents the mapping between a coffee and its Pokemon
type CoffeePokemon struct {
	ID                string          `json:"id"`
	CoffeeID          string          `json:"coffee_id"`
	PokemonID         int             `json:"pokemon_id"`
	PokemonName       string          `json:"pokemon_name"`
	Nickname          string          `json:"nickname"`
	Level             int             `json:"level"`
	MappingConfidence float64         `json:"mapping_confidence"`
	LLMDescription    string          `json:"llm_description"`
	TraitMapping      []TraitMapping  `json:"trait_mapping"`
	CreatedAt         time.Time       `json:"created_at"`
}

// TraitMapping represents how a coffee trait maps to Pokemon characteristics
type TraitMapping struct {
	Trait      string `json:"trait"`
	PokemonStat string `json:"pokemon_stat"`
	Reasoning  string `json:"reasoning"`
}

// LLMMappingRequest represents the request sent to LLM for Pokemon mapping
type LLMMappingRequest struct {
	CoffeeName    string        `json:"coffee_name"`
	Origin        string        `json:"origin"`
	TastingTraits TastingTraits `json:"tasting_traits"`
	TastingNotes  []string      `json:"tasting_notes"`
	Candidates    []Pokemon     `json:"candidates"`
}

// LLMMappingResponse represents the LLM response for Pokemon mapping
type LLMMappingResponse struct {
	SelectedPokemon string        `json:"selected_pokemon"`
	Confidence      float64       `json:"confidence"`
	Description     string        `json:"description"`
	TraitMapping    []TraitMapping `json:"trait_mapping"`
}

// PokemonMappingRequest represents a request to generate Pokemon for a coffee
type PokemonMappingRequest struct {
	CoffeeID string `json:"coffee_id"`
}

// PokemonMappingResponse represents the response for Pokemon mapping
type PokemonMappingResponse struct {
	Success bool           `json:"success"`
	Data    CoffeePokemon  `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}