package dto

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ListPullRequestsResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

type StatsDTO struct {
	Teams        int `json:"teams"`
	Users        int `json:"users"`
	PullRequests int `json:"pull_requests"`
	Assignments  int `json:"assignments"`
}

type StatsResponse struct {
	Stats StatsDTO `json:"stats"`
}
