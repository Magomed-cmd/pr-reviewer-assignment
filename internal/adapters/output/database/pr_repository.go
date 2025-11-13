package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	domainErrors "pr-reviewer-assignment/internal/core/domain/errors"
	"pr-reviewer-assignment/internal/core/domain/types"
	"pr-reviewer-assignment/internal/infrastructure/database/postgres"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type PullRequestRepository struct {
	db     postgres.DB
	logger *zap.Logger
}

func NewPullRequestRepository(db postgres.DB, logger *zap.Logger) *PullRequestRepository {
	return &PullRequestRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PullRequestRepository) Create(ctx context.Context, tx postgres.DB, pr *entities.PullRequest) error {
	const query = `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	r.logger.Debug("Creating pull request",
		zap.String("pr_id", pr.ID),
		zap.String("author_id", pr.AuthorID))

	var mergedAt any
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	}

	_, err := tx.Exec(ctx, query,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status.String(),
		pr.CreatedAt,
		mergedAt,
	)
	if err != nil {
		if isPgError(err, pgCodeUniqueViolation) {
			r.logger.Warn("Pull request already exists",
				zap.String("pr_id", pr.ID))
			return domainErrors.PRExists(pr.ID)
		}

		r.logger.Error("Failed to create pull request",
			zap.String("pr_id", pr.ID),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *PullRequestRepository) AddReviewers(ctx context.Context, tx postgres.DB, prID string, reviewers []string) error {
	if len(reviewers) == 0 {
		return nil
	}

	r.logger.Debug("Adding reviewers",
		zap.String("pr_id", prID),
		zap.Strings("reviewers", reviewers))

	const query = `
		INSERT INTO pr_reviewers (pull_request_id, user_id)
		VALUES ($1, $2)
	`

	for _, reviewer := range reviewers {
		if reviewer == "" {
			continue
		}

		_, err := tx.Exec(ctx, query, prID, reviewer)
		if err != nil {
			if isPgError(err, pgCodeForeignKeyViolation) {
				r.logger.Warn("Reviewer not found while assigning to PR",
					zap.String("pr_id", prID),
					zap.String("reviewer", reviewer))
				return domainErrors.NotFound(fmt.Sprintf("user %s", reviewer))
			}

			if isPgError(err, pgCodeUniqueViolation) {
				continue
			}

			r.logger.Error("Failed to add reviewer",
				zap.String("pr_id", prID),
				zap.String("reviewer", reviewer),
				zap.Error(err))
			return err
		}
	}

	return nil
}

func (r *PullRequestRepository) Update(ctx context.Context, pr *entities.PullRequest) error {
	const query = `
		UPDATE pull_requests
		SET pull_request_name = $2,
		    status = $3,
		    merged_at = $4
		WHERE pull_request_id = $1
	`

	r.logger.Debug("Updating pull request", zap.String("pr_id", pr.ID))

	var mergedAt any
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	}

	tag, err := r.db.Exec(ctx, query,
		pr.ID,
		pr.Name,
		pr.Status.String(),
		mergedAt,
	)
	if err != nil {
		r.logger.Error("Failed to update pull request",
			zap.String("pr_id", pr.ID),
			zap.Error(err))
		return err
	}

	if tag.RowsAffected() == 0 {
		r.logger.Warn("Pull request not found while updating",
			zap.String("pr_id", pr.ID))
		return domainErrors.NotFound(fmt.Sprintf("pull request %s", pr.ID))
	}

	return nil
}

func (r *PullRequestRepository) UpdateReviewers(ctx context.Context, tx postgres.DB, prID string, reviewers []string) error {
	r.logger.Debug("Updating reviewers",
		zap.String("pr_id", prID),
		zap.Strings("reviewers", reviewers))

	_, err := tx.Exec(ctx, `DELETE FROM pr_reviewers WHERE pull_request_id = $1`, prID)
	if err != nil {
		r.logger.Error("Failed to delete old reviewers",
			zap.String("pr_id", prID),
			zap.Error(err))
		return err
	}

	return r.AddReviewers(ctx, tx, prID, reviewers)
}

func (r *PullRequestRepository) GetByID(ctx context.Context, prID string) (*entities.PullRequest, error) {
	const query = `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	pr, err := scanPullRequest(r.db.QueryRow(ctx, query, prID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("Pull request not found",
				zap.String("pr_id", prID))
			return nil, domainErrors.NotFound(fmt.Sprintf("pull request %s", prID))
		}

		r.logger.Error("Failed to get pull request",
			zap.String("pr_id", prID),
			zap.Error(err))
		return nil, err
	}

	reviewers, err := r.fetchReviewers(ctx, pr.ID)
	if err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return pr, nil
}

