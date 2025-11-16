package repositories

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
)

type UserRepository interface {
	UpsertMany(ctx context.Context, users []*entities.User) error
	GetByID(ctx context.Context, userID string) (*entities.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]*entities.User, error)
	SetActivity(ctx context.Context, userID string, isActive bool) (*entities.User, error)
	Count(ctx context.Context) (int, error)
}
