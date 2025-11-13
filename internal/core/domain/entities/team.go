package entities

import "time"

type Team struct {
	Name      string
	Members   map[string]*User
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTeam(name string, createdAt, updatedAt time.Time) *Team {
	return &Team{
		Name:      name,
		Members:   make(map[string]*User),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (t *Team) ActiveMembersExcluding(userID string) []*User {
	var result []*User

	for _, member := range t.Members {
		if member.IsActive && member.ID != userID {
			result = append(result, member)
		}
	}

	return result
}
