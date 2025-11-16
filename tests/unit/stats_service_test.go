package unit

import (
	"context"
	"testing"

	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestStatsService_GetAssignmentStats_Success(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)

	statsService := service.NewStatsService(mockPRRepo, mockUserRepo)

	// Mock data
	totalStats := map[string]int{
		"user1": 5,
		"user2": 3,
	}

	activeStats := map[string]int{
		"user1": 2,
		"user2": 1,
	}

	users := map[string]*models.User{
		"user1": {UserID: "user1", Username: "alice"},
		"user2": {UserID: "user2", Username: "bob"},
	}

	// Setup expectations
	mockPRRepo.On("GetAssignmentStats", ctx).Return(totalStats, nil)
	mockUserRepo.On("GetReviewerLoad", ctx, []string{"user1", "user2"}).Return(activeStats, nil)
	mockUserRepo.On("GetByID", ctx, "user1").Return(users["user1"], nil)
	mockUserRepo.On("GetByID", ctx, "user2").Return(users["user2"], nil)

	// Execute
	result, err := statsService.GetAssignmentStats(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Find and verify user1 stats
	var user1Stats, user2Stats service.AssignmentStats

	for _, stat := range result {
		switch stat.UserID {
		case "user1":
			user1Stats = stat
		case "user2":
			user2Stats = stat
		}
	}

	assert.Equal(t, "user1", user1Stats.UserID)
	assert.Equal(t, "alice", user1Stats.Username)
	assert.Equal(t, 5, user1Stats.TotalAssigned)
	assert.Equal(t, 2, user1Stats.ActiveReviews)

	assert.Equal(t, "user2", user2Stats.UserID)
	assert.Equal(t, "bob", user2Stats.Username)
	assert.Equal(t, 3, user2Stats.TotalAssigned)
	assert.Equal(t, 1, user2Stats.ActiveReviews)

	mockPRRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestStatsService_GetAssignmentStats_UserNotFound(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)

	statsService := service.NewStatsService(mockPRRepo, mockUserRepo)

	// Mock data
	totalStats := map[string]int{
		"user1": 5,
		"user2": 3,
		"user3": 2, // This user will not be found
	}

	activeStats := map[string]int{
		"user1": 2,
		"user2": 1,
		"user3": 0,
	}

	users := map[string]*models.User{
		"user1": {UserID: "user1", Username: "alice"},
		"user2": {UserID: "user2", Username: "bob"},
		// user3 is missing - will return error
	}

	// Setup expectations
	mockPRRepo.On("GetAssignmentStats", ctx).Return(totalStats, nil)
	mockUserRepo.On("GetReviewerLoad", ctx, []string{"user1", "user2", "user3"}).Return(activeStats, nil)
	mockUserRepo.On("GetByID", ctx, "user1").Return(users["user1"], nil)
	mockUserRepo.On("GetByID", ctx, "user2").Return(users["user2"], nil)
	mockUserRepo.On("GetByID", ctx, "user3").Return(nil, assert.AnError) // user3 not found

	// Execute
	result, err := statsService.GetAssignmentStats(ctx)

	// Assert
	assert.NoError(t, err)   // User not found is handled gracefully with continue
	assert.Len(t, result, 2) // Only 2 users returned, user3 is skipped

	// Verify only found users are in result
	userIDs := make(map[string]bool)
	for _, stat := range result {
		userIDs[stat.UserID] = true
	}

	assert.True(t, userIDs["user1"])
	assert.True(t, userIDs["user2"])
	assert.False(t, userIDs["user3"])

	mockPRRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestStatsService_GetAssignmentStats_EmptyResults(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)

	statsService := service.NewStatsService(mockPRRepo, mockUserRepo)

	// Mock empty data
	totalStats := map[string]int{}
	activeStats := map[string]int{}

	// Setup expectations
	mockPRRepo.On("GetAssignmentStats", ctx).Return(totalStats, nil)
	mockUserRepo.On("GetReviewerLoad", ctx, []string{}).Return(activeStats, nil)

	// Execute
	result, err := statsService.GetAssignmentStats(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)

	mockPRRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "GetByID")
}
