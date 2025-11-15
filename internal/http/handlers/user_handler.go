package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
)

/*

User handler for user management and review tracking.
Handles user activation status and review history retrieval.

*/

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)

	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {

	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	prs, err := h.userService.GetUserReviews(r.Context(), userID)

	if err != nil {
		handleServiceError(w, err)
		return
	}

	// PullRequest -> PRShort
	type PRShort struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	}

	shortPRs := make([]PRShort, len(prs))

	for i, pr := range prs {
		shortPRs[i] = PRShort{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          string(pr.Status),
		}
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"user_id":       userID,
		"pull_requests": shortPRs,
	})
}
