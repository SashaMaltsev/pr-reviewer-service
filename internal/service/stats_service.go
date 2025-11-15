package service

import (
	"context"

	repository "github.com/SashaMalcev/pr-reviewer-service/internal/repository/interfaces"
)


type StatsService struct {
    prRepo   repository.PRRepository
    userRepo repository.UserRepository
}


func NewStatsService(prRepo repository.PRRepository, userRepo repository.UserRepository) *StatsService {
    return &StatsService{
        prRepo:   prRepo,
        userRepo: userRepo,
    }
}


type AssignmentStats struct {
    UserID         string `json:"user_id"`
    Username       string `json:"username"`
    TotalAssigned  int    `json:"total_assigned"`
    ActiveReviews  int    `json:"active_reviews"`
}


func(s *StatsService) GetAssignmentStats(ctx context.Context) ([]AssignmentStats, error) {
    
	// Get total assignments per user
    totalStats, err := s.prRepo.GetAssignmentStats(ctx)
   
	if err != nil {
        return nil, err
    }

    // Get user IDs
    userIDs := make([]string, 0, len(totalStats))
    
	for userID := range totalStats {
        userIDs = append(userIDs, userID)
    }

    // Get active review counts
    activeStats, err := s.userRepo.GetReviewerLoad(ctx, userIDs)
    
	if err != nil {
        return nil, err
    }

    // Build response
    stats := []AssignmentStats{} 

    for userID, total := range totalStats {
        
		user, err := s.userRepo.GetByID(ctx, userID)
        
		if err != nil {
            continue
        }

        stats = append(stats, AssignmentStats{
            UserID:        userID,
            Username:      user.Username,
            TotalAssigned: total,
            ActiveReviews: activeStats[userID],
        })
    }

    return stats, nil
}
