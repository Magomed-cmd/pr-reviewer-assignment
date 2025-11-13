package errors

import "fmt"

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	code    ErrorCode
	message string
}

func (e DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

func (e DomainError) Code() ErrorCode {
	return e.code
}

func (e DomainError) Message() string {
	return e.message
}

func NewDomainError(code ErrorCode, message string) error {
	return DomainError{
		code:    code,
		message: message,
	}
}

func TeamExists(teamName string) error {
	return NewDomainError(ErrorCodeTeamExists, fmt.Sprintf("team %s already exists", teamName))
}

func PRExists(prID string) error {
	return NewDomainError(ErrorCodePRExists, fmt.Sprintf("pull request %s already exists", prID))
}

func PRMerged(prID string) error {
	return NewDomainError(ErrorCodePRMerged, fmt.Sprintf("pull request %s is already merged", prID))
}

func NotAssigned(userID, prID string) error {
	return NewDomainError(ErrorCodeNotAssigned, fmt.Sprintf("user %s is not assigned to pull request %s", userID, prID))
}

func NoCandidate(teamName string) error {
	return NewDomainError(ErrorCodeNoCandidate, fmt.Sprintf("no active candidates found in team %s", teamName))
}

func NotFound(resource string) error {
	return NewDomainError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource))
}
