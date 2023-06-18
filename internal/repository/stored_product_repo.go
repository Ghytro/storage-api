package repository

import (
	"context"
	"errors"
	"storageapi/internal/entity"
	"strings"

	"go.uber.org/zap"
)

type StoredProductRepository struct {
	*repoMixin
}

var _ IStoredProductRepository = (*StoredProductRepository)(nil)

func NewStoredProductRepository(db DBI, log *zap.SugaredLogger) *StoredProductRepository {
	return &StoredProductRepository{
		repoMixin: &repoMixin{
			db:  db,
			log: log,
		},
	}
}

func (r *StoredProductRepository) GetStorageData(ctx context.Context, id entity.PK) (*entity.StoredProduct, error) {
	rows, err := r.DBI(ctx).QueryContext(ctx, "SELECT * FROM stored_products WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	var sp *entity.StoredProduct
	for rows.Next() {
		sp = &entity.StoredProduct{}
		if err := sp.Scan(rows); err != nil {
			return nil, err
		}
	}
	if sp == nil {
		return nil, errors.New("product reservation by id not found")
	}
	return sp, nil
}

func (r *StoredProductRepository) GetStorageDataByProduct(ctx context.Context, productIDs ...entity.PK) ([]*entity.StoredProduct, error) {
	rows, err := r.DBI(ctx).QueryContext(ctx, "SELECT * FROM stored_products WHERE product_id IN ($1)", productIDs)
	if err != nil {
		return nil, err
	}
	result := []*entity.StoredProduct{}
	for rows.Next() {
		sp := &entity.StoredProduct{}
		if err := sp.Scan(rows); err != nil {
			return nil, err
		}
		result = append(result, sp)
	}
	return result, nil
}

func (r *StoredProductRepository) CreateStorageData(ctx context.Context, data ...*entity.StoredProduct) ([]*entity.StoredProduct, error) {
	q := strings.Builder{}
	q.WriteString("INSERT INTO stored_products (storage_id, product_id, amount) VALUES ")
	argB := argBuilder{}
	for _, sp := range data {
		argB.add(sp.StorageID, sp.ProductID, sp.Amount)
	}
	expr, args := argB.done()
	q.WriteString(expr + " RETURNING *")

	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	result := []*entity.StoredProduct{}
	for rows.Next() {
		sp := &entity.StoredProduct{}
		if err := sp.Scan(rows); err != nil {
			return nil, err
		}
		result = append(result, sp)
	}
	return result, nil
}

func (r *StoredProductRepository) UpdateStorageData(ctx context.Context, data ...*entity.StoredProduct) ([]*entity.StoredProduct, error) {
	q := strings.Builder{}
	q.WriteString(`UPDATE stored_products AS sp SET
		amount = c.amount
	FROM (VALUES `)
	argB := argBuilder{}
	for _, r := range data {
		argB.add(r.StorageID, r.ProductID, r.Amount)
	}
	expr, args := argB.done()
	q.WriteString(expr + ") AS c (storage_id, product_id, amount) WHERE c.storage_id = sp.storage_id AND c.product_id = sp.product_id RETURNING *")
	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	result := []*entity.StoredProduct{}
	for rows.Next() {
		p := &entity.StoredProduct{}
		if err := p.Scan(rows); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *StoredProductRepository) DeleteStorageData(ctx context.Context, ids ...entity.PK) error {
	_, err := r.DBI(ctx).ExecContext(ctx, "DELETE FROM stored_products WHERE id IN ($1)", ids)
	return err
}

type IStoredProductRepository interface {
	GetStorageData(ctx context.Context, id entity.PK) (*entity.StoredProduct, error)
	GetStorageDataByProduct(ctx context.Context, productIDs ...entity.PK) ([]*entity.StoredProduct, error)
	CreateStorageData(ctx context.Context, data ...*entity.StoredProduct) ([]*entity.StoredProduct, error)
	UpdateStorageData(ctx context.Context, data ...*entity.StoredProduct) ([]*entity.StoredProduct, error)
	DeleteStorageData(ctx context.Context, ids ...entity.PK) error
}
