package service

import (
	"context"

	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	repository "github.com/SashaMalcev/pr-reviewer-service/internal/repository/interfaces"
)


type UserService struct {
    userRepo    repository.UserRepository
    prRepo      repository.PRRepository
}


func NewUserService(userRepo repository.UserRepository, prRepo repository.PRRepository) *UserService {
    return &UserService{
        userRepo: userRepo,
        prRepo: prRepo,
    }
}


func(s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
    
	user, err := s.userRepo.GetByID(ctx, userID)
    
	if err != nil {
        return nil, err
    }

    user.SetActive(isActive)

    err = s.userRepo.Update(ctx, user); 
	
	if err != nil {
        return nil, err
    }

    return user, nil
}


func(s *UserService) GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequest, error) {

	// Verify user exists
    _, err := s.userRepo.GetByID(ctx, userID); 

	if err != nil {
        return nil, err
    }

    return s.prRepo.GetByReviewer(ctx, userID)
}