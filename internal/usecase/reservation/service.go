package reservation

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"storageapi/internal/database"
	"storageapi/internal/entity"
	"storageapi/internal/repository"
	"storageapi/pkg/algo"

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

func (s *Service) ReserveProducts(ctx context.Context, req ReserveProductsReq) error {
	// todo: make this method singlethreaded only by attaching to channel
	productIDs := algo.Map(req, func(r ReserveProductsReqItem, _ int) entity.PK {
		return entity.PK(r.ID)
	})

	stResData, err := s.getStorageDataWithReservation(ctx, productIDs...)
	if err != nil {
		return err
	}

	addedReservations, err := s.getReservationsToAdd(req, stResData.storeData, stResData.reservations)
	if err != nil {
		return err
	}

	// upsert all the reservations to database
	updated := []*entity.ProductReservation{}
	created := []*entity.ProductReservation{}
	for _, r := range addedReservations {
		reservations, ok := stResData.reservations[r.ProductID]
		if !ok {
			return errors.New("cannot find reservation data by product id, returned by added reservations (debug)")
		}
		res, ok := algo.Find(reservations, func(_r *entity.ProductReservation) bool {
			return _r.StorageID == r.StorageID
		})
		if ok {
			// create an entity that will be updated to new amount
			e := *res
			e.Amount += r.Amount
			updated = append(updated, &e)
		} else {
			created = append(created, r)
		}
	}
	return s.repo.RunInTransaction(ctx, func(ctx database.TxContext, repo repository.IRepository) error {
		if _, err := repo.UpdateReservation(ctx, updated...); err != nil {
			return err
		}
		_, err := repo.CreateReservation(ctx, created...)
		return err
	})
}

type storageDataReservation struct {
	storeData    map[entity.PK][]*entity.StoredProduct
	reservations map[entity.PK][]*entity.ProductReservation
}

func (s *Service) getStorageDataWithReservation(ctx context.Context, productIDs ...entity.PK) (*storageDataReservation, error) {
	var (
		storedData      []*entity.StoredProduct
		reservationData []*entity.ProductReservation
	)
	err := s.repo.RunInTransaction(ctx, func(ctx database.TxContext, repo repository.IRepository) (err error) {
		// fetch data associated with products from request
		storedData, err = repo.GetStorageDataByProduct(ctx, productIDs...)
		if err != nil {
			return err
		}
		reservationData, err = repo.GetReservationByProduct(ctx, productIDs...)
		return err
	})
	if err != nil {
		return nil, err
	}

	// map aggregation for speed up
	storedDataByProductID := map[entity.PK][]*entity.StoredProduct{}
	reservationDataByProductID := map[entity.PK][]*entity.ProductReservation{}
	for _, stData := range storedData {
		l := storedDataByProductID[stData.ProductID]
		l = append(l, stData)
		storedDataByProductID[stData.ProductID] = l
	}
	for _, resData := range reservationData {
		l := reservationDataByProductID[resData.ProductID]
		l = append(l, resData)
		reservationDataByProductID[resData.ProductID] = l
	}

	return &storageDataReservation{
		storeData:    storedDataByProductID,
		reservations: reservationDataByProductID,
	}, nil
}

// according to user request data and stored products data determine how to reserve the products
// (cpu only computations + greedy algorithm)
func (s *Service) getReservationsToAdd(
	req ReserveProductsReq,
	storedDataByProductID map[entity.PK][]*entity.StoredProduct,
	reservationDataByProductID map[entity.PK][]*entity.ProductReservation,
) ([]*entity.ProductReservation, error) {
	// check if we can reserve the needed amount of products
	productIDs := algo.Map(req, func(r ReserveProductsReqItem, _ int) entity.PK {
		return entity.PK(r.ID)
	})
	unreservedProducts, err := s.getFreeReservations(storedDataByProductID, reservationDataByProductID, productIDs...)
	if err != nil {
		return nil, err
	}
	// check that we have enough space to reserve products
	{
		unreservedByProductID := map[entity.PK]uint{}
		for _, p := range unreservedProducts {
			unreservedByProductID[p.productID] += p.amount
		}
		for _, reqItem := range req {
			pID := entity.PK(reqItem.ID)
			unreservedAmount, ok := unreservedByProductID[pID]
			if !ok {
				return nil, errors.New("unreserved not found by product_id %d (debug)")
			}
			if unreservedAmount < reqItem.Amount {
				return nil, fmt.Errorf(
					"cannot reserve more than %d for product with id %d (tried to reserve %d)",
					unreservedAmount,
					reqItem.ID,
					reqItem.Amount,
				)
			}
		}
	}

	// descending order by free space available for reservation
	sort.Slice(unreservedProducts, func(i, j int) bool {
		return unreservedProducts[i].amount > unreservedProducts[j].amount
	})

	// reserve the products in determined order
	// something like the greedy algorithm
	addedReservations := []*entity.ProductReservation{}
	for _, r := range req {
		unreserved := algo.Filter(unreservedProducts, func(u *unreservedProduct, _ int) bool {
			return u.productID == entity.PK(r.ID)
		})
		amount := r.Amount
		for _, u := range unreserved {
			min := algo.Min(amount, u.amount)
			amount -= min
			u.amount -= min
			addedReservations = append(addedReservations, &entity.ProductReservation{
				ProductID: u.productID,
				StorageID: u.storageID,
				Amount:    min,
			})
			if amount == 0 {
				break
			}
		}
	}
	return addedReservations, nil
}

