package handlers

import (
	"encoding/json"
	"go-coffee-log/models"
	"go-coffee-log/service"
	"log"
	"net/http"
)

// PokemonHandler handles HTTP requests for Pokemon operations
type PokemonHandler struct {
	pokemonService *service.PokemonService
	coffeeService  *service.CoffeeService
}

// NewPokemonHandler creates a new Pokemon handler
func NewPokemonHandler(pokemonService *service.PokemonService, coffeeService *service.CoffeeService) *PokemonHandler {
	return &PokemonHandler{
		pokemonService: pokemonService,
		coffeeService:  coffeeService,
	}
}

// GeneratePokemon handles POST /coffees/{id}/pokemon
func (h *PokemonHandler) GeneratePokemon(w http.ResponseWriter, r *http.Request) {
	coffeeID := r.PathValue("coffee_id")
	log.Printf("GeneratePokemon called for coffee ID: %s", coffeeID)
	
	// Get coffee from service
	coffee, err := h.coffeeService.GetCoffee(coffeeID)
	if err != nil {
		log.Printf("Error getting coffee: %v", err)
		respondError(w, http.StatusNotFound, "Coffee not found")
		return
	}
	
	// Generate Pokemon mapping
	mapping, err := h.pokemonService.MapCoffeeToPokemon(coffee)
	if err != nil {
		log.Printf("Error mapping coffee to Pokemon: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	log.Printf("Successfully generated Pokemon mapping: %+v", mapping)
	respondJSON(w, http.StatusCreated, mapping)
}

// GetCoffeePokemon handles GET /coffees/{id}/pokemon
func (h *PokemonHandler) GetCoffeePokemon(w http.ResponseWriter, r *http.Request) {
	coffeeID := r.PathValue("coffee_id")
	
	mapping, err := h.pokemonService.GetCoffeePokemon(coffeeID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Pokemon mapping not found")
		return
	}
	
	respondJSON(w, http.StatusOK, mapping)
}

// GetCoffeeDex handles GET /pokedex
func (h *PokemonHandler) GetCoffeeDex(w http.ResponseWriter, r *http.Request) {
	mappings, err := h.pokemonService.GetAllCoffeePokemon()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch CoffeeDex")
		return
	}
	
	respondJSON(w, http.StatusOK, mappings)
}

// UpdateNickname handles PUT /coffees/{id}/pokemon/nickname
func (h *PokemonHandler) UpdateNickname(w http.ResponseWriter, r *http.Request) {
	coffeeID := r.PathValue("coffee_id")
	
	var request struct {
		Nickname string `json:"nickname"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	if err := h.pokemonService.UpdateNickname(coffeeID, request.Nickname); err != nil {
		respondError(w, http.StatusNotFound, "Pokemon mapping not found")
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Nickname updated successfully"})
}

// GetPokemonStats handles GET /pokedex/stats
func (h *PokemonHandler) GetPokemonStats(w http.ResponseWriter, r *http.Request) {
	mappings, err := h.pokemonService.GetAllCoffeePokemon()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}
	
	stats := map[string]interface{}{
		"total_coffees": len(mappings),
		"pokemon_used":  len(mappings),
		"collection_complete": len(mappings) >= 151, // Gen 1 has 151 Pokemon
		"average_confidence": calculateAverageConfidence(mappings),
	}
	
	respondJSON(w, http.StatusOK, stats)
}

// Helper functions

func calculateAverageConfidence(mappings []models.CoffeePokemon) float64 {
	if len(mappings) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, mapping := range mappings {
		total += mapping.MappingConfidence
	}
	
	return total / float64(len(mappings))
}