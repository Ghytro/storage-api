package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx"
)

type DB struct {
	*pgx.ConnPool
}

func NewDBWithPgx(conf pgx.ConnPoolConfig) (*DB, error) {
	conn, err := pgx.NewConnPool(conf)
	if err != nil {
		return nil, err
	}
	return &DB{
		ConnPool: conn,
	}, nil
}

func (db *DB) Exec(query string, args ...interface{}) (pgx.CommandTag, error) {
	return db.ConnPool.Exec(query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (pgx.CommandTag, error) {
	return db.ConnPool.ExecEx(ctx, query, nil, args...)
}

func (db *DB) Query(query string, args ...interface{}) (*pgx.Rows, error) {
	return db.ConnPool.Query(query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*pgx.Rows, error) {
	return db.ConnPool.QueryEx(ctx, query, nil, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *pgx.Row {
	return db.ConnPool.QueryRow(query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *pgx.Row {
	return db.ConnPool.QueryRowEx(ctx, query, nil, args...)
}

func (db *DB) RunInTransaction(ctx context.Context, fn func(ctx TxContext) error) error {
	var txCtx TxContext
	if tx := GetTX(ctx); tx == nil {
		pgTx, err := db.ConnPool.Begin()
		if err != nil {
			return err
		}
		tx = &Tx{pgTx}
		txCtx = WithTX(ctx, tx)
	} else {
		var ok bool
		txCtx, ok = ctx.(TxContext)
		if !ok {
			return errors.New("runtime error: incorrect txCtx type assertion") // такого вообще не должно возникать
		}
	}
	if err := fn(txCtx); err != nil {
		return txCtx.TX().Rollback()
	}
	return txCtx.TX().Commit()
}

type Tx struct {
	*pgx.Tx
}

func (tx *Tx) Exec(query string, args ...interface{}) (pgx.CommandTag, error) {
	return tx.Tx.Exec(query)
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (pgx.CommandTag, error) {
	return tx.Tx.ExecEx(ctx, query, nil, args...)
}

func (tx *Tx) Query(query string, args ...interface{}) (*pgx.Rows, error) {
	return tx.Tx.Query(query, args...)
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*pgx.Rows, error) {
	return tx.Tx.QueryEx(ctx, query, nil, args...)
}

func (tx *Tx) QueryRow(query string, args ...interface{}) *pgx.Row {
	return tx.Tx.QueryRow(query, args...)
}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *pgx.Row {
	return tx.Tx.QueryRowEx(ctx, query, nil, args...)
}

func (tx *Tx) RunInTransaction(ctx context.Context, fn func(ctx TxContext) error) error {
	return fn(WithTX(ctx, tx))
}

type ctxValKey string

const _txCtxKey ctxValKey = "__tx"

func GetTX(ctx context.Context) *Tx {
	if txCtx, ok := ctx.(TxContext); ok {
		return txCtx.TX()
	}
	val := ctx.Value(_txCtxKey)
	if _val, ok := val.(*Tx); ok {
		return _val
	}
	return nil
}

func WithTX(ctx context.Context, tx *Tx) TxContext {
	if txCtx, ok := ctx.(TxContext); ok {
		return txCtx
	}
	return &TxCtx{context.WithValue(ctx, _txCtxKey, tx)}
}

type TxContext interface {
	context.Context
	TX() *Tx
}

type TxCtx struct {
	context.Context
}

func (tx *TxCtx) TX() *Tx {
	return GetTX(tx.Context)
}
