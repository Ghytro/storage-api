package repository

import (
	"context"
	"errors"
	"storageapi/internal/entity"
	"strings"

	"go.uber.org/zap"
)

type ReservationsRepository struct {
	*repoMixin
}

var _ IReservationsRepository = (*ReservationsRepository)(nil)

func NewReservationsRepository(db DBI, log *zap.SugaredLogger) *ReservationsRepository {
	return &ReservationsRepository{
		repoMixin: &repoMixin{
			db:  db,
			log: log,
		},
	}
}

func (r *ReservationsRepository) GetReservationByProduct(ctx context.Context, productIDs ...entity.PK) ([]*entity.ProductReservation, error) {
	rows, err := r.DBI(ctx).QueryContext(ctx, "SELECT * FROM product_reservations WHERE product_id IN ($1)", productIDs)
	if err != nil {
		return nil, err
	}
	result := []*entity.ProductReservation{}
	for rows.Next() {
		r := &entity.ProductReservation{}
		if err := r.Scan(rows); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (r *ReservationsRepository) GetReservation(ctx context.Context, id entity.PK) (*entity.ProductReservation, error) {
	rows, err := r.DBI(ctx).QueryContext(ctx, "SELECT * FROM product_reservations WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	var p *entity.ProductReservation
	for rows.Next() {
		p = &entity.ProductReservation{}
		if err := p.Scan(rows); err != nil {
			return nil, err
		}
	}
	if p == nil {
		return nil, errors.New("product reservation by id not found")
	}
	return p, nil
}

func (r *ReservationsRepository) CreateReservation(ctx context.Context, reservations ...*entity.ProductReservation) ([]*entity.ProductReservation, error) {
	q := strings.Builder{}
	q.WriteString("INSERT INTO product_reservations (storage_id, product_id, amount) VALUES ")
	argB := argBuilder{}
	for _, r := range reservations {
		argB.add(r.StorageID, r.ProductID, r.Amount)
	}
	expr, args := argB.done()
	q.WriteString(expr + " RETURNING *")

	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	result := []*entity.ProductReservation{}
	for rows.Next() {
		p := &entity.ProductReservation{}
		if err := p.Scan(rows); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *ReservationsRepository) UpdateReservation(ctx context.Context, reservations ...*entity.ProductReservation) ([]*entity.ProductReservation, error) {
	q := strings.Builder{}
	q.WriteString(`UPDATE product_reservations AS r SET
		amount = c.amount
	FROM (VALUES `)
	argB := argBuilder{}
	for _, r := range reservations {
		argB.add(r.StorageID, r.ProductID, r.Amount)
	}
	expr, args := argB.done()
	q.WriteString(expr + ") AS c (storage_id, product_id, amount) WHERE c.storage_id = r.storage_id AND c.product_id = r.product_id RETURNING *")
	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	result := []*entity.ProductReservation{}
	for rows.Next() {
		p := &entity.ProductReservation{}
		if err := p.Scan(rows); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *ReservationsRepository) DeleteReservation(ctx context.Context, ids ...entity.PK) error {
	_, err := r.DBI(ctx).ExecContext(ctx, "DELETE FROM product_reservations WHERE id IN ($1)", ids)
	return err
}

type IReservationsRepository interface {
	GetReservationByProduct(ctx context.Context, productIDs ...entity.PK) ([]*entity.ProductReservation, error)
	GetReservation(ctx context.Context, id entity.PK) (*entity.ProductReservation, error)
	CreateReservation(ctx context.Context, reservations ...*entity.ProductReservation) ([]*entity.ProductReservation, error)
	UpdateReservation(ctx context.Context, reservations ...*entity.ProductReservation) ([]*entity.ProductReservation, error)
	DeleteReservation(ctx context.Context, ids ...entity.PK) error
}
