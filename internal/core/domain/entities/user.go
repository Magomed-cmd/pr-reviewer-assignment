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
	now := time.Now().UTC()

	if createdAt.IsZero() {
		createdAt = now
	}

	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	return &User{
		ID:        id,
		Username:  username,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (u *User) Rename(username string, updatedAt time.Time) {
	u.Username = username
	u.touch(updatedAt)
}

func (u *User) MoveToTeam(teamName string, updatedAt time.Time) {
	u.TeamName = teamName
	u.touch(updatedAt)
}

func (u *User) SetActivity(active bool, updatedAt time.Time) {
	u.IsActive = active
	u.touch(updatedAt)
}

func (u *User) BelongsTo(teamName string) bool {
	return u.TeamName == teamName
}

func (u *User) touch(updatedAt time.Time) {
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	u.UpdatedAt = updatedAt
}
