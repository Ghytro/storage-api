package repository

import (
	"context"
	"errors"
	"storageapi/internal/entity"
	"strings"

	"go.uber.org/zap"
)

type StorageRepository struct {
	*repoMixin
}

var _ IStorageRepository = (*StorageRepository)(nil)

func NewStorageRepository(db DBI, log *zap.SugaredLogger) *StorageRepository {
	return &StorageRepository{
		repoMixin: &repoMixin{
			db:  db,
			log: log,
		},
	}
}

func (r *StorageRepository) GetStorage(ctx context.Context, id entity.PK) (*entity.Storage, error) {
	rows, err := r.DBI(ctx).QueryContext(ctx, "SELECT * FROM storages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	var s *entity.Storage
	for rows.Next() {
		s = &entity.Storage{}
		if err := s.Scan(rows); err != nil {
			return nil, err
		}
	}
	if s == nil {
		return nil, errors.New("cannot find storage by id")
	}
	return s, nil
}

func (r *StorageRepository) ListStorages(ctx context.Context, isAvailable ...bool) ([]*entity.Storage, error) {
	q := "SELECT * FROM storages"
	if len(isAvailable) > 0 {
		q += " WHERE is_available = $1"
	}
	args := make([]interface{}, 0, len(isAvailable))
	rows, err := r.DBI(ctx).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return entity.ScannedRows[*entity.Storage](rows)
}

func (r *StorageRepository) CreateStorage(ctx context.Context, storages ...*entity.Storage) ([]*entity.Storage, error) {
	q := strings.Builder{}
	q.WriteString("INSERT INTO storages (is_available) VALUES ")
	argB := argBuilder{}
	for _, s := range storages {
		argB.add(s.IsAvailable)
	}
	expr, args := argB.done()
	q.WriteString(expr + " RETURNING *")

	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	return entity.ScannedRows[*entity.Storage](rows)
}

func (r *StorageRepository) UpdateStorage(ctx context.Context, storages ...*entity.Storage) ([]*entity.Storage, error) {
	q := strings.Builder{}
	q.WriteString(`UPDATE storages AS s SET
		id = c.id,
		is_available = c.is_available
	FROM (VALUES `)
	argB := argBuilder{}
	for _, s := range storages {
		argB.add(s.ID, s.IsAvailable)
	}
	expr, args := argB.done()
	q.WriteString(expr + ") AS c(id, is_available) WHERE c.id = s.id RETURNING *")
	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	return entity.ScannedRows[*entity.Storage](rows)
}

func (r *StorageRepository) DeleteStorage(ctx context.Context, ids ...entity.PK) error {
	_, err := r.DBI(ctx).ExecContext(ctx, "DELETE FROM storages WHERE id IN ($1)", ids)
	return err
}

func (r *StorageRepository) TruncateStorages(ctx context.Context) error {
	_, err := r.DBI(ctx).ExecContext(ctx, "TRUNCATE TABLE storages")
	return err
}

type IStorageRepository interface {
	GetStorage(ctx context.Context, id entity.PK) (*entity.Storage, error)
	ListStorages(ctx context.Context, isAvailable ...bool) ([]*entity.Storage, error)
	CreateStorage(ctx context.Context, storages ...*entity.Storage) ([]*entity.Storage, error)
	UpdateStorage(ctx context.Context, storages ...*entity.Storage) ([]*entity.Storage, error)
	DeleteStorage(ctx context.Context, ids ...entity.PK) error
	TruncateStorages(ctx context.Context) error
}
