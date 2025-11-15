package services

import (
	"context"

	"pr-reviewer-assignment/internal/core/domain/entities"
	repo "pr-reviewer-assignment/internal/core/ports/repositories"
	"pr-reviewer-assignment/internal/validation"

	"go.uber.org/zap"
)

type UserService struct {
	userRepo repo.UserRepository
	prRepo   repo.PullRequestRepository
	logger   *zap.Logger
}

func NewUserService(userRepo repo.UserRepository, prRepo repo.PullRequestRepository, logger *zap.Logger) *UserService {

	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
		logger:   logger,
	}
}

func (s *UserService) SetActivity(ctx context.Context, userID string, isActive bool) (*entities.User, error) {
	validatedID, err := validation.RequireString("user_id", userID)
	if err != nil {
		s.logger.Error("Invalid user id", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}
	userID = validatedID

	user, err := s.userRepo.SetActivity(ctx, userID, isActive)
	if err != nil {
		s.logger.Error("Failed to set user activity", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetReviewerAssignments(ctx context.Context, userID string) ([]*entities.PullRequest, error) {
	validatedID, err := validation.RequireString("user_id", userID)
	if err != nil {
		s.logger.Error("Invalid user id", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}
	userID = validatedID

	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		s.logger.Error("Failed to load user", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	prs, err := s.prRepo.ListByReviewer(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to list reviewer assignments", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	return prs, nil
}
