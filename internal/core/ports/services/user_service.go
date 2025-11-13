package services

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
)

type UserService interface {
	SetActivity(ctx context.Context, userID string, isActive bool) (*entities.User, error)
	GetReviewerAssignments(ctx context.Context, userID string) ([]*entities.PullRequest, error)
}
