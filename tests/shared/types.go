package shared

import "pr-reviewer-assignment/internal/dto"

// DTO-обёртки для декодирования JSON в тестах.
type CreateTeamResponse struct {
	Team *dto.TeamDTO `json:"team"`
}

type PullRequestResponse struct {
	PR *dto.PullRequestDTO `json:"pr"`
}

type ErrorResponse struct {
	Error dto.ErrorBody `json:"error"`
}

type UserResponse struct {
	User *dto.UserDTO `json:"user"`
}

type ReassignResponse struct {
	PR         *dto.PullRequestDTO `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}
