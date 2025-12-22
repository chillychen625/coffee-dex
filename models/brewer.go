package models

import (
	"fmt"
	"time"
)

// Recipe represents a standalone brewing recipe
type Recipe struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

// Brewer represents a coffee brewer with associated pokeball sprite
type Brewer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PokeballType string   `json:"pokeball_type"` // "poke-ball", "great-ball", "ultra-ball", "fast-ball"
	Recipes     []Recipe  `json:"recipes"`       // Up to 4 standalone recipes
	CreatedAt   time.Time `json:"created_at"`
}


// Validate validates the brewer data
func (b *Brewer) Validate() error {
	if b.Name == "" {
		return fmt.Errorf("brewer name cannot be empty")
	}
	
	validPokeballs := map[string]bool{
		"poke-ball":  true,
		"great-ball": true,
		"ultra-ball": true,
		"fast-ball":  true,
	}
	
	if !validPokeballs[b.PokeballType] {
		return fmt.Errorf("invalid pokeball type: %s", b.PokeballType)
	}
	
	return nil
}