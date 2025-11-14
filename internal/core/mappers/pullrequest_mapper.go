package mappers

import (
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	"pr-reviewer-assignment/internal/dto"
)

func PullRequestToDTO(pr *entities.PullRequest) *dto.PullRequestDTO {
	if pr == nil {
		return nil
	}

	var createdAt string
	if !pr.CreatedAt.IsZero() {
		createdAt = pr.CreatedAt.UTC().Format(time.RFC3339)
	}

	var mergedAt *string
	if pr.MergedAt != nil {
		formatted := pr.MergedAt.UTC().Format(time.RFC3339)
		mergedAt = &formatted
	}

	return &dto.PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status.String(),
		AssignedReviewers: append([]string(nil), pr.AssignedReviewers...),
		NeedMoreReviewers: pr.NeedMoreReviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}

func PullRequestsToShortDTO(prs []*entities.PullRequest) []dto.PullRequestShortDTO {
	if len(prs) == 0 {
		return nil
	}

	result := make([]dto.PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		if pr == nil {
			continue
		}

		result = append(result, dto.PullRequestShortDTO{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status.String(),
		})
	}

	return result
}
