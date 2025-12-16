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

func (t *Team) AddMember(user *User, fallbackTime time.Time) error {
	if user == nil {
		return nil
	}

	if user.ID == "" || user.Username == "" {
		return nil
	}

	if _, exists := t.Members[user.ID]; exists {
		return nil
	}


	memberCopy := &User{
		ID:        user.ID,
		Username:  user.Username,
		TeamName:  t.Name, 
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if memberCopy.CreatedAt.IsZero() {
		memberCopy.CreatedAt = fallbackTime
	}
	if memberCopy.UpdatedAt.IsZero() {
		memberCopy.UpdatedAt = memberCopy.CreatedAt
	}

	t.Members[memberCopy.ID] = memberCopy
	return nil
}

func (t *Team) AddMembers(users []*User, fallbackTime time.Time) []*User {
	if len(users) == 0 {
		return nil
	}

	added := make([]*User, 0, len(users))
	seenIDs := make(map[string]struct{})

	for _, user := range users {
		if user == nil || user.ID == "" {
			continue
		}

		if _, seen := seenIDs[user.ID]; seen {
			continue
		}

		if _, exists := t.Members[user.ID]; exists {
			seenIDs[user.ID] = struct{}{}
			continue
		}

		if err := t.AddMember(user, fallbackTime); err == nil {
			added = append(added, t.Members[user.ID])
			seenIDs[user.ID] = struct{}{}
		}
	}

	return added
}
