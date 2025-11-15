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



type teamRepository struct {
    db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) repository.TeamRepository {
    return &teamRepository{db: db}
}


func(r *teamRepository) Create(ctx context.Context, team *models.Team) error {
    
	query := `
	    INSERT INTO teams (team_name, created_at)
        VALUES ($1, $2)
	`

    _, err := r.db.Exec(ctx, query, team.TeamName, team.CreatedAt)
    
	if err != nil {
        return err
    }

    return nil
}

func(r *teamRepository) GetByName(ctx context.Context, teamName string) (*models.Team, error) {
    
	team := models.Team{}

	queryGetTeam := `
		SELECT team_name, created_at FROM teams WHERE team_name = $1
	`
    
    err := r.db.QueryRow(ctx, queryGetTeam, teamName).Scan(&team.TeamName, &team.CreatedAt)
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperrors.ErrTeamNotFound
        }
        return nil, err
    }

	queryGetMembers := `
        SELECT user_id, username, is_active FROM users
        WHERE team_name = $1
        ORDER BY username
	`

    rows, err := r.db.Query(ctx, queryGetMembers, teamName)

	if err != nil {
        return nil, err
    }

    defer rows.Close()

    team.Members = []models.TeamMember{}

    for rows.Next() {
        var member models.TeamMember
        
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
            return nil, err
        }

        team.Members = append(team.Members, member)
    }

    return &team, nil
}

func(r *teamRepository) Exists(ctx context.Context, teamName string) (bool, error) {

	exists := false

	query := `
		SELECT EXISTS(
			SELECT 1 FROM teams WHERE team_name = $1
		)
	`

    err := r.db.QueryRow(ctx, query, teamName).Scan(&exists)

    return exists, err
}