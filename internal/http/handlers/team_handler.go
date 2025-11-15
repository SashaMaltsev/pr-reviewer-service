package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
)


type TeamHandler struct {
    teamService *service.TeamService
}


func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
    return &TeamHandler{teamService: teamService}
}


func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
    var req struct {
        TeamName string                `json:"team_name"`
        Members  []models.TeamMember   `json:"members"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
        return
    }

    team, err := h.teamService.CreateTeam(r.Context(), req.TeamName, req.Members)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    respondJSON(w, http.StatusCreated, map[string]any{"team": team})
}


func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
    teamName := r.URL.Query().Get("team_name")
    if teamName == "" {
        respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
        return
    }

    team, err := h.teamService.GetTeam(r.Context(), teamName)
    if err != nil {
        handleServiceError(w, err)
        return
    }

    respondJSON(w, http.StatusOK, team)
}
