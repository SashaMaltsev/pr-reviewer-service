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


type userRepository struct {
    db *pgxpool.Pool
}


func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
    return &userRepository{db: db}
}


func(r *userRepository) Create(ctx context.Context, user *models.User) error {
    
    query :=  `
        INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (user_id) DO UPDATE SET
            username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active,
            updated_at = EXCLUDED.updated_at
    `
    
    _, err := r.db.Exec(ctx, query, user.UserID, user.Username, user.TeamName, user.IsActive, user.CreatedAt, user.UpdatedAt)
    return err
}


func(r *userRepository) Update(ctx context.Context, user *models.User) error {
    
    query := `
        UPDATE users SET
            username = $2,
            team_name = $3,
            is_active = $4,
            updated_at = $5
        WHERE user_id = $1
    `

    result, err := r.db.Exec(ctx, query, 
        user.UserID, user.Username, user.TeamName, 
        user.IsActive, user.UpdatedAt,
    )

    if err != nil {
        return err
    }

    if result.RowsAffected() == 0 {
        return apperrors.ErrUserNotFound
    }

    return nil
}


func(r *userRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
    
    user := models.User{}
    
    query := `
        SELECT user_id, username, team_name, is_active, created_at, updated_at
        FROM users WHERE user_id = $1
    `

    err := r.db.QueryRow(ctx, query, userID).Scan(
        &user.UserID, &user.Username, &user.TeamName,
        &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperrors.ErrUserNotFound
        }
        return nil, err
    }

    return &user, nil
}


func(r *userRepository) GetActiveByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*models.User, error) {
    
    query := `
        SELECT user_id, username, team_name, is_active, created_at, updated_at
        FROM users
        WHERE team_name = $1 AND is_active = true AND user_id != $2
        ORDER BY username
    `

    rows, err := r.db.Query(ctx, query, teamName, excludeUserID)
    
    if err != nil {
        return nil, err
    }

    defer rows.Close()

    var users []*models.User

    for rows.Next() {
        var user models.User
        
        err := rows.Scan(
            &user.UserID, &user.Username, &user.TeamName,
            &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
        )

        if err != nil {
            return nil, err
        }

        users = append(users, &user)
    }

    return users, nil
}

func(r *userRepository) GetReviewerLoad(ctx context.Context, userIDs []string) (map[string]int, error) {
    
    if len(userIDs) == 0 {
        return map[string]int{}, nil
    }

    query := `
        SELECT r.user_id, COUNT(*) FROM pr_reviewers r
        JOIN pull_requests p ON r.pull_request_id = p.pull_request_id
        WHERE r.user_id = ANY($1) AND p.status = 'OPEN'
        GROUP BY r.user_id
    `

    rows, err := r.db.Query(ctx, query, userIDs)
    
    if err != nil {
        return nil, err
    }

    defer rows.Close()

    load := make(map[string]int)

    for rows.Next() {
        var userID string
        var count int
        
        err := rows.Scan(&userID, &count); 
        
        if err != nil {
            return nil, err
        }

        load[userID] = count
    }

    for _, userID := range userIDs {
        if _, exists := load[userID]; !exists {
            load[userID] = 0
        }
    }

    return load, nil
}