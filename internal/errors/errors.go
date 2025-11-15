package apperrors

import "errors"

var (
    ErrTeamExists   = errors.New("team already exists")
    ErrTeamNotFound = errors.New("team not found")
    ErrUserNotFound = errors.New("user not found")
    ErrPRExists     = errors.New("pull request already exists")
    ErrPRNotFound   = errors.New("pull request not found")
    ErrPRMerged     = errors.New("cannot reassign on merged PR")
    ErrNotAssigned  = errors.New("reviewer is not assigned to this PR")
    ErrNoCandidate  = errors.New("no active replacement candidate in team")
)