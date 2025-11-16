package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*

End-to-end tests for PR Reviewer Service API.
Tests complete workflow from team creation to PR management with error scenarios.

*/

var baseURL string

func init() {
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
}

func TestE2E_CompleteWorkflow(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	// Wait for service to be ready
	waitForService(t)

	// Use timestamp to make test data unique
	timestamp := time.Now().Unix()
	teamName := fmt.Sprintf("e2e-team-%d", timestamp)
	u1 := fmt.Sprintf("e2e-u1-%d", timestamp)
	u2 := fmt.Sprintf("e2e-u2-%d", timestamp)
	u3 := fmt.Sprintf("e2e-u3-%d", timestamp)
	u4 := fmt.Sprintf("e2e-u3-%d", timestamp)
	prID := fmt.Sprintf("e2e-pr-1-%d", timestamp)

	// 1. Create team
	t.Log("Step 1: Creating team...")

	teamReq := map[string]any{
		"team_name": teamName,
		"members": []map[string]any{
			{"user_id": u1, "username": "Alice", "is_active": true},
			{"user_id": u2, "username": "Bob", "is_active": true},
			{"user_id": u3, "username": "Charlie", "is_active": true},
			{"user_id": u4, "username": "Sasha", "is_active": true},
		},
	}

	resp, body := postJSON(t, "/team/add", teamReq)

	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create team. Body: %s", body) {
		t.FailNow()
	}

	var teamResp map[string]any

	err := json.Unmarshal([]byte(body), &teamResp)
	require.NoError(t, err)
	require.NotNil(t, teamResp["team"], "Team response is nil")

	// 2. Create PR
	t.Log("Step 2: Creating PR...")

	prReq := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "E2E Test PR",
		"author_id":         u1,
	}

	resp, body = postJSON(t, "/pullRequest/create", prReq)

	if !assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create PR. Body: %s", body) {
		t.FailNow()
	}

	var prResp map[string]any

	err = json.Unmarshal([]byte(body), &prResp)
	require.NoError(t, err)
	require.NotNil(t, prResp["pr"], "PR response is nil")

	pr := prResp["pr"].(map[string]any)
	reviewers := pr["assigned_reviewers"].([]any)

	// Should have 2 reviewers (Bob and Charlie, not Alice who is the author)
	assert.Equal(t, 2, len(reviewers))
	assert.NotContains(t, reviewers, u1)

	t.Logf("Assigned reviewers: %v", reviewers)

	// 3. Get user reviews
	t.Log("Step 3: Getting user reviews...")

	resp, body = get(t, fmt.Sprintf("/users/getReview?user_id=%s", u2))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get reviews. Body: %s", body)

	var reviewsResp map[string]any

	err = json.Unmarshal([]byte(body), &reviewsResp)
	require.NoError(t, err)

	prs := reviewsResp["pull_requests"].([]any)
	assert.GreaterOrEqual(t, len(prs), 1)

	t.Logf("User %s has %d PRs", u2, len(prs))

	// 4. Deactivate user
	t.Log("Step 4: Deactivating user...")

	deactivateReq := map[string]any{
		"user_id":   u2,
		"is_active": false,
	}

	resp, body = postJSON(t, "/users/setIsActive", deactivateReq)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Failed to deactivate user. Body: %s", body)

	// 5. Merge PR
	t.Log("Step 6: Merging PR...")

	mergeReq := map[string]any{
		"pull_request_id": prID,
	}

	resp, body = postJSON(t, "/pullRequest/merge", mergeReq)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Failed to merge PR. Body: %s", body)

	var mergeResp map[string]any

	err = json.Unmarshal([]byte(body), &mergeResp)
	require.NoError(t, err)

	mergedPR := mergeResp["pr"].(map[string]any)
	assert.Equal(t, "MERGED", mergedPR["status"])

	// 6. Try to reassign on merged PR (should fail)
	t.Log("Step 7: Try to reassign on merged PR (should fail)...")

	reassignReq := map[string]any{
		"pull_request_id": prID,
		"old_user_id":     u3,
	}

	resp, body = postJSON(t, "/pullRequest/reassign", reassignReq)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should not allow reassign on merged PR. Body: %s", body)

	// 7. Merge again (idempotent)
	t.Log("Step 8: Merge again (idempotent)...")

	resp, body = postJSON(t, "/pullRequest/merge", mergeReq)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Merge should be idempotent. Body: %s", body)

	// 8. Get team
	t.Log("Step 9: Getting team...")

	resp, body = get(t, fmt.Sprintf("/team/get?team_name=%s", teamName))
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get team. Body: %s", body)

	// 9. Check health
	t.Log("Step 10: Checking health...")

	resp, body = get(t, "/health")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check failed. Body: %s", body)

	t.Log("Complete workflow test passed!")
}

func TestE2E_ErrorCases(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	waitForService(t)

	timestamp := time.Now().Unix()

	// Team already exists
	t.Log("Test: Team already exists...")

	teamName := fmt.Sprintf("duplicate-team-%d", timestamp)
	teamReq := map[string]any{
		"team_name": teamName,
		"members":   []map[string]any{},
	}

	postJSON(t, "/team/add", teamReq)

	resp, body := postJSON(t, "/team/add", teamReq)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for duplicate team. Body: %s", body)

	// PR already exists
	t.Log("Test: PR already exists...")

	errTeamName := fmt.Sprintf("err-team-%d", timestamp)
	u1 := fmt.Sprintf("err-u1-%d", timestamp)
	prID := fmt.Sprintf("err-pr-1-%d", timestamp)

	setup := map[string]any{
		"team_name": errTeamName,
		"members": []map[string]any{
			{"user_id": u1, "username": "Alice", "is_active": true},
		},
	}

	postJSON(t, "/team/add", setup)

	prReq := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "Duplicate PR",
		"author_id":         u1,
	}

	postJSON(t, "/pullRequest/create", prReq)

	resp, body = postJSON(t, "/pullRequest/create", prReq)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Should return 409 for duplicate PR. Body: %s", body)

	// User not found
	t.Log("Test: User not found...")

	resp, body = get(t, "/users/getReview?user_id=nonexistent")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent user. Body: %s", body)

	// Team not found
	t.Log("Test: Team not found...")

	resp, body = get(t, "/team/get?team_name=nonexistent")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent team. Body: %s", body)

	t.Log("Error cases test passed!")
}

func waitForService(t *testing.T) {

	t.Log("Waiting for service to be ready...")
	maxRetries := 30

	for i := range maxRetries {
		resp, err := http.Get(baseURL + "/health")

		if err == nil && resp.StatusCode == http.StatusOK {
			t.Log("Service is ready!")
			return
		}

		if i < maxRetries-1 {
			time.Sleep(time.Second)
		}
	}
	require.Fail(t, "Service did not become ready in time")
}

func postJSON(t *testing.T, path string, body any) (*http.Response, string) {
	data, err := json.Marshal(body)
	require.NoError(t, err)

	resp, err := http.Post(baseURL+path, "application/json", bytes.NewBuffer(data))
	require.NoError(t, err)

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)

	return resp, string(responseBody)
}

func get(t *testing.T, path string) (*http.Response, string) {
	resp, err := http.Get(baseURL + path)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)

	return resp, string(body)
}
