package unit

import (
	"context"
	"testing"

	apperrors "github.com/SashaMalcev/pr-reviewer-service/internal/errors"
	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	"github.com/SashaMalcev/pr-reviewer-service/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestUserService_SetIsActive_Success(t *testing.T) {

	ctx := context.Background()
	mockUserRepo := new(MockUserRepo)
	mockPRRepo := new(MockPRRepo)

	service := service.NewUserService(mockUserRepo, mockPRRepo)

	existingUser := &models.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	mockUserRepo.On("GetByID", ctx, "u1").Return(existingUser, nil)
	mockUserRepo.On("Update", ctx, existingUser).Return(nil)

	user, err := service.SetIsActive(ctx, "u1", false)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.False(t, user.IsActive)

	mockUserRepo.AssertExpectations(t)
}

func TestUserService_SetIsActive_UserNotFound(t *testing.T) {

	ctx := context.Background()
	mockUserRepo := new(MockUserRepo)
	mockPRRepo := new(MockPRRepo)

	service := service.NewUserService(mockUserRepo, mockPRRepo)

	mockUserRepo.On("GetByID", ctx, "u99").Return(nil, apperrors.ErrUserNotFound)

	user, err := service.SetIsActive(ctx, "u99", false)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, apperrors.ErrUserNotFound, err)
}
