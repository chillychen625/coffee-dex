package handlers

import (
	"encoding/json"
	"go-coffee-log/service"
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
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Check brewer limit
	if err := h.brewerService.ValidateBrewerLimit(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	brewer, err := h.brewerService.CreateBrewer(req.Name, req.PokeballType)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	respondJSON(w, http.StatusCreated, brewer)
}

// GetAllBrewers handles GET /brewers
func (h *BrewerHandler) GetAllBrewers(w http.ResponseWriter, r *http.Request) {
	brewers, err := h.brewerService.GetAllBrewers()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get brewers")
		return
	}
	
	respondJSON(w, http.StatusOK, brewers)
}

// GetAllBrewersWithRecipes handles GET /brewers/with-recipes
func (h *BrewerHandler) GetAllBrewersWithRecipes(w http.ResponseWriter, r *http.Request) {
	brewers, err := h.brewerService.GetAllBrewersWithRecipes()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get brewers with recipes")
		return
	}
	
	respondJSON(w, http.StatusOK, brewers)
}

// GetBrewerWithRecipes handles GET /brewers/{id}/recipes
func (h *BrewerHandler) GetBrewerWithRecipes(w http.ResponseWriter, r *http.Request) {
	brewerID := r.PathValue("id")
	
	brewer, err := h.brewerService.GetBrewerWithRecipes(brewerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Brewer not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to get brewer")
		}
		return
	}
	
	respondJSON(w, http.StatusOK, brewer)
}

// DeleteBrewer handles DELETE /brewers/{id}
func (h *BrewerHandler) DeleteBrewer(w http.ResponseWriter, r *http.Request) {
	brewerID := r.PathValue("id")
	
	if err := h.brewerService.DeleteBrewer(brewerID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Brewer not found")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to delete brewer")
		}
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Brewer deleted"})
}

// AddRecipeToBrewer handles POST /brewers/{id}/recipes
func (h *BrewerHandler) AddRecipeToBrewer(w http.ResponseWriter, r *http.Request) {
	brewerID := r.PathValue("id")
	
	var req struct {
		CoffeeID string `json:"coffee_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if err := h.brewerService.AddRecipeToBrewer(brewerID, req.CoffeeID); err != nil {
		if strings.Contains(err.Error(), "maximum") {
			respondError(w, http.StatusBadRequest, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to add recipe")
		}
		return
	}
	
	respondJSON(w, http.StatusCreated, map[string]string{"message": "Recipe added to brewer"})
}

// RemoveRecipeFromBrewer handles DELETE /brewers/{id}/recipes/{coffee_id}
func (h *BrewerHandler) RemoveRecipeFromBrewer(w http.ResponseWriter, r *http.Request) {
	brewerID := r.PathValue("id")
	coffeeID := r.PathValue("coffee_id")
	
	if err := h.brewerService.RemoveRecipeFromBrewer(brewerID, coffeeID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Recipe not found for this brewer")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to remove recipe")
		}
		return
	}
	
	respondJSON(w, http.StatusOK, map[string]string{"message": "Recipe removed from brewer"})
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