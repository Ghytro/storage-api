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

func (s *Service) DefineStorageSchema(ctx context.Context, req StorageSchemaReq) (StorageSchemaResp, error) {
	var result StorageSchemaResp
	err := s.repo.RunInTransaction(ctx, func(ctx database.TxContext, repo repository.IRepository) error {
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
			storeResp := StorageSchemaRespItem{
				ID:          store.ID.ToUint(),
				IsAvailable: r.IsAvailable,
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
				storeResp.Products = append(storeResp.Products, StorageSchemaRespProduct{
					ID:     product.ID.ToUint(),
					Name:   p.Name,
					Vendor: p.Vendor,
					Size:   p.Size,
					Amount: p.Amount,
				})
			}
			result = append(result, storeResp)
		}
		_, err = repo.CreateStorageData(ctx, relations...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) GetStorageSchema(ctx context.Context) (StorageSchemaResp, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) GetUnreservedStorage(ctx context.Context, storageID entity.PK) (*StorageSchemaRespItem, error) {
	var (
		reservations []*entity.ProductReservation
		storageData  []*entity.StoredProduct
		productData  []*entity.Product
	)
	err := s.repo.RunInTransaction(ctx, func(ctx database.TxContext, repo repository.IRepository) error {
		storage, err := repo.GetStorage(ctx, storageID)
		if err != nil {
			return err
		}
		if !storage.IsAvailable {
			return errors.New("storage is not available")
		}
		if storageData, err = repo.GetStorageDataByStorage(ctx, storageID); err != nil {
			return err
		}
		if reservations, err = repo.GetReservationByStorage(ctx, storageID); err != nil {
			return err
		}
		productIDs := algo.Map(
			algo.UniqBy(
				storageData,
				func(sd *entity.StoredProduct) entity.PK {
					return sd.ProductID
				},
			),
			func(sd *entity.StoredProduct, _ int) entity.PK {
				return sd.ProductID
			})
		if productData, err = repo.GetProducts(ctx, productIDs...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := &StorageSchemaRespItem{
		ID: storageID.ToUint(),
	}

	reservationsByProductID := map[entity.PK][]*entity.ProductReservation{}
	productByID := map[entity.PK]*entity.Product{}
	for _, resData := range reservations {
		l := reservationsByProductID[resData.ProductID]
		l = append(l, resData)
		reservationsByProductID[resData.ProductID] = l
	}
	for _, p := range productData {
		productByID[p.ID] = p
	}

	for _, st := range storageData {
		product, ok := productByID[st.ProductID]
		if !ok {
			return nil, errors.New("product by id not found (debug)")
		}
		productSchema := StorageSchemaRespProduct{
			ID:     product.ID.ToUint(),
			Name:   product.Name,
			Vendor: product.Vendor,
			Size:   product.Size,
			Amount: st.Amount,
		}
		resData, ok := reservationsByProductID[st.ProductID]
		if ok {
			if res, ok := algo.Find(resData, func(r *entity.ProductReservation) bool {
				return r.StorageID == st.StorageID
			}); ok {
				productSchema.Amount -= res.Amount
			}
		}
		result.Products = append(result.Products, productSchema)
	}
	return result, nil
}
