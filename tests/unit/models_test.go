package unit

import (
	"testing"
	"time"

	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewPullRequest(t *testing.T) {

	pr := models.NewPullRequest("pr-1", "Test PR", "u1")

	assert.Equal(t, "pr-1", pr.PullRequestID)
	assert.Equal(t, "Test PR", pr.PullRequestName)
	assert.Equal(t, "u1", pr.AuthorID)
	assert.Equal(t, models.PRStatusOpen, pr.Status)
	assert.Empty(t, pr.AssignedReviewers)
	assert.Nil(t, pr.MergedAt)
}

func TestPullRequest_Merge(t *testing.T) {

	pr := models.NewPullRequest("pr-1", "Test PR", "u1")

	assert.False(t, pr.IsMerged())

	pr.Merge()

	assert.True(t, pr.IsMerged())
	assert.Equal(t, models.PRStatusMerged, pr.Status)
	assert.NotNil(t, pr.MergedAt)
}

func TestPullRequest_AddRemoveReviewer(t *testing.T) {

	pr := models.NewPullRequest("pr-1", "Test PR", "u1")

	pr.AddReviewer("u2")
	pr.AddReviewer("u3")

	assert.Equal(t, 2, len(pr.AssignedReviewers))
	assert.True(t, pr.HasReviewer("u2"))
	assert.True(t, pr.HasReviewer("u3"))
	assert.False(t, pr.HasReviewer("u4"))

	removed := pr.RemoveReviewer("u2")
	assert.True(t, removed)
	assert.Equal(t, 1, len(pr.AssignedReviewers))
	assert.False(t, pr.HasReviewer("u2"))

	removed = pr.RemoveReviewer("u99")
	assert.False(t, removed)
}

func TestUser_SetActive(t *testing.T) {

	user := models.NewUser("u1", "Alice", "backend", true)
	initialTime := user.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	user.SetActive(false)

	assert.False(t, user.IsActive)
	assert.True(t, user.UpdatedAt.After(initialTime))
}
