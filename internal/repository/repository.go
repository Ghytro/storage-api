package repository

import (
	"context"
	"fmt"
	"storageapi/internal/database"
	"strings"

	"go.uber.org/zap"
)

type IRepoMixin interface {
	DBI(ctx context.Context) DBI
	RunInTransaction(ctx context.Context, fn func(ctx database.TxContext, repo IRepository) error) error
}

type repoMixin struct {
	db  DBI
	log *zap.SugaredLogger
}

func (r *repoMixin) DBI(ctx context.Context) DBI {
	if tx := database.GetTX(ctx); tx != nil {
		return tx
	}
	return r.db
}

func (r *repoMixin) RunInTransaction(
	ctx context.Context,
	fn func(ctx database.TxContext, repo IRepository) error,
) error {
	return r.db.RunInTransaction(ctx, func(ctx database.TxContext) error {
		return fn(ctx, NewRepository(r.db, r.log))
	})
}

type Repository struct {
	*repoMixin
	*StorageRepository
	*ProductRepository
	*StoredProductRepository
	*ReservationsRepository
}

func NewRepository(db DBI, log *zap.SugaredLogger) IRepository {
	mixin := &repoMixin{
		db:  db,
		log: log,
	}
	return &Repository{
		repoMixin:               mixin,
		StorageRepository:       NewStorageRepository(db, log),
		ProductRepository:       NewProductRepository(db, log),
		StoredProductRepository: NewStoredProductRepository(db, log),
		ReservationsRepository:  NewReservationsRepository(db, log),
	}
}

type IRepository interface {
	IRepoMixin
	IStorageRepository
	IProductRepository
	IStoredProductRepository
	IReservationsRepository
}

// нужен для сбора значений в аргументы insert
// т.к. в pgx аргументы нумерованные, хранит инкремент
type argBuilder struct {
	i    int
	b    strings.Builder
	args []interface{}
}

func (b *argBuilder) add(args ...interface{}) {
	placeholders := make([]string, 0, len(args))
	for _, a := range args {
		placeholders = append(placeholders, fmt.Sprintf("$%d", b.i+1))
		args = append(args, a)
		b.i++
	}
	b.b.WriteString("(" + strings.Join(placeholders, ",") + "),")
}

func (b *argBuilder) done() (string, []interface{}) {
	return b.b.String()[:b.b.Len()-1], b.args
}
