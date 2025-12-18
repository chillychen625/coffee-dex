package handlers

import (
	"encoding/json"
	"go-coffee-log/models"
	"go-coffee-log/service"
	"net/http"
)

// CoffeeHandler handles HTTP requests for coffee operations
// TODO: Add the following field:
//   - service (*service.CoffeeService) - the service layer to use
type CoffeeHandler struct {
	service *service.CoffeeService
}

// NewCoffeeHandler creates a new coffee handler
func NewCoffeeHandler(service *service.CoffeeService) *CoffeeHandler {
	return &CoffeeHandler{
		service: service,
	}
}

// CreateCoffee handles POST /coffees
// TODO: Implement this method
// Requirements:
//   - Decode JSON from request body
//   - Call service.CreateCoffee
//   - Return 201 Created with the created coffee
//   - Handle errors appropriately
// HINT: Use json.NewDecoder(r.Body).Decode() to parse JSON
// HINT: Use w.WriteHeader(http.StatusCreated) for 201 status
func (h *CoffeeHandler) CreateCoffee(w http.ResponseWriter, r *http.Request) {
	var coffee models.Coffee
	err := json.NewDecoder(r.Body).Decode(&coffee)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	
	createdCoffee, err := h.service.CreateCoffee(coffee)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	respondJSON(w, http.StatusCreated, createdCoffee)
}

// GetCoffee handles GET /coffees/{id}
// TODO: Implement this method
// Requirements:
//   - Extract ID from URL path
//   - Call service.GetCoffee
//   - Return 200 OK with the coffee
//   - Return 404 Not Found if coffee doesn't exist
// HINT: You'll need to extract the ID from the URL - we'll set this up in main.go
func (h *CoffeeHandler) GetCoffee(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	coffee, err := h.service.GetCoffee(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Coffee not found")
		return
	}
	respondJSON(w, http.StatusOK, coffee)
}

// ListCoffees handles GET /coffees
// TODO: Implement this method
// Requirements:
//   - Call service.ListCoffees
//   - Return 200 OK with array of coffees
// HINT: Even if no coffees exist, return an empty array []
func (h *CoffeeHandler) ListCoffees(w http.ResponseWriter, r *http.Request) {
	coffees, err := h.service.ListCoffees()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list coffees")
		return
	}
	
	if coffees == nil {
		coffees = []models.Coffee{}
	}
	
	respondJSON(w, http.StatusOK, coffees)
}

// GetRecentCoffees handles GET /coffees/recent
func (h *CoffeeHandler) GetRecentCoffees(w http.ResponseWriter, r *http.Request) {
	// Default to 10 recent coffees
	limit := 10
	
	coffees, err := h.service.GetRecentCoffees(limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get recent coffees")
		return
	}
	
	if coffees == nil {
		coffees = []models.Coffee{}
	}
	
	respondJSON(w, http.StatusOK, coffees)
}

// UpdateCoffee handles PUT /coffees/{id}
// TODO: Implement this method
// Requirements:
//   - Extract ID from URL
//   - Decode JSON from request body
//   - Call service.UpdateCoffee
//   - Return 200 OK with updated coffee
func (h *CoffeeHandler) UpdateCoffee(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path parameter
	id := r.PathValue("id")  // ← Use PathValue instead of manual parsing
	
	var coffee models.Coffee
	err := json.NewDecoder(r.Body).Decode(&coffee)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return  // ← Added missing return
	}
	defer r.Body.Close()
	
	updatedCoffee, err := h.service.UpdateCoffee(id, coffee)  // ← Renamed variable to avoid shadowing
	if err != nil {
		respondError(w, http.StatusNotFound, "Coffee not found")  // ← Better status code
		return  // ← Added missing return
	}
	respondJSON(w, http.StatusOK, updatedCoffee)  // ← Changed to StatusOK (200)
}

// DeleteCoffee handles DELETE /coffees/{id}
// TODO: Implement this method
// Requirements:
//   - Extract ID from URL
//   - Call service.DeleteCoffee
//   - Return 204 No Content on success
func (h *CoffeeHandler) DeleteCoffee(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path parameter
	id := r.PathValue("id")  // ← Use PathValue instead of manual parsing
	
	err := h.service.DeleteCoffee(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Coffee not found")  // ← Better status code
		return  // ← Added missing return
	}
	
	// For 204 No Content, just set the status (no body)
	w.WriteHeader(http.StatusNoContent)  // ← Don't use respondJSON for 204
}

// respondJSON is a helper function to send JSON responses
// TODO: Implement this helper method
// Requirements:
//   - Set Content-Type header to "application/json"
//   - Set status code
//   - Encode data as JSON
// HINT: Use json.NewEncoder(w).Encode(data)
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	// IMPORTANT: Headers MUST be set BEFORE WriteHeader
	w.Header().Set("Content-Type", "application/json")  // ← Set header first
	w.WriteHeader(status)                                // ← Then status
	json.NewEncoder(w).Encode(data)                     // ← Then encode body
}

// respondError is a helper function to send error responses
// TODO: Implement this helper method
// Requirements:
//   - Create a JSON object with an "error" field
//   - Use respondJSON to send it
// HINT: You can create an anonymous struct like: struct{Error string `json:"error"`}{Error: message}
func respondError(w http.ResponseWriter, status int, message string) {
	errorResponse := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}
	respondJSON(w, status, errorResponse)
}