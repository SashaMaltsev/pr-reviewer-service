package service

import (
	"context"
	"math/rand"
	"sort"
	"time"

	apperrors "github.com/SashaMalcev/pr-reviewer-service/internal/errors"
	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	repository "github.com/SashaMalcev/pr-reviewer-service/internal/repository/interfaces"
)

/*
PR Service - business logic for managing pull requests

Key features of reviewer assignment:

1. Automatic reviewer selection when creating PR:
   - PR author is excluded from candidate list
   - Only active users from the same team are selected
   - Up to 2 reviewers are assigned

2. Load balancing:
   - Current number of OPEN PRs per reviewer is considered
   - Candidates are sorted by ascending load
   - Random selection is used for equal load

3. Reviewer reassignment:
   - Current reviewers and author are excluded during replacement
   - Candidate with minimum load is selected from available ones
   - Reassignment is prohibited for merged PRs

The algorithm ensures even distribution of PRs among team reviewers.
*/

type PRService struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	rand     *rand.Rand
}

func NewPRService(prRepo repository.PRRepository, userRepo repository.UserRepository, teamRepo repository.TeamRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error) {
	// Check if PR exists
	exists, err := s.prRepo.Exists(ctx, prID)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, apperrors.ErrPRExists
	}

	// Get author
	author, err := s.userRepo.GetByID(ctx, authorID)

	if err != nil {
		return nil, err
	}

	// Create PR
	pr := models.NewPullRequest(prID, prName, authorID)

	// Assign reviewers
	reviewers, err := s.selectReviewers(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, err
	}

	for _, reviewer := range reviewers {
		pr.AddReviewer(reviewer.UserID)
	}

	// Save PR
	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {

	pr, err := s.prRepo.GetByID(ctx, prID)

	if err != nil {
		return nil, err
	}

	// if already merged, return as is
	if pr.IsMerged() {
		return pr, nil
	}

	pr.Merge()

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*models.PullRequest, string, error) {

	// Get pr
	pr, err := s.prRepo.GetByID(ctx, prID)

	if err != nil {
		return nil, "", err
	}

	// Check if pr is merged
	if pr.IsMerged() {
		return nil, "", apperrors.ErrPRMerged
	}

	// Check if old reviewer is assigned
	if !pr.HasReviewer(oldUserID) {
		return nil, "", apperrors.ErrNotAssigned
	}

	// Get old reviewer's team
	oldReviewer, err := s.userRepo.GetByID(ctx, oldUserID)

	if err != nil {
		return nil, "", err
	}

	// Get candidates from the same team (excluding author and current reviewers)
	excludeIDs := append(pr.AssignedReviewers, pr.AuthorID)
	candidates, err := s.getCandidatesExcluding(ctx, oldReviewer.TeamName, excludeIDs)

	if err != nil {
		return nil, "", err
	}

	if len(candidates) == 0 {
		return nil, "", apperrors.ErrNoCandidate
	}

	// Select new reviewer with load balancing
	newReviewer := s.selectBestCandidate(ctx, candidates)

	// Replace reviewer
	pr.RemoveReviewer(oldUserID)
	pr.AddReviewer(newReviewer.UserID)

	// Update pr
	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, "", err
	}

	return pr, newReviewer.UserID, nil
}

// selects up to 2 reviewers from team
// using load balancing
func (s *PRService) selectReviewers(ctx context.Context, teamName, excludeUserID string) ([]*models.User, error) {

	candidates, err := s.userRepo.GetActiveByTeam(ctx, teamName, excludeUserID)

	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return []*models.User{}, nil
	}

	// Get current load for all candidates
	userIDs := make([]string, len(candidates))

	for i, u := range candidates {
		userIDs[i] = u.UserID
	}

	load, err := s.userRepo.GetReviewerLoad(ctx, userIDs)

	if err != nil {
		return nil, err
	}

	// Sort by load (ascending) and shuffle users with same load
	sort.Slice(candidates, func(i, j int) bool {
		loadI := load[candidates[i].UserID]
		loadJ := load[candidates[j].UserID]
		if loadI == loadJ {
			return s.rand.Intn(2) == 0
		}
		return loadI < loadJ
	})

	// Select up to 2 reviewers
	count := min(len(candidates), 2)

	return candidates[:count], nil
}

// gets active users from team excluding specified IDs
func (s *PRService) getCandidatesExcluding(ctx context.Context, teamName string, excludeIDs []string) ([]*models.User, error) {

	allCandidates, err := s.userRepo.GetActiveByTeam(ctx, teamName, "")

	if err != nil {
		return nil, err
	}

	// Filter out excluded IDs
	excludeMap := make(map[string]bool)

	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	filtered := []*models.User{}

	for _, candidate := range allCandidates {
		if !excludeMap[candidate.UserID] {
			filtered = append(filtered, candidate)
		}
	}

	return filtered, nil
}

// selects the candidate with lowest load
func (s *PRService) selectBestCandidate(ctx context.Context, candidates []*models.User) *models.User {

	if len(candidates) == 0 {
		return nil
	}

	userIDs := make([]string, len(candidates))

	for i, u := range candidates {
		userIDs[i] = u.UserID
	}

	load, err := s.userRepo.GetReviewerLoad(ctx, userIDs)

	if err != nil {
		// Fallback to random selection
		return candidates[s.rand.Intn(len(candidates))]
	}

	// Find candidates with minimum load
	minLoad := load[candidates[0].UserID]
	minLoadCandidates := []*models.User{}

	for _, candidate := range candidates {
		candidateLoad := load[candidate.UserID]
		if candidateLoad < minLoad {
			minLoad = candidateLoad
			minLoadCandidates = []*models.User{candidate}
		} else if candidateLoad == minLoad {
			minLoadCandidates = append(minLoadCandidates, candidate)
		}
	}

	// Random selection among candidates with minimum load
	return minLoadCandidates[s.rand.Intn(len(minLoadCandidates))]
}
