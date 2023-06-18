package reservation

import (
	"errors"
)

// create reservation request

type ReserveProductsReq []ReserveProductsReqItem

func (req ReserveProductsReq) Validate() error {
	for _, r := range req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	{
		m := map[uint]struct{}{}
		for _, r := range req {
			m[r.ID] = struct{}{}
		}
		if len(m) != len(req) {
			return errors.New("all the product ids must be unique")
		}
	}
	return nil
}

type ReserveProductsReqItem struct {
	ID     uint `json:"id"`
	Amount uint `json:"amount"`
}

func (req ReserveProductsReqItem) Validate() error {
	if req.Amount == 0 {
		return errors.New("amount for product reservation cannot be nil")
	}
	return nil
}

// undo reservation

type UndoReservationReq []UndoReservationReqItem

type UndoReservationReqItem ReserveProductsReqItem
