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

// Mock repositories
type MockPRRepo struct {
	mock.Mock
}

func (m *MockPRRepo) Create(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepo) Update(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepo) GetByID(ctx context.Context, prID string) (*models.PullRequest, error) {
	args := m.Called(ctx, prID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *MockPRRepo) Exists(ctx context.Context, prID string) (bool, error) {
	args := m.Called(ctx, prID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPRRepo) GetByReviewer(ctx context.Context, userID string) ([]*models.PullRequest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PullRequest), args.Error(1)
}

func (m *MockPRRepo) GetAssignmentStats(ctx context.Context) (map[string]int, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]int), args.Error(1)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetActiveByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*models.User, error) {
	args := m.Called(ctx, teamName, excludeUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepo) GetReviewerLoad(ctx context.Context, userIDs []string) (map[string]int, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).(map[string]int), args.Error(1)
}

type MockTeamRepo struct {
	mock.Mock
}

func (m *MockTeamRepo) Create(ctx context.Context, team *models.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepo) GetByName(ctx context.Context, teamName string) (*models.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepo) Exists(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}

func TestCreatePR_AssignsTwoReviewers(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	// Setup mocks
	author := &models.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	reviewers := []*models.User{
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true},
		{UserID: "u3", Username: "Charlie", TeamName: "backend", IsActive: true},
	}

	mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveByTeam", ctx, "backend", "u1").Return(reviewers, nil)
	mockUserRepo.On("GetReviewerLoad", ctx, []string{"u2", "u3"}).Return(map[string]int{"u2": 0, "u3": 0}, nil)
	mockPRRepo.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).Return(nil)

	// Execute
	pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 2, len(pr.AssignedReviewers))
	assert.NotContains(t, pr.AssignedReviewers, "u1") // Author not assigned

	mockPRRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestMergePR_Idempotent(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	// Already merged PR
	mergedPR := &models.PullRequest{
		PullRequestID: "pr-1",
		Status:        models.PRStatusMerged,
	}

	mockPRRepo.On("GetByID", ctx, "pr-1").Return(mergedPR, nil)

	// Execute
	pr, err := service.MergePR(ctx, "pr-1")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, models.PRStatusMerged, pr.Status)
	mockPRRepo.AssertNotCalled(t, "Update") // Should not update if already merged
}

func TestCreatePR_NoActiveCandidates(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	author := &models.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveByTeam", ctx, "backend", "u1").Return([]*models.User{}, nil)
	mockPRRepo.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).Return(nil)

	pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, 0, len(pr.AssignedReviewers))
}

func TestCreatePR_OnlyOneCandidate(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	author := &models.User{UserID: "u1", TeamName: "backend"}
	reviewers := []*models.User{
		{UserID: "u2", Username: "Bob", IsActive: true},
	}

	mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveByTeam", ctx, "backend", "u1").Return(reviewers, nil)
	mockUserRepo.On("GetReviewerLoad", ctx, []string{"u2"}).Return(map[string]int{"u2": 0}, nil)
	mockPRRepo.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).Return(nil)

	pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(pr.AssignedReviewers))
}

func TestCreatePR_LoadBalancing(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	author := &models.User{UserID: "u1", TeamName: "backend"}
	reviewers := []*models.User{
		{UserID: "u2", Username: "Bob", IsActive: true},
		{UserID: "u3", Username: "Charlie", IsActive: true},
		{UserID: "u4", Username: "Dave", IsActive: true},
	}

	mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveByTeam", ctx, "backend", "u1").Return(reviewers, nil)
	mockUserRepo.On("GetReviewerLoad", ctx, []string{"u2", "u3", "u4"}).Return(
		map[string]int{"u2": 5, "u3": 2, "u4": 2}, nil,
	)
	mockPRRepo.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).Return(nil)

	pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "u1")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(pr.AssignedReviewers))
	// u3 and u4 should be selected (lowest load)
	assert.NotContains(t, pr.AssignedReviewers, "u2") // Has highest load
}

func TestReassignReviewer_Success(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	openPR := &models.PullRequest{
		PullRequestID:     "pr-1",
		AuthorID:          "u1",
		Status:            models.PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}

	oldReviewer := &models.User{UserID: "u2", TeamName: "backend"}
	newCandidate := &models.User{UserID: "u4", Username: "Dave", IsActive: true}

	mockPRRepo.On("GetByID", ctx, "pr-1").Return(openPR, nil)
	mockUserRepo.On("GetByID", ctx, "u2").Return(oldReviewer, nil)
	mockUserRepo.On("GetActiveByTeam", ctx, "backend", "").Return(
		[]*models.User{newCandidate, oldReviewer}, nil,
	)
	mockUserRepo.On("GetReviewerLoad", ctx, []string{"u4"}).Return(map[string]int{"u4": 0}, nil)
	mockPRRepo.On("Update", ctx, mock.AnythingOfType("*models.PullRequest")).Return(nil)

	pr, replacedBy, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, "u4", replacedBy)
	assert.Contains(t, pr.AssignedReviewers, "u4")
	assert.NotContains(t, pr.AssignedReviewers, "u2")
}

func TestReassignReviewer_PRMerged(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	mergedPR := &models.PullRequest{
		PullRequestID: "pr-1",
		Status:        models.PRStatusMerged,
	}

	mockPRRepo.On("GetByID", ctx, "pr-1").Return(mergedPR, nil)

	pr, replacedBy, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Empty(t, replacedBy)
	assert.Equal(t, apperrors.ErrPRMerged, err)
}

func TestReassignReviewer_NotAssigned(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	openPR := &models.PullRequest{
		PullRequestID:     "pr-1",
		Status:            models.PRStatusOpen,
		AssignedReviewers: []string{"u3"},
	}

	mockPRRepo.On("GetByID", ctx, "pr-1").Return(openPR, nil)

	pr, replacedBy, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Empty(t, replacedBy)
	assert.Equal(t, apperrors.ErrNotAssigned, err)
}

func TestReassignReviewer_NoCandidate(t *testing.T) {
	ctx := context.Background()

	mockPRRepo := new(MockPRRepo)
	mockUserRepo := new(MockUserRepo)
	mockTeamRepo := new(MockTeamRepo)

	service := service.NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)

	openPR := &models.PullRequest{
		PullRequestID:     "pr-1",
		AuthorID:          "u1",
		Status:            models.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	oldReviewer := &models.User{UserID: "u2", TeamName: "backend"}

	mockPRRepo.On("GetByID", ctx, "pr-1").Return(openPR, nil)
	mockUserRepo.On("GetByID", ctx, "u2").Return(oldReviewer, nil)
	mockUserRepo.On("GetActiveByTeam", ctx, "backend", "").Return([]*models.User{oldReviewer}, nil)

	pr, replacedBy, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, pr)
	assert.Empty(t, replacedBy)
	assert.Equal(t, apperrors.ErrNoCandidate, err)
}
