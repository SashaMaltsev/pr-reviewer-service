package models

import (
	"time"
)


type Team struct {
    TeamName  string
    Members   []TeamMember
    CreatedAt time.Time
}


type TeamMember struct {
    UserID   string
    Username string
    IsActive bool
}


func NewTeam(teamName string, members []TeamMember) *Team {
    return &Team{
        TeamName:  teamName,
        Members:   members,
        CreatedAt: time.Now(),
    }
}