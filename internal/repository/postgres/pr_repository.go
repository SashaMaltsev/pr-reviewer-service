package postgres

import (
	"context"
	"errors"

	apperrors "github.com/SashaMalcev/pr-reviewer-service/internal/errors"
	"github.com/SashaMalcev/pr-reviewer-service/internal/models"
	repository "github.com/SashaMalcev/pr-reviewer-service/internal/repository/interfaces"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


type prRepository struct {
    db *pgxpool.Pool
}


func NewPRRepository(db *pgxpool.Pool) repository.PRRepository {
    return &prRepository{db: db}
}


func(r *prRepository) Create(ctx context.Context, pr *models.PullRequest) error {
    tx, err := r.db.Begin(ctx)
    
	if err != nil {
        return err
    }

    defer tx.Rollback(ctx)

	queryInsertPR :=  `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `

    _, err = tx.Exec(ctx, queryInsertPR, pr.PullRequestID, pr.PullRequestName, 
		pr.AuthorID, pr.Status, pr.CreatedAt,
	)
    
	if err != nil {
        return err
    }

	queryInsertReviewers := `
		INSERT INTO pr_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
	`

    for _, reviewerID := range pr.AssignedReviewers {
        _, err = tx.Exec(ctx, queryInsertReviewers, pr.PullRequestID, reviewerID)
        if err != nil {
            return err
        }
    }

    return tx.Commit(ctx)
}

func(r *prRepository) Update(ctx context.Context, pr *models.PullRequest) error {
    
	tx, err := r.db.Begin(ctx)

	if err != nil {
        return err
    }

    defer tx.Rollback(ctx)

	queryUpdatePR := `
        UPDATE pull_requests SET
            pull_request_name = $2,
            status = $3,
            merged_at = $4
        WHERE pull_request_id = $1
    `

    _, err = tx.Exec(ctx, queryUpdatePR, pr.PullRequestID, pr.PullRequestName, pr.Status, pr.MergedAt)
    
	if err != nil {
        return err
    }

	queryDeleteOld := `
		DELETE FROM pr_reviewers WHERE pull_request_id = $1
	`

    _, err = tx.Exec(ctx, queryDeleteOld, pr.PullRequestID)

    if err != nil {
        return err
    }

	queryInsertNew := `
        INSERT INTO pr_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
    `

    for _, reviewerID := range pr.AssignedReviewers {
        
		_, err = tx.Exec(ctx, queryInsertNew, pr.PullRequestID, reviewerID)
        
		if err != nil {
            return err
        }
    }

    return tx.Commit(ctx)
}


func(r *prRepository) GetByID(ctx context.Context, prID string) (*models.PullRequest, error) {
    
	pr := models.PullRequest{}

	query := `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests WHERE pull_request_id = $1
	`
    
    err := r.db.QueryRow(ctx, query, prID).Scan(
        &pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID,
        &pr.Status, &pr.CreatedAt, &pr.MergedAt,
    )
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperrors.ErrPRNotFound
        }
        return nil, err
    }

	queryGetReviewers := `
	    SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY assigned_at
	`

    rows, err := r.db.Query(ctx, queryGetReviewers, prID)
    
	if err != nil {
        return nil, err
    }

    defer rows.Close()

    pr.AssignedReviewers = []string{}

    for rows.Next() {
        var reviewerID string
        if err := rows.Scan(&reviewerID); err != nil {
            return nil, err
        }
        pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
    }

    return &pr, nil
}

func (r *prRepository) Exists(ctx context.Context, prID string) (bool, error) {

    exists := false

	query := `
		SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)
	`

    err := r.db.QueryRow(ctx, query, prID).Scan(&exists)

    return exists, err
}

func (r *prRepository) GetByReviewer(ctx context.Context, userID string) ([]*models.PullRequest, error) {
    
	query := `
        SELECT DISTINCT p.pull_request_id, p.pull_request_name, p.author_id, p.status
        FROM pull_requests p
        JOIN pr_reviewers r ON p.pull_request_id = r.pull_request_id
        WHERE r.user_id = $1
        ORDER BY p.created_at DESC
    `
	
	rows, err := r.db.Query(ctx, query, userID)
    
    if err != nil {
        return nil, err
    }

    defer rows.Close()

    prs := []*models.PullRequest{}

    for rows.Next() {
        pr := models.PullRequest{}
        
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
            return nil, err
        }

        prs = append(prs, &pr)
    }

    return prs, nil
}

func (r *prRepository) GetAssignmentStats(ctx context.Context) (map[string]int, error) {
    
	query :=  `
        SELECT user_id, COUNT(*)
        FROM pr_reviewers
        GROUP BY user_id
    `
	
	rows, err := r.db.Query(ctx, query)
    
	if err != nil {
        return nil, err
    }

    defer rows.Close()

    stats := make(map[string]int)
    
	for rows.Next() {
        var userID string
        var count int
        
		if err := rows.Scan(&userID, &count); err != nil {
            return nil, err
        }

        stats[userID] = count
    }

    return stats, nil
}