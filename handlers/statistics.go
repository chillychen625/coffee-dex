package handlers

import (
	"go-coffee-log/service"
	"net/http"
)

// StatisticsHandler handles HTTP requests for statistics operations
type StatisticsHandler struct {
	statsService *service.StatisticsService
}

// NewStatisticsHandler creates a new statistics handler
func NewStatisticsHandler(statsService *service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statsService: statsService,
	}
}

// GetStatistics handles GET /statistics
func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.CalculateStatistics()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to calculate statistics")
		return
	}
	
	respondJSON(w, http.StatusOK, stats)
}