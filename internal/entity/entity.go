package entity

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

type Product struct {
	baseEntity
	Name   string `db:"name" json:"name"`
	Vendor string `db:"vendor" json:"vendor"`
	Size   string `db:"size" json:"size"`
}

type StoredProduct struct {
	StorageID PK   `db:"storage_id"`
	ProductID PK   `db:"product_id"`
	Amount    uint `db:"amount"`
}

type ProductReservation struct {
	StorageID PK   `db:"storage_id"`
	ProductID PK   `db:"product_id"`
	Amount    uint `db:"amount"`
}
