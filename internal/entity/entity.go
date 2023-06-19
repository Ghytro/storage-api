package entity

import "github.com/jackc/pgx"

type PK uint

func (p PK) ToUint() uint {
	return uint(p)
}

type baseEntity struct {
	ID PK `db:"id" json:"id"`
}

type Storage struct {
	baseEntity
	IsAvailable bool `db:"is_available" json:"is_available"`
}

var _ IEntity = (*Storage)(nil)

func (s *Storage) Scan(rows *pgx.Rows) error {
	return rows.Scan(&s.ID, &s.IsAvailable)
}

type Product struct {
	baseEntity
	Name   string `db:"name" json:"name"`
	Vendor string `db:"vendor" json:"vendor"`
	Size   string `db:"size" json:"size"`

	Storage *Storage `json:"-"` // relation
}

var _ IEntity = (*Product)(nil)

func (p *Product) Scan(rows *pgx.Rows) error {
	return rows.Scan(&p.ID, &p.Name, &p.Vendor, &p.Size)
}

type StoredProduct struct {
	ID        PK   `db:"id"`
	StorageID PK   `db:"storage_id"`
	ProductID PK   `db:"product_id"`
	Amount    uint `db:"amount"`
}

var _ IEntity = (*StoredProduct)(nil)

func (sp *StoredProduct) Scan(rows *pgx.Rows) error {
	return rows.Scan(&sp.ID, &sp.StorageID, &sp.ProductID, &sp.Amount)
}

type ProductReservation struct {
	ID        PK   `db:"id"`
	StorageID PK   `db:"storage_id"`
	ProductID PK   `db:"product_id"`
	Amount    uint `db:"amount"`
}

var _ IEntity = (*ProductReservation)(nil)

func (pr *ProductReservation) Scan(rows *pgx.Rows) error {
	return rows.Scan(&pr.ID, &pr.StorageID, &pr.ProductID, &pr.Amount)
}

func ScannedRows[T IEntity](rows *pgx.Rows) (result []T, err error) {
	for rows.Next() {
		var t T
		if err = t.Scan(rows); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return
}

type IEntity interface {
	Scan(rows *pgx.Rows) error
}
