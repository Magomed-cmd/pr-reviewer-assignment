package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pr-reviewer-assignment/internal/core/domain/entities"
	domainErrors "pr-reviewer-assignment/internal/core/domain/errors"

	"go.uber.org/zap"
)

type TeamRepository struct {
	db     DB
	logger *zap.Logger
}

func NewTeamRepository(db DB, logger *zap.Logger) *TeamRepository {
	return &TeamRepository{
		db:     db,
		logger: logger,
	}
}

func (r *TeamRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM teams`
	var count int
	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		r.logger.Error("Failed to count teams", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func (r *TeamRepository) Create(ctx context.Context, team *entities.Team) error {
	const query = `
		INSERT INTO teams (team_name, created_at, updated_at)
		VALUES ($1, $2, $3)
	`

	r.logger.Debug("Creating team", zap.String("team_name", team.Name))

	db := r.dbFor(ctx)

	_, err := db.Exec(ctx, query, team.Name, team.CreatedAt, team.UpdatedAt)
	if err != nil {
		if isPgError(err, pgCodeUniqueViolation) {
			r.logger.Warn("Team already exists",
				zap.String("team_name", team.Name))
			return domainErrors.TeamExists(team.Name)
		}

		r.logger.Error("Failed to create team",
			zap.String("team_name", team.Name),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *TeamRepository) Update(ctx context.Context, team *entities.Team) error {
	const query = `
		UPDATE teams
		SET updated_at = $2
		WHERE team_name = $1
	`

	r.logger.Debug("Updating team", zap.String("team_name", team.Name))

	db := r.dbFor(ctx)

	tag, err := db.Exec(ctx, query, team.Name, team.UpdatedAt)
	if err != nil {
		r.logger.Error("Failed to update team",
			zap.String("team_name", team.Name),
			zap.Error(err))
		return err
	}

	if tag.RowsAffected() == 0 {
		r.logger.Warn("Team not found while updating",
			zap.String("team_name", team.Name))
		return domainErrors.NotFound(fmt.Sprintf("team %s", team.Name))
	}

	return nil
}

func (r *TeamRepository) Get(ctx context.Context, teamName string) (*entities.Team, error) {
	const query = `
		SELECT 
			t.team_name, t.created_at, t.updated_at,
			u.user_id, u.username, u.is_active, u.created_at, u.updated_at
		FROM teams t
		LEFT JOIN users u ON u.team_name = t.team_name
		WHERE t.team_name = $1
		ORDER BY u.username
	`

	db := r.dbFor(ctx)

	rows, err := db.Query(ctx, query, teamName)
	if err != nil {
		r.logger.Error("Failed to get team",
			zap.String("team_name", teamName),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var team *entities.Team

	for rows.Next() {
		var (
			tName              string
			tCreated           time.Time
			tUpdated           time.Time
			userID, username   sql.NullString
			isActive           sql.NullBool
			uCreated, uUpdated sql.NullTime
		)

		err := rows.Scan(&tName, &tCreated, &tUpdated, &userID, &username, &isActive, &uCreated, &uUpdated)
		if err != nil {
			r.logger.Error("Failed to scan team row",
				zap.String("team_name", teamName),
				zap.Error(err))
			return nil, err
		}

		if team == nil {
			team = entities.NewTeam(tName, tCreated, tUpdated)
		}

		if userID.Valid {
			user := entities.NewUser(userID.String, username.String, teamName, isActive.Bool, uCreated.Time, uUpdated.Time)
			team.Members[user.ID] = user
		}
	}

	if team == nil {
		r.logger.Warn("Team not found",
			zap.String("team_name", teamName))
		return nil, domainErrors.NotFound(fmt.Sprintf("team %s", teamName))
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Team rows iteration failed",
			zap.String("team_name", teamName),
			zap.Error(err))
		return nil, err
	}

	return team, nil
}

func (r *TeamRepository) dbFor(ctx context.Context) DB {
	if tx := DBFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
