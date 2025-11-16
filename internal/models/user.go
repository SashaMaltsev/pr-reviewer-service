package models

import (
	"time"
)

type User struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	TeamName  string    `json:"team_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(userID, username, teamName string, isActive bool) *User {
	now := time.Now()
	return &User{
		UserID:    userID,
		Username:  username,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) SetActive(isActive bool) {
	u.IsActive = isActive
	u.UpdatedAt = time.Now()
}