type unreservedProduct struct {
	storageID, productID entity.PK
	amount               uint
}

func (s *Service) getFreeReservations(
	storedDataByProductID map[entity.PK][]*entity.StoredProduct,
	reservationDataByProductID map[entity.PK][]*entity.ProductReservation,
	productIDs ...entity.PK,
) ([]*unreservedProduct, error) {
	result := []*unreservedProduct{}
	for _, productID := range productIDs {
		stData, ok := storedDataByProductID[productID]
		if !ok {
			return nil, errors.New("storage data not found (debug)")
		}
		resData := reservationDataByProductID[productID]

		// determine amount which can be reserved for each product by storage id
		for _, st := range stData {
			var reservedAmount uint
			if reserved, ok := algo.Find(resData, func(r *entity.ProductReservation) bool {
				return r.ProductID == st.ProductID && r.StorageID == st.StorageID
			}); ok {
				reservedAmount = reserved.Amount
			}
			result = append(result, &unreservedProduct{
				productID: st.ProductID,
				storageID: st.StorageID,
				amount:    st.Amount - reservedAmount,
			})
		}
	}
	return result, nil
}

func (s *Service) UndoReserve(ctx context.Context, req UndoReservationReq) error {
	productIDs := algo.Map(req, func(r UndoReservationReqItem, _ int) entity.PK {
		return entity.PK(r.ID)
	})

	stResData, err := s.getStorageDataWithReservation(ctx, productIDs...)
	if err != nil {
		return err
	}

	freedReservations, err := s.getReservationsToFree(req, stResData.storeData, stResData.reservations)
	if err != nil {
		return err
	}

	updated := []*entity.ProductReservation{}
	deletedIDs := []entity.PK{}
	for _, freedRes := range freedReservations {
		reservations, ok := stResData.reservations[freedRes.ProductID]
		if !ok {
			return errors.New("UndoReserve: reservation not found (debug)")
		}
		res, ok := algo.Find(reservations, func(st *entity.ProductReservation) bool {
			return st.StorageID == freedRes.StorageID
		})
		if !ok {
			return errors.New("UndoReserve: reservation not found by storage id (debug)")
		}
		if res.Amount != freedRes.Amount {
			e := *res
			e.Amount -= freedRes.Amount
			updated = append(updated, &e)
		} else {
			deletedIDs = append(deletedIDs, res.ID)
		}
	}

	return s.repo.RunInTransaction(ctx, func(ctx database.TxContext, repo repository.IRepository) error {
		if _, err := repo.UpdateReservation(ctx, updated...); err != nil {
			return err
		}
		return repo.DeleteReservation(ctx, deletedIDs...)
	})
}

// according to user request data and stored products data determine how to undo the reservation of products
// (cpu only computations + greedy algorithm)
func (s *Service) getReservationsToFree(
	req UndoReservationReq,
	storedDataByProductID map[entity.PK][]*entity.StoredProduct,
	reservationDataByProductID map[entity.PK][]*entity.ProductReservation,
) ([]*entity.ProductReservation, error) {
	productIDs := algo.Map(req, func(r UndoReservationReqItem, _ int) entity.PK {
		return entity.PK(r.ID)
	})
	unreservedProducts, err := s.getFreeReservations(storedDataByProductID, reservationDataByProductID, productIDs...)
	if err != nil {
		return nil, err
	}

	sort.Slice(unreservedProducts, func(i, j int) bool {
		return unreservedProducts[i].amount < unreservedProducts[j].amount
	})

	freedReservations := []*entity.ProductReservation{}
	for _, r := range req {
		unreserved := algo.Filter(unreservedProducts, func(u *unreservedProduct, _ int) bool {
			return u.productID == entity.PK(r.ID)
		})
		amount := r.Amount
		storeData, ok := storedDataByProductID[entity.PK(r.ID)]
		if !ok {
			return nil, errors.New("getReservationsToFree: reservation not found (debug)")
		}
		for _, u := range unreserved {
			st, ok := algo.Find(storeData, func(r *entity.StoredProduct) bool {
				return r.StorageID == u.storageID
			})
			if !ok {
				return nil, errors.New("getReservationsToFree: reservation by storage id not found (debug)")
			}
			min := algo.Min(amount, st.Amount-u.amount)
			amount -= min
			u.amount += min
			freedReservations = append(freedReservations, &entity.ProductReservation{
				ProductID: u.productID,
				StorageID: u.storageID,
				Amount:    min,
			})
			if amount == 0 {
				break
			}
		}
	}
	return freedReservations, nil
}
