package handler

import (
	"net/http"
)

/*

Health check handler for monitoring service status.
Returns 200 OK with healthy status for load balancers and health checks.

*/

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
