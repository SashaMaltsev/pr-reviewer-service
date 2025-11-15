package service

import (
	"context"

	apperrors "github.com/SashaMalcev/pr-reviewer-service/internal/errors"
	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	repository "github.com/SashaMalcev/pr-reviewer-service/internal/repository/interfaces"
)


type TeamService struct {
    teamRepo repository.TeamRepository
    userRepo repository.UserRepository
}


func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) *TeamService {
    return &TeamService{
        teamRepo: teamRepo,
        userRepo: userRepo,
    }
}


func(s *TeamService) CreateTeam(ctx context.Context, teamName string, members []models.TeamMember) (*models.Team, error) {

	// Check if team exists
    exists, err := s.teamRepo.Exists(ctx, teamName)
    
	if err != nil {
        return nil, err
    }

    if exists {
        return nil, apperrors.ErrTeamExists
    }

    // Create team
    team := models.NewTeam(teamName, members)

    err = s.teamRepo.Create(ctx, team);

	if err != nil {
        return nil, err
    }

    // Create/update users
    for _, member := range members {
        
		user := models.NewUser(member.UserID, member.Username, teamName, member.IsActive)

        err := s.userRepo.Create(ctx, user) 
		
		if err != nil {
            return nil, err
        }
    }

    return team, nil
}

func(s *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
    return s.teamRepo.GetByName(ctx, teamName)
}