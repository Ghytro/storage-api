package repository

import (
	"context"
	"storageapi/internal/database"

	"github.com/jackc/pgx"
)

type DBI interface {
	Exec(query string, args ...interface{}) (pgx.CommandTag, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (pgx.CommandTag, error)

	Query(query string, args ...interface{}) (*pgx.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*pgx.Rows, error)

	QueryRow(query string, args ...interface{}) *pgx.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *pgx.Row

	RunInTransaction(ctx context.Context, fn func(ctx database.TxContext) error) error
}
