package database

import "context"

type dbKey struct{}

func ContextWithDB(ctx context.Context, db DB) context.Context {
	if ctx == nil || db == nil {
		return ctx
	}

	return context.WithValue(ctx, dbKey{}, db)
}

func DBFromContext(ctx context.Context) DB {
	if ctx == nil {
		return nil
	}

	if db, ok := ctx.Value(dbKey{}).(DB); ok {
		return db
	}

	return nil
}
