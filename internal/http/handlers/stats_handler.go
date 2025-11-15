package handler

import (
	"net/http"

	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
)

/*

Stats handler for retrieving assignment statistics.
Provides PR assignment metrics and team activity data.

*/

type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetAssignmentStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetAssignmentStats(r.Context())

	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get stats")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"stats": stats})
}
