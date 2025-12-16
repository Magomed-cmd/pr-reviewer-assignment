package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	domainErrors "pr-reviewer-assignment/internal/core/domain/errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type UserRepository struct {
	db     DB
	logger *zap.Logger
}

func NewUserRepository(db DB, logger *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM users`
	var count int
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		r.logger.Error("Failed to count users", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) UpsertMany(ctx context.Context, users []*entities.User) error {
	if len(users) == 0 {
		return nil
	}

	const query = `
		INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE
		SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at
	`

	db := r.dbFor(ctx)

	for _, user := range users {
		if user == nil {
			continue
		}

		createdAt := user.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}

		updatedAt := user.UpdatedAt
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}

		if _, err := db.Exec(ctx, query,
			user.ID,
			user.Username,
			user.TeamName,
			user.IsActive,
			createdAt,
			updatedAt,
		); err != nil {
			if isPgError(err, pgCodeForeignKeyViolation) {
				r.logger.Warn("Team not found while upserting user",
					zap.String("user_id", user.ID),
					zap.String("team_name", user.TeamName))
				return domainErrors.NotFound(fmt.Sprintf("team %s", user.TeamName))
			}

			r.logger.Error("Failed to upsert user",
				zap.String("user_id", user.ID),
				zap.Error(err))
			return err
		}
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entities.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	db := r.dbFor(ctx)

	user, err := scanUser(db.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("User not found",
				zap.String("user_id", userID))
			return nil, domainErrors.NotFound(fmt.Sprintf("user %s", userID))
		}

		r.logger.Error("Failed to get user",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) ListByTeam(ctx context.Context, teamName string) ([]*entities.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1
		ORDER BY username ASC, user_id ASC
	`

	db := r.dbFor(ctx)

	rows, err := db.Query(ctx, query, teamName)
	if err != nil {
		r.logger.Error("Failed to list users by team",
			zap.String("team_name", teamName),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var users []*entities.User

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			r.logger.Error("Failed to scan user row",
				zap.String("team_name", teamName),
				zap.Error(err))
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Rows iteration failed while listing team users",
			zap.String("team_name", teamName),
			zap.Error(err))
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) SetActivity(ctx context.Context, userID string, isActive bool) (*entities.User, error) {
	const query = `
		UPDATE users
		SET is_active = $2,
		    updated_at = $3
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active, created_at, updated_at
	`

	updatedAt := time.Now().UTC()

	db := r.dbFor(ctx)

	user, err := scanUser(db.QueryRow(ctx, query, userID, isActive, updatedAt))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Warn("User not found while updating activity",
				zap.String("user_id", userID))
			return nil, domainErrors.NotFound(fmt.Sprintf("user %s", userID))
		}

		r.logger.Error("Failed to set user activity",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	return user, nil
}

func scanUser(row rowScanner) (*entities.User, error) {
	var (
		id        string
		username  string
		teamName  string
		isActive  bool
		createdAt time.Time
		updatedAt time.Time
	)

	if err := row.Scan(&id, &username, &teamName, &isActive, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	return entities.NewUser(id, username, teamName, isActive, createdAt, updatedAt), nil
}

func (r *UserRepository) dbFor(ctx context.Context) DB {
	if tx := DBFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
