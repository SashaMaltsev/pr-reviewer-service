package apperrors

import "errors"

// Basic business logic errors
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


// Error codes for API responses
type ErrorCode string


const (
    CodeTeamExists  ErrorCode = "TEAM_EXISTS"
    CodePRExists    ErrorCode = "PR_EXISTS"
    CodePRMerged    ErrorCode = "PR_MERGED"
    CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
    CodeNoCandidate ErrorCode = "NO_CANDIDATE"
    CodeNotFound    ErrorCode = "NOT_FOUND"
)


// Mapping errors to codes for HTTP responses
func GetErrorCode(err error) ErrorCode {
    switch {
    case errors.Is(err, ErrTeamExists):
        return CodeTeamExists
    case errors.Is(err, ErrPRExists):
        return CodePRExists
    case errors.Is(err, ErrPRMerged):
        return CodePRMerged
    case errors.Is(err, ErrNotAssigned):
        return CodeNotAssigned
    case errors.Is(err, ErrNoCandidate):
        return CodeNoCandidate
    case errors.Is(err, ErrTeamNotFound), errors.Is(err, ErrUserNotFound), errors.Is(err, ErrPRNotFound):
        return CodeNotFound
    default:
        return CodeNotFound
    }
}
