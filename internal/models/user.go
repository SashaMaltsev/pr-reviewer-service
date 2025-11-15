package models

import (
	"time"
)


type User struct {
    UserID    string
    Username  string
    TeamName  string
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
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

func(u *User) SetActive(isActive bool) {
    u.IsActive = isActive
    u.UpdatedAt = time.Now()
}