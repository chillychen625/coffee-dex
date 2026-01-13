package handlers

import (
	"encoding/json"
	"go-coffee-log/models"
	"go-coffee-log/service"
	"log"
	"net/http"
)

// BrewHandler handles HTTP requests for brew operations
type BrewHandler struct {
	brewService    *service.BrewService
	pokemonService *service.PokemonService
}

// NewBrewHandler creates a new brew handler
func NewBrewHandler(brewService *service.BrewService, pokemonService *service.PokemonService) *BrewHandler {
	return &BrewHandler{
		brewService:    brewService,
		pokemonService: pokemonService,
	}
}

// CreateBrew handles POST /brews
func (h *BrewHandler) CreateBrew(w http.ResponseWriter, r *http.Request) {
	var brew models.Brew
	if err := json.NewDecoder(r.Body).Decode(&brew); err != nil {
		log.Printf("Error decoding brew: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	created, err := h.brewService.CreateBrew(brew)
	if err != nil {
		log.Printf("Error creating brew: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

// GetBrew handles GET /brews/{id}
func (h *BrewHandler) GetBrew(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	brew, err := h.brewService.GetBrew(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Brew not found")
		return
	}

	respondJSON(w, http.StatusOK, brew)
}

// ListBrews handles GET /brews
func (h *BrewHandler) ListBrews(w http.ResponseWriter, r *http.Request) {
	brews, err := h.brewService.ListBrews()
	if err != nil {
		log.Printf("Error listing brews: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list brews")
		return
	}

	respondJSON(w, http.StatusOK, brews)
}

// GetRecentBrews handles GET /brews/recent
func (h *BrewHandler) GetRecentBrews(w http.ResponseWriter, r *http.Request) {
	brews, err := h.brewService.GetRecentBrews(10)
	if err != nil {
		log.Printf("Error getting recent brews: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get recent brews")
		return
	}

	respondJSON(w, http.StatusOK, brews)
}

// GetBrewsForCoffee handles GET /coffees/{coffee_id}/brews
func (h *BrewHandler) GetBrewsForCoffee(w http.ResponseWriter, r *http.Request) {
	coffeeID := r.PathValue("coffee_id")

	brews, err := h.brewService.GetBrewsForCoffee(coffeeID)
	if err != nil {
		log.Printf("Error getting brews for coffee %s: %v", coffeeID, err)
		respondError(w, http.StatusInternalServerError, "Failed to get brews")
		return
	}

	respondJSON(w, http.StatusOK, brews)
}

// GetBrewProgress handles GET /coffees/{coffee_id}/brew-progress
func (h *BrewHandler) GetBrewProgress(w http.ResponseWriter, r *http.Request) {
	coffeeID := r.PathValue("coffee_id")

	// Check if coffee has a Pokemon already
	hasPokemon := false
	if h.pokemonService != nil {
		hasPokemon = h.pokemonService.HasPokemon(coffeeID)
	}

	progress, err := h.brewService.GetBrewProgress(coffeeID, hasPokemon)
	if err != nil {
		log.Printf("Error getting brew progress for coffee %s: %v", coffeeID, err)
		respondError(w, http.StatusInternalServerError, "Failed to get brew progress")
		return
	}

	respondJSON(w, http.StatusOK, progress)
}

// DeleteBrew handles DELETE /brews/{id}
func (h *BrewHandler) DeleteBrew(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.brewService.DeleteBrew(id); err != nil {
		respondError(w, http.StatusNotFound, "Brew not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
