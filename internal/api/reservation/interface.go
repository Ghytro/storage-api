package reservation

import (
	"context"
	"storageapi/internal/usecase/reservation"
)

type UseCase interface {
	ReserveProducts(ctx context.Context, req reservation.ReserveProductsReq) error
	UndoReserve(ctx context.Context, req reservation.UndoReservationReq) error
}

var _ UseCase = (*reservation.Service)(nil)
