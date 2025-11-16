package models

import (
	"time"
)

type Team struct {
	TeamName  string       `json:"team_name"`
	Members   []TeamMember `json:"members"`
	CreatedAt time.Time    `json:"created_at"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func NewTeam(teamName string, members []TeamMember) *Team {
	return &Team{
		TeamName:  teamName,
		Members:   members,
		CreatedAt: time.Now(),
	}
}
