package repositories

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
)

type TeamRepository interface {
	Create(ctx context.Context, team *entities.Team) error
	Update(ctx context.Context, team *entities.Team) error
	Get(ctx context.Context, teamName string) (*entities.Team, error)
	Count(ctx context.Context) (int, error)
}
