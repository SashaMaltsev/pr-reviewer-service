package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
)

/*

PR handler for managing pull requests.
Handles PR creation, merging and reviewer reassignment with proper error handling.

*/

type PRHandler struct {
	prService *service.PRService
}

func NewPRHandler(prService *service.PRService) *PRHandler {
	return &PRHandler{prService: prService}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {

	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)

	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"pr": pr})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {

	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)

	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"pr": pr})
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {

	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	pr, replacedBy, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)

	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"pr":          pr,
		"replaced_by": replacedBy,
	})
}
