package types

import "strings"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

func (s PRStatus) String() string {
	return string(s)
}

func (s PRStatus) IsValid() bool {
	return s == PRStatusOpen || s == PRStatusMerged
}

func ParsePRStatus(value string) (PRStatus, bool) {
	switch PRStatus(strings.ToUpper(value)) {
	case PRStatusOpen:
		return PRStatusOpen, true
	case PRStatusMerged:
		return PRStatusMerged, true
	default:
		return "", false
	}
}

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}

func ParseRole(value string) (Role, bool) {
	switch Role(strings.ToLower(value)) {
	case RoleAdmin:
		return RoleAdmin, true
	case RoleUser:
		return RoleUser, true
	default:
		return "", false
	}
}
