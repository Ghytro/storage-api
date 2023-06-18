package repository

type Repository struct {
}

type IRepository interface {
	IStorageRepository
	IProductRepository
}