func (r *PullRequestRepository) ListByReviewer(ctx context.Context, reviewerID string) ([]*entities.PullRequest, error) {
	const query = `
		SELECT DISTINCT 
			pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at,
			rev2.user_id
		FROM pull_requests pr
		JOIN pr_reviewers rev ON rev.pull_request_id = pr.pull_request_id
		LEFT JOIN pr_reviewers rev2 ON rev2.pull_request_id = pr.pull_request_id
		WHERE rev.user_id = $1
		ORDER BY pr.created_at DESC, rev2.user_id ASC
	`

	rows, err := r.db.Query(ctx, query, reviewerID)
	if err != nil {
		r.logger.Error("Failed to list PRs by reviewer",
			zap.String("reviewer_id", reviewerID),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	prMap := make(map[string]*entities.PullRequest)
	var prOrder []string

	for rows.Next() {
		var (
			id, name, authorID, statusStr string
			createdAt                     time.Time
			mergedAt                      sql.NullTime
			reviewerUserID                sql.NullString
		)

		if err := rows.Scan(&id, &name, &authorID, &statusStr, &createdAt, &mergedAt, &reviewerUserID); err != nil {
			r.logger.Error("Failed to scan pull request row",
				zap.String("reviewer_id", reviewerID),
				zap.Error(err))
			return nil, err
		}

		if _, exists := prMap[id]; !exists {
			status, ok := types.ParsePRStatus(statusStr)
			if !ok {
				r.logger.Error("Invalid pull request status while listing by reviewer",
					zap.String("status", statusStr))
				return nil, fmt.Errorf("invalid pull request status: %s", statusStr)
			}

			var mergedPtr *time.Time
			if mergedAt.Valid {
				t := mergedAt.Time.UTC()
				mergedPtr = &t
			}

			prMap[id] = &entities.PullRequest{
				ID:                id,
				Name:              name,
				AuthorID:          authorID,
				Status:            status,
				AssignedReviewers: make([]string, 0, 2),
				CreatedAt:         createdAt,
				MergedAt:          mergedPtr,
			}
			prOrder = append(prOrder, id)
		}

		if reviewerUserID.Valid {
			prMap[id].AssignedReviewers = append(prMap[id].AssignedReviewers, reviewerUserID.String)
		}
	}

	result := make([]*entities.PullRequest, 0, len(prOrder))
	for _, id := range prOrder {
		result = append(result, prMap[id])
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Rows iteration failed while listing PRs by reviewer",
			zap.String("reviewer_id", reviewerID),
			zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (r *PullRequestRepository) fetchReviewers(ctx context.Context, prID string) ([]string, error) {
	const query = `
		SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at ASC, user_id ASC
	`

	rows, err := r.db.Query(ctx, query, prID)
	if err != nil {
		r.logger.Error("Failed to fetch reviewers",
			zap.String("pr_id", prID),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0, 2)

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			r.logger.Error("Failed to scan reviewer row",
				zap.String("pr_id", prID),
				zap.Error(err))
			return nil, err
		}

		reviewers = append(reviewers, userID)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Reviewer rows iteration failed",
			zap.String("pr_id", prID),
			zap.Error(err))
		return nil, err
	}

	return reviewers, nil
}

func scanPullRequest(row rowScanner) (*entities.PullRequest, error) {
	var (
		id        string
		name      string
		authorID  string
		statusStr string
		createdAt time.Time
		mergedAt  sql.NullTime
	)

	if err := row.Scan(&id, &name, &authorID, &statusStr, &createdAt, &mergedAt); err != nil {
		return nil, err
	}

	status, ok := types.ParsePRStatus(statusStr)
	if !ok {
		return nil, fmt.Errorf("invalid pull request status: %s", statusStr)
	}

	var mergedPtr *time.Time
	if mergedAt.Valid {
		t := mergedAt.Time.UTC()
		mergedPtr = &t
	}

	return &entities.PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            status,
		AssignedReviewers: make([]string, 0, 2),
		CreatedAt:         createdAt,
		MergedAt:          mergedPtr,
	}, nil
}
