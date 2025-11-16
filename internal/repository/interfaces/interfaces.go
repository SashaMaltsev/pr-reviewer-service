package repository

import (
	"context"

	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
)

/*

Repository interfaces for data access layer.
Defines contracts for team, user and pull request data operations.

*/

// TeamRepository defines the interface for team-related data operations
type TeamRepository interface {
	Create(ctx context.Context, team *models.Team) error
	GetByName(ctx context.Context, teamName string) (*models.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

// UserRepository defines the interface for user-related data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetActiveByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*models.User, error)
	GetReviewerLoad(ctx context.Context, userIDs []string) (map[string]int, error)
}

// PRRepository defines the interface for pull request-related data operations
type PRRepository interface {
	Create(ctx context.Context, pr *models.PullRequest) error
	Update(ctx context.Context, pr *models.PullRequest) error
	GetByID(ctx context.Context, prID string) (*models.PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
	GetByReviewer(ctx context.Context, userID string) ([]*models.PullRequest, error)
	GetAssignmentStats(ctx context.Context) (map[string]int, error)
}
