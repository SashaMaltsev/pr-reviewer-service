package handler

import (
	"encoding/json"
	"log"
	"net/http"

	apperrors "github.com/SashaMalcev/pr-reviewer-service/internal/errors"
)

/*

HTTP response utilities with error handling.
Includes JSON response helpers and service error to HTTP status mapping.

*/

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)

	if err != nil {
        log.Printf("Error encoding response: %v", err)
    }
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func handleServiceError(w http.ResponseWriter, err error) {

	code := apperrors.GetErrorCode(err)

	var status int

	switch code {
	case apperrors.CodeTeamExists:
		status = http.StatusBadRequest
	case apperrors.CodePRExists:
		status = http.StatusConflict
	case apperrors.CodePRMerged, apperrors.CodeNotAssigned, apperrors.CodeNoCandidate:
		status = http.StatusConflict
	case apperrors.CodeNotFound:
		status = http.StatusNotFound
	default:
		status = http.StatusInternalServerError
		code = "INTERNAL_ERROR"
	}

	respondError(w, status, string(code), err.Error())
}
