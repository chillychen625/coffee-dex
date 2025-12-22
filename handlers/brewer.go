package handlers

import (
	"encoding/json"
	"fmt"
	"go-coffee-log/service"
	"log"
	"net/http"
	"strings"
)

// BrewerHandler handles HTTP requests for brewer operations
type BrewerHandler struct {
	brewerService *service.BrewerService
}

// NewBrewerHandler creates a new brewer handler
func NewBrewerHandler(brewerService *service.BrewerService) *BrewerHandler {
	return &BrewerHandler{
		brewerService: brewerService,
	}
}

// CreateBrewer handles POST /brewers
func (h *BrewerHandler) CreateBrewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		PokeballType string `json:"pokeball_type"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: CreateBrewer decode failed: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Check brewer limit
	if err := h.brewerService.ValidateBrewerLimit(); err != nil {
		log.Printf("ERROR: ValidateBrewerLimit failed: %v", err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	brewer, err := h.brewerService.CreateBrewer(req.Name, req.PokeballType)
	if err != nil {
		log.Printf("ERROR: CreateBrewer failed: %v", err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	log.Printf("INFO: Created brewer: %s (ID: %s)", brewer.Name, brewer.ID)
	respondJSON(w, http.StatusCreated, brewer)
}

// GetAllBrewers handles GET /brewers
func (h *BrewerHandler) GetAllBrewers(w http.ResponseWriter, r *http.Request) {
	brewers, err := h.brewerService.GetAllBrewers()
	if err != nil {
		log.Printf("ERROR: GetAllBrewers failed: %v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get brewers: %v", err))
		return
	}
	
	respondJSON(w, http.StatusOK, brewers)
}


// DeleteBrewer handles DELETE /brewers/{id}
func (h *BrewerHandler) DeleteBrewer(w http.ResponseWriter, r *http.Request) {
	brewerID := r.PathValue("id")
	
	if err := h.brewerService.DeleteBrewer(brewerID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("ERROR: DeleteBrewer - brewer not found: %s", brewerID)
			respondError(w, http.StatusNotFound, "Brewer not found")
		} else {
			log.Printf("ERROR: DeleteBrewer failed for ID %s: %v", brewerID, err)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete brewer: %v", err))
		}
		return
	}
	
	log.Printf("INFO: Deleted brewer: %s", brewerID)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Brewer deleted"})
}


// GetAvailablePokeballTypes handles GET /brewers/pokeball-types
func (h *BrewerHandler) GetAvailablePokeballTypes(w http.ResponseWriter, r *http.Request) {
	types := h.brewerService.GetAvailablePokeballTypes()
	respondJSON(w, http.StatusOK, types)
}

// AddStandaloneRecipe handles POST /brewers/{id}/standalone-recipes
func (h *BrewerHandler) AddStandaloneRecipe(w http.ResponseWriter, r *http.Request) {
	brewerID := r.PathValue("id")
	
	var req struct {
		Name  string   `json:"name"`
		Steps []string `json:"steps"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if err := h.brewerService.AddStandaloneRecipe(brewerID, req.Name, req.Steps); err != nil {
		if strings.Contains(err.Error(), "maximum") {
			respondError(w, http.StatusBadRequest, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to add recipe")
		}
		return
	}
	
	respondJSON(w, http.StatusCreated, map[string]string{"message": "Recipe added to brewer"})
}

// RemoveStandaloneRecipe handles DELETE /brewers/{id}/standalone-recipes/{recipe_id}
func (h *BrewerHandler) RemoveStandaloneRecipe(w http.ResponseWriter, r *http.Request) {
	brewerID := r.PathValue("id")
	recipeID := r.PathValue("recipe_id")
	
	if err := h.brewerService.RemoveStandaloneRecipe(brewerID, recipeID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Recipe not found for this brewer")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to remove recipe")
		}
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Recipe removed from brewer"})
}