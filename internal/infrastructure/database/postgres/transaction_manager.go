package postgres

import (
	"context"

	"pr-reviewer-assignment/internal/core/ports/transactions"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type TransactionManager struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewTransactionManager(pool *pgxpool.Pool, logger *zap.Logger) *TransactionManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &TransactionManager{
		pool:   pool,
		logger: logger,
	}
}

func (m *TransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if fn == nil {
		return nil
	}

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		m.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}

	txCtx := ContextWithDB(ctx, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			m.logger.Error("transaction rollback failed", zap.Error(rbErr))
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		m.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	return nil
}

var _ transactions.Manager = (*TransactionManager)(nil)
