package services

import (
	"context"
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
		txManager = transactions.NoopManager{}
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
		s.logger.Error("Invalid team name", zap.String("team_name", name), zap.Error(err))
		return nil, err
	}
	name = validatedName

	now := time.Now().UTC()
	team := entities.NewTeam(name, now, now)

	preparedMembers := s.prepareMembers(name, members, now)

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.teamRepo.Create(txCtx, team); err != nil {
			s.logger.Error("Failed to create team", zap.String("team_name", name), zap.Error(err))
			return err
		}

		if len(preparedMembers) > 0 {
			if err := s.userRepo.UpsertMany(txCtx, preparedMembers); err != nil {
				s.logger.Error("Failed to upsert team members", zap.String("team_name", name), zap.Error(err))
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if len(preparedMembers) > 0 {
		team.Members = make(map[string]*entities.User, len(preparedMembers))
		for _, member := range preparedMembers {
			team.Members[member.ID] = member
		}
	}

	return team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, name string) (*entities.Team, error) {
	validatedName, err := validation.RequireString("team_name", name)
	if err != nil {
		s.logger.Error("Invalid team name", zap.String("team_name", name), zap.Error(err))
		return nil, err
	}
	name = validatedName

	team, err := s.teamRepo.Get(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get team", zap.String("team_name", name), zap.Error(err))
		return nil, err
	}

	return team, nil
}

func (s *TeamService) prepareMembers(teamName string, members []*entities.User, fallbackTime time.Time) []*entities.User {
	if len(members) == 0 {
		return nil
	}

	sanitized := make([]*entities.User, 0, len(members))
	seen := make(map[string]struct{})

	for _, member := range members {
		if member == nil {
			continue
		}

		memberID, memberErr := validation.RequireString("user_id", member.ID)
		if memberErr != nil {
			continue
		}

		memberUsername, usernameErr := validation.RequireString("username", member.Username)
		if usernameErr != nil {
			continue
		}

		if _, exists := seen[memberID]; exists {
			continue
		}
		seen[memberID] = struct{}{}

		member.ID = memberID
		member.TeamName = teamName
		member.Username = memberUsername

		if member.CreatedAt.IsZero() {
			member.CreatedAt = fallbackTime
		}

		if member.UpdatedAt.IsZero() {
			member.UpdatedAt = member.CreatedAt
		}

		sanitized = append(sanitized, member)
	}

	return sanitized
}
