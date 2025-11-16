package services

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
)

type StatsService interface {
	GetStats(ctx context.Context) (*entities.Stats, error)
}
