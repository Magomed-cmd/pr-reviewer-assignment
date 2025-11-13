package services

import (
	"context"
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	domainErrors "pr-reviewer-assignment/internal/core/domain/errors"
	repo "pr-reviewer-assignment/internal/core/ports/repositories"
	"pr-reviewer-assignment/internal/validation"

	"go.uber.org/zap"
)

type PullRequestService struct {
	prRepo   repo.PullRequestRepository
	userRepo repo.UserRepository
	teamRepo repo.TeamRepository
	logger   *zap.Logger
}

func NewPullRequestService(
	prRepo repo.PullRequestRepository,
	userRepo repo.UserRepository,
	teamRepo repo.TeamRepository,
	logger *zap.Logger,
) *PullRequestService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		logger:   logger,
	}
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, pr *entities.PullRequest) (*entities.PullRequest, error) {
	if err := validation.RequireNotNil("pull_request", pr); err != nil {
		s.logger.Error("Invalid pull request payload", zap.Error(err))
		return nil, err
	}

	validatedID, err := validation.RequireString("pull_request_id", pr.ID)
	if err != nil {
		s.logger.Error("Invalid pull request id", zap.String("pull_request_id", pr.ID), zap.Error(err))
		return nil, err
	}
	pr.ID = validatedID

	validatedName, err := validation.RequireString("pull_request_name", pr.Name)
	if err != nil {
		s.logger.Error("Invalid pull request name", zap.String("pull_request_name", pr.Name), zap.Error(err))
		return nil, err
	}
	pr.Name = validatedName

	validatedAuthorID, err := validation.RequireString("author_id", pr.AuthorID)
	if err != nil {
		s.logger.Error("Invalid author id", zap.String("author_id", pr.AuthorID), zap.Error(err))
		return nil, err
	}
	pr.AuthorID = validatedAuthorID

	if pr.CreatedAt.IsZero() {
		pr.CreatedAt = time.Now().UTC()
	}

	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		s.logger.Error("Failed to load author", zap.String("author_id", pr.AuthorID), zap.Error(err))
		return nil, err
	}

	team, err := s.teamRepo.Get(ctx, author.TeamName)
	if err != nil {
		s.logger.Error("Failed to load team for author", zap.String("team_name", author.TeamName), zap.Error(err))
		return nil, err
	}

	candidateIDs := s.buildReviewerPool(team.ActiveMembersExcluding(pr.AuthorID))
	if len(candidateIDs) == 0 {
		s.logger.Warn("No reviewers available", zap.String("team_name", team.Name), zap.String("pr_id", pr.ID))
		return nil, domainErrors.NoCandidate(team.Name)
	}

	if err := pr.AssignReviewers(candidateIDs); err != nil {
		s.logger.Error("Failed to assign reviewers", zap.String("pr_id", pr.ID), zap.Error(err))
		return nil, err
	}

	if len(pr.AssignedReviewers) == 0 {
		s.logger.Warn("No reviewers assigned", zap.String("team_name", team.Name), zap.String("pr_id", pr.ID))
		return nil, domainErrors.NoCandidate(team.Name)
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		s.logger.Error("Failed to persist pull request", zap.String("pr_id", pr.ID), zap.Error(err))
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) MergePullRequest(ctx context.Context, prID string) (*entities.PullRequest, error) {
	validatedID, err := validation.RequireString("pull_request_id", prID)
	if err != nil {
		return nil, err
	}
	prID = validatedID

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.logger.Error("Failed to load pull request", zap.String("pr_id", prID), zap.Error(err))
		return nil, err
	}

	pr.Merge(time.Now().UTC())

	if err := s.prRepo.Update(ctx, pr); err != nil {
		s.logger.Error("Failed to update pull request", zap.String("pr_id", pr.ID), zap.Error(err))
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*entities.PullRequest, string, error) {
	validatedPrID, err := validation.RequireString("pull_request_id", prID)
	if err != nil {
		s.logger.Error("Invalid pull request id", zap.String("pull_request_id", prID), zap.Error(err))
		return nil, "", err
	}
	prID = validatedPrID

	validatedOldReviewerID, err := validation.RequireString("reviewer_id", oldReviewerID)
	if err != nil {
		s.logger.Error("Invalid reviewer id", zap.String("reviewer_id", oldReviewerID), zap.Error(err))
		return nil, "", err
	}
	oldReviewerID = validatedOldReviewerID

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.logger.Error("Failed to load pull request", zap.String("pr_id", prID), zap.Error(err))
		return nil, "", err
	}

	author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
	if err != nil {
		s.logger.Error("Failed to load author", zap.String("author_id", pr.AuthorID), zap.Error(err))
		return nil, "", err
	}

	team, err := s.teamRepo.Get(ctx, author.TeamName)
	if err != nil {
		s.logger.Error("Failed to load team for author", zap.String("team_name", author.TeamName), zap.Error(err))
		return nil, "", err
	}

	replacementID, err := s.pickReplacement(team, pr, oldReviewerID)
	if err != nil {
		s.logger.Error("Failed to pick replacement reviewer", zap.String("pr_id", prID), zap.Error(err))
		return nil, "", err
	}

	newReviewerID, err := pr.ReplaceReviewer(oldReviewerID, replacementID)
	if err != nil {
		s.logger.Error("Failed to replace reviewer", zap.String("pr_id", prID), zap.Error(err))
		return nil, "", err
	}

	if err := s.prRepo.Update(ctx, pr); err != nil {
		s.logger.Error("Failed to update pull request", zap.String("pr_id", prID), zap.Error(err))
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

func (s *PullRequestService) buildReviewerPool(members []*entities.User) []string {
	if len(members) == 0 {
		return nil
	}

	pool := make([]string, 0, len(members))
	seen := make(map[string]struct{}, len(members))

	for _, member := range members {
		if member == nil {
			continue
		}

		id, err := validation.RequireString("user_id", member.ID)
		if err != nil {
			continue
		}

		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}

		pool = append(pool, id)
	}

	return pool
}

func (s *PullRequestService) pickReplacement(team *entities.Team, pr *entities.PullRequest, oldReviewerID string) (string, error) {
	candidates := team.ActiveMembersExcluding(pr.AuthorID)
	if len(candidates) == 0 {
		return "", domainErrors.NoCandidate(team.Name)
	}

	excluded := make(map[string]struct{}, len(pr.AssignedReviewers)+2)
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			continue
		}
		excluded[reviewer] = struct{}{}
	}
	excluded[oldReviewerID] = struct{}{}
	excluded[pr.AuthorID] = struct{}{}

	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}

		id, err := validation.RequireString("user_id", candidate.ID)
		if err != nil {
			continue
		}

		if _, exists := excluded[id]; exists {
			continue
		}

		return id, nil
	}

	return "", domainErrors.NoCandidate(team.Name)
}
