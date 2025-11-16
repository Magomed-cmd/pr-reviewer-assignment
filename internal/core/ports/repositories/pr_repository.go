package repositories

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
)

type PullRequestRepository interface {
	Create(ctx context.Context, pr *entities.PullRequest) error
	Update(ctx context.Context, pr *entities.PullRequest) error
	GetByID(ctx context.Context, prID string) (*entities.PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID string) ([]*entities.PullRequest, error)
	Count(ctx context.Context) (int, error)
	CountAssignments(ctx context.Context) (int, error)
}
