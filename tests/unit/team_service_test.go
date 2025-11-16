package unit

import (
	"context"
	"testing"

	apperrors "github.com/SashaMalcev/pr-reviewer-service/internal/errors"
	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTeamService_CreateTeam_Success(t *testing.T) {

	ctx := context.Background()

	mockTeamRepo := new(MockTeamRepo)
	mockUserRepo := new(MockUserRepo)

	service := service.NewTeamService(mockTeamRepo, mockUserRepo)

	members := []models.TeamMember{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
	}

	mockTeamRepo.On("Exists", ctx, "backend").Return(false, nil)
	mockTeamRepo.On("Create", ctx, mock.AnythingOfType("*models.Team")).Return(nil)
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Twice()

	team, err := service.CreateTeam(ctx, "backend", members)

	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "backend", team.TeamName)
	assert.Equal(t, 2, len(team.Members))

	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamService_CreateTeam_AlreadyExists(t *testing.T) {

	ctx := context.Background()

	mockTeamRepo := new(MockTeamRepo)
	mockUserRepo := new(MockUserRepo)

	service := service.NewTeamService(mockTeamRepo, mockUserRepo)

	mockTeamRepo.On("Exists", ctx, "backend").Return(true, nil)

	team, err := service.CreateTeam(ctx, "backend", []models.TeamMember{})

	assert.Error(t, err)
	assert.Nil(t, team)
	assert.Equal(t, apperrors.ErrTeamExists, err)
}

func TestTeamService_GetTeam(t *testing.T) {

	ctx := context.Background()

	mockTeamRepo := new(MockTeamRepo)
	mockUserRepo := new(MockUserRepo)

	service := service.NewTeamService(mockTeamRepo, mockUserRepo)

	expectedTeam := &models.Team{
		TeamName: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	mockTeamRepo.On("GetByName", ctx, "backend").Return(expectedTeam, nil)

	team, err := service.GetTeam(ctx, "backend")

	assert.NoError(t, err)
	assert.Equal(t, expectedTeam, team)
}
