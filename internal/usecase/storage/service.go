package storage

import (
	"context"
	"errors"
	"storageapi/internal/database"
	"storageapi/internal/entity"
	"storageapi/internal/repository"
	"storageapi/pkg/algo"
	"storageapi/pkg/stack"

	"go.uber.org/zap"
)

type Service struct {
	repo Repository
	log  *zap.SugaredLogger
}

func NewService(r repository.IRepository, log *zap.SugaredLogger) *Service {
	return &Service{
		repo: r,
		log:  log,
	}
}

func (s *Service) DefineStorageSchema(ctx context.Context, req StorageSchemaReq) error {
	return s.repo.RunInTransaction(ctx, func(ctx database.TxContext, repo repository.IRepository) error {
		// create all storages
		storages := algo.Map(req, func(r StorageSchemaReqItem, _ int) *entity.Storage {
			return &entity.Storage{
				IsAvailable: r.IsAvailable,
			}
		})
		storages, err := repo.CreateStorage(ctx, storages...)
		if err != nil {
			return err
		}
		// insert all products
		products := algo.FlatMap(req, func(r StorageSchemaReqItem, _ int) []*entity.Product {
			return algo.Map(r.Products, func(p StorageSchemaReqProduct, _ int) *entity.Product {
				return &entity.Product{
					Name:   p.Name,
					Vendor: p.Vendor,
					Size:   p.Size,
				}
			})
		})
		products, err = repo.CreateProduct(ctx, products...)
		if err != nil {
			return err
		}
		// create relations
		productsByVendor := map[string]*entity.Product{}
		for _, p := range products {
			productsByVendor[p.Vendor] = p
		}
		available, notAvailable := stack.New[*entity.Storage](), stack.New[*entity.Storage]()
		for _, s := range storages {
			if s.IsAvailable {
				available.Push(s)
				continue
			}
			notAvailable.Push(s)
		}
		relations := make([]*entity.StoredProduct, 0, len(products))
		for _, r := range req {
			var store *entity.Storage
			if r.IsAvailable {
				store, _ = available.Pop()
			} else {
				store, _ = notAvailable.Pop()
			}
			for _, p := range r.Products {
				product, ok := productsByVendor[p.Vendor]
				if !ok {
					return errors.New("not found product by vendor (debug)")
				}
				relations = append(relations, &entity.StoredProduct{
					StorageID: store.ID,
					ProductID: product.ID,
					Amount:    p.Amount,
				})
			}
		}
		_, err = repo.CreateStorageData(ctx, relations...)
		return err
	})
}

func (s *Service) GetStorageSchema(ctx context.Context) (StorageSchemaResp, error) {
	return nil, errors.New("not implemented")
}
