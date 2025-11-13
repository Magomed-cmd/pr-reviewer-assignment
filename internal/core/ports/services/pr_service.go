package services

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
)

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, pr *entities.PullRequest) (*entities.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*entities.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*entities.PullRequest, string, error)
}
