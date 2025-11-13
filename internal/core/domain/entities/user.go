package entities

import "time"

type User struct {
	ID        string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(id, username, teamName string, isActive bool, createdAt, updatedAt time.Time) *User {
	return &User{
		ID:        id,
		Username:  username,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
