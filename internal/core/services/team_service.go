package services

import (
	"context"
	"strings"
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	repo "pr-reviewer-assignment/internal/core/ports/repositories"
	"pr-reviewer-assignment/internal/core/ports/transactions"
	"pr-reviewer-assignment/internal/validation"

	"go.uber.org/zap"
)

type TeamService struct {
	teamRepo  repo.TeamRepository
	userRepo  repo.UserRepository
	logger    *zap.Logger
	txManager transactions.Manager
}

func NewTeamService(teamRepo repo.TeamRepository, userRepo repo.UserRepository, logger *zap.Logger, txManager transactions.Manager) *TeamService {
	if txManager == nil {
		panic("txManager is required")
	}

	return &TeamService{
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		logger:    logger,
		txManager: txManager,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, name string, members []*entities.User) (*entities.Team, error) {
	validatedName, err := validation.RequireString("team_name", name)
	if err != nil {
		s.logger.Warn("Invalid team name", zap.String("team_name", name), zap.Error(err))
		return nil, err
	}

	now := time.Now().UTC()
	team := entities.NewTeam(validatedName, now, now)

	validMembers := s.validateMembers(members)

	addedMembers := team.AddMembers(validMembers, now)

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.teamRepo.Create(txCtx, team); err != nil {
			s.logger.Error("Failed to create team", zap.String("team_name", validatedName), zap.Error(err))
			return err
		}

		if len(addedMembers) > 0 {
			if err := s.userRepo.UpsertMany(txCtx, addedMembers); err != nil {
				s.logger.Error("Failed to upsert team members", zap.String("team_name", validatedName), zap.Error(err))
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, name string) (*entities.Team, error) {
	validatedName, err := validation.RequireString("team_name", name)
	if err != nil {
		s.logger.Warn("Invalid team name", zap.String("team_name", name), zap.Error(err))
		return nil, err
	}

	team, err := s.teamRepo.Get(ctx, validatedName)
	if err != nil {
		s.logger.Error("Failed to get team", zap.String("team_name", validatedName), zap.Error(err))
		return nil, err
	}

	return team, nil
}

func (s *TeamService) validateMembers(members []*entities.User) []*entities.User {
	if len(members) == 0 {
		return nil
	}

	valid := make([]*entities.User, 0, len(members))

	for _, member := range members {
		if member == nil {
			continue
		}

		memberID := strings.TrimSpace(member.ID)
		if memberID == "" {
			continue
		}

		memberUsername := strings.TrimSpace(member.Username)
		if memberUsername == "" {
			continue
		}

		validUser := &entities.User{
			ID:        memberID,
			Username:  memberUsername,
			IsActive:  member.IsActive,
			CreatedAt: member.CreatedAt,
			UpdatedAt: member.UpdatedAt,
		}

		valid = append(valid, validUser)
	}

	return valid
}
