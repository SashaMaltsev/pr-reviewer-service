package models

import (
	"slices"
	"time"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"created_at"`
	MergedAt          *time.Time `json:"merged_at,omitempty"`
}

func NewPullRequest(prID, prName, authorID string) *PullRequest {
	return &PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            PRStatusOpen,
		AssignedReviewers: []string{},
		CreatedAt:         time.Now(),
	}
}

func (pr *PullRequest) Merge() {
	pr.Status = PRStatusMerged
	now := time.Now()
	pr.MergedAt = &now
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == PRStatusMerged
}

func (pr *PullRequest) AddReviewer(userID string) {
	pr.AssignedReviewers = append(pr.AssignedReviewers, userID)
}

func (pr *PullRequest) RemoveReviewer(userID string) bool {
	if i := slices.Index(pr.AssignedReviewers, userID); i != -1 {
		pr.AssignedReviewers = slices.Delete(pr.AssignedReviewers, i, i+1)
		return true
	}
	return false
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	return slices.Contains(pr.AssignedReviewers, userID)
}
