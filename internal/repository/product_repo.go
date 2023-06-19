package repository

import (
	"context"
	"errors"
	"storageapi/internal/entity"
	"strings"

	"go.uber.org/zap"
)

type ProductRepository struct {
	*repoMixin
}

var _ IProductRepository = (*ProductRepository)(nil)

func NewProductRepository(db DBI, log *zap.SugaredLogger) *ProductRepository {
	return &ProductRepository{
		repoMixin: &repoMixin{
			db:  db,
			log: log,
		},
	}
}

func (r *ProductRepository) ListProducts(ctx context.Context, filter *ListProductFilter) ([]*entity.Product, error) {
	q := strings.Builder{}
	q.WriteString(
		`SELECT products.*
		FROM products
		LEFT JOIN stored_products
		ON stored_products.product_id = products.id
		WHERE TRUE `,
	)
	args, err := filter.apply(&q)
	if err != nil {
		return nil, err
	}
	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}

	return entity.ScannedRows[*entity.Product](rows)
}

func (r *ProductRepository) GetProduct(ctx context.Context, id entity.PK) (*entity.Product, error) {
	rows, err := r.DBI(ctx).QueryContext(ctx, "SELECT * FROM products WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	var p *entity.Product
	for rows.Next() {
		p = &entity.Product{}
		if err := p.Scan(rows); err != nil {
			return nil, err
		}
	}
	if p == nil {
		return nil, errors.New("product by id not found")
	}
	return p, nil
}

func (r *ProductRepository) GetProducts(ctx context.Context, id ...entity.PK) ([]*entity.Product, error) {
	rows, err := r.DBI(ctx).QueryContext(ctx, "SELECT * FROM products WHERE id IN ($1)", id)
	if err != nil {
		return nil, err
	}
	return entity.ScannedRows[*entity.Product](rows)
}

func (r *ProductRepository) CreateProduct(ctx context.Context, products ...*entity.Product) ([]*entity.Product, error) {
	q := strings.Builder{}
	q.WriteString("INSERT INTO products (name, vendor, size) VALUES ")
	argB := argBuilder{}
	for _, p := range products {
		argB.add(p.Name, p.Vendor, p.Size)
	}
	expr, args := argB.done()
	q.WriteString(expr + " RETURNING *")

	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	return entity.ScannedRows[*entity.Product](rows)
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, products ...*entity.Product) ([]*entity.Product, error) {
	q := strings.Builder{}
	q.WriteString(`UPDATE products AS p SET
		name = c.name,
		vendor = c.vendor,
		size = c.size
	FROM (VALUES `)
	argB := argBuilder{}
	for _, p := range products {
		argB.add(p.ID, p.Name, p.Vendor, p.Size)
	}
	expr, args := argB.done()
	q.WriteString(expr + ") AS c(id, name, vendor, size) WHERE c.id = p.id RETURNING *")
	rows, err := r.DBI(ctx).QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, err
	}
	return entity.ScannedRows[*entity.Product](rows)
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, ids ...entity.PK) error {
	_, err := r.DBI(ctx).ExecContext(ctx, "DELETE FROM products WHERE id IN ($1)", ids)
	return err
}

func (r *ProductRepository) TruncateProducts(ctx context.Context) error {
	_, err := r.DBI(ctx).ExecContext(ctx, "TRUNCATE TABLE products")
	return err
}

type ListProductFilter struct {
	Vendors []string
}

func (f *ListProductFilter) apply(q *strings.Builder) ([]interface{}, error) {
	args := []interface{}{}
	if len(f.Vendors) > 0 {
		q.WriteString("AND products.vendor IN ($1) ")
		args = append(args, f.Vendors)
	}
	return args, nil
}

type IProductRepository interface {
	ListProducts(ctx context.Context, filter *ListProductFilter) ([]*entity.Product, error)
	GetProduct(ctx context.Context, id entity.PK) (*entity.Product, error)
	GetProducts(ctx context.Context, id ...entity.PK) ([]*entity.Product, error)
	CreateProduct(ctx context.Context, products ...*entity.Product) ([]*entity.Product, error)
	UpdateProduct(ctx context.Context, products ...*entity.Product) ([]*entity.Product, error)
	DeleteProduct(ctx context.Context, ids ...entity.PK) error
	TruncateProducts(ctx context.Context) error
}
