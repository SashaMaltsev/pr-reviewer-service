package integration

import (
	"context"
	"os"
	"testing"

	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	"github.com/SashaMalcev/pr-reviewer-service/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*

Integration tests for PostgreSQL repositories.
Tests database operations for teams, users and pull requests with real database.

*/

func getTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(t, dbURL, "TEST_DATABASE_URL must be set")

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	return pool
}

func cleanDB(t *testing.T, pool *pgxpool.Pool) {
	_, err := pool.Exec(context.Background(), `
        TRUNCATE TABLE pr_reviewers, pull_requests, users, teams CASCADE
    `)
	require.NoError(t, err)
}

func TestTeamRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := getTestDB(t)
	defer pool.Close()
	defer cleanDB(t, pool)

	ctx := context.Background()
	repo := postgres.NewTeamRepository(pool)

	// Create team
	team := models.NewTeam("backend", []models.TeamMember{})
	err := repo.Create(ctx, team)
	assert.NoError(t, err)

	// Get team
	retrieved, err := repo.GetByName(ctx, "backend")
	assert.NoError(t, err)
	assert.Equal(t, "backend", retrieved.TeamName)

	// Check exists
	exists, err := repo.Exists(ctx, "backend")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := getTestDB(t)
	defer pool.Close()
	defer cleanDB(t, pool)

	ctx := context.Background()
	teamRepo := postgres.NewTeamRepository(pool)
	userRepo := postgres.NewUserRepository(pool)

	// Create team first
	team := models.NewTeam("backend", []models.TeamMember{})
	require.NoError(t, teamRepo.Create(ctx, team))

	// Create user
	user := models.NewUser("u1", "Alice", "backend", true)
	err := userRepo.Create(ctx, user)
	assert.NoError(t, err)

	// Get user
	retrieved, err := userRepo.GetByID(ctx, "u1")
	assert.NoError(t, err)
	assert.Equal(t, "Alice", retrieved.Username)
	assert.True(t, retrieved.IsActive)

	// Update user
	user.SetActive(false)
	err = userRepo.Update(ctx, user)
	assert.NoError(t, err)

	// Verify update
	retrieved, err = userRepo.GetByID(ctx, "u1")
	assert.NoError(t, err)
	assert.False(t, retrieved.IsActive)
}

func TestPRRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	pool := getTestDB(t)
	defer pool.Close()
	defer cleanDB(t, pool)

	ctx := context.Background()
	teamRepo := postgres.NewTeamRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	prRepo := postgres.NewPRRepository(pool)

	// Setup
	team := models.NewTeam("backend", []models.TeamMember{})
	require.NoError(t, teamRepo.Create(ctx, team))

	author := models.NewUser("u1", "Alice", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author))

	reviewer := models.NewUser("u2", "Sasha", "backend", true)
	require.NoError(t, userRepo.Create(ctx, reviewer))

	// Create PR
	pr := models.NewPullRequest("pr-1", "Test PR", "u1")
	pr.AddReviewer("u2")
	err := prRepo.Create(ctx, pr)
	assert.NoError(t, err)

	// Get PR
	retrieved, err := prRepo.GetByID(ctx, "pr-1")
	assert.NoError(t, err)
	assert.Equal(t, "Test PR", retrieved.PullRequestName)
	assert.Equal(t, 1, len(retrieved.AssignedReviewers))
	assert.Contains(t, retrieved.AssignedReviewers, "u2")

	// Get by reviewer
	prs, err := prRepo.GetByReviewer(ctx, "u2")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(prs))
}
