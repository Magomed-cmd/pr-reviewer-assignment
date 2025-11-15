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
