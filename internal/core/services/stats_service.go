package services

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
	repo "pr-reviewer-assignment/internal/core/ports/repositories"
)

type StatsService struct {
	teamRepo repo.TeamRepository
	userRepo repo.UserRepository
	prRepo   repo.PullRequestRepository
}

func NewStatsService(teamRepo repo.TeamRepository, userRepo repo.UserRepository, prRepo repo.PullRequestRepository) *StatsService {
	return &StatsService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *StatsService) GetStats(ctx context.Context) (*entities.Stats, error) {
	teams, err := s.teamRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	prs, err := s.prRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	assignments, err := s.prRepo.CountAssignments(ctx)
	if err != nil {
		return nil, err
	}

	return &entities.Stats{
		Teams:        teams,
		Users:        users,
		PullRequests: prs,
		Assignments:  assignments,
	}, nil
}
