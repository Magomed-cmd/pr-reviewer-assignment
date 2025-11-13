package services

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
)

type TeamService interface {
	CreateTeam(ctx context.Context, name string, members []*entities.User) (*entities.Team, error)
	GetTeam(ctx context.Context, name string) (*entities.Team, error)
}
