package entities

import (
	"sort"
	"time"
)

type Team struct {
	Name      string
	members   map[string]*User
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTeam(name string, createdAt, updatedAt time.Time) *Team {
	now := time.Now().UTC()

	if createdAt.IsZero() {
		createdAt = now
	}

	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	return &Team{
		Name:      name,
		members:   make(map[string]*User),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (t *Team) UpsertMember(user *User) {
	if user == nil {
		return
	}

	if t.members == nil {
		t.members = make(map[string]*User)
	}

	user.TeamName = t.Name
	t.members[user.ID] = user
	t.touch(time.Now().UTC())
}

func (t *Team) RemoveMember(userID string) {
	if t.members == nil {
		return
	}

	delete(t.members, userID)
	t.touch(time.Now().UTC())
}

func (t *Team) Member(userID string) (*User, bool) {
	if t.members == nil {
		return nil, false
	}

	user, ok := t.members[userID]
	return user, ok
}

func (t *Team) Members() []*User {
	return t.collectMembers(func(_ *User) bool { return true })
}

func (t *Team) ActiveMembers() []*User {
	return t.collectMembers(func(u *User) bool { return u.IsActive })
}

func (t *Team) ActiveMembersExcluding(userID string) []*User {
	return t.collectMembers(func(u *User) bool {
		return u.IsActive && u.ID != userID
	})
}

func (t *Team) MemberCount() int {
	return len(t.members)
}

func (t *Team) touch(updatedAt time.Time) {
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	t.UpdatedAt = updatedAt
}

func (t *Team) collectMembers(predicate func(*User) bool) []*User {
	if len(t.members) == 0 {
		return nil
	}

	result := make([]*User, 0, len(t.members))

	for _, member := range t.members {
		if predicate(member) {
			result = append(result, member)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Username == result[j].Username {
			return result[i].ID < result[j].ID
		}

		return result[i].Username < result[j].Username
	})

	return result
}
