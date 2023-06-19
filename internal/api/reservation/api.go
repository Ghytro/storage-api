package reservation

import (
	"context"
	"storageapi/internal/api"
	"storageapi/internal/usecase/reservation"
	"time"

	"go.uber.org/zap"
)

type API struct {
	log            *zap.SugaredLogger
	service        UseCase
	requestTimeout time.Duration
}

func NewAPI(log *zap.SugaredLogger, s UseCase, conf api.ApiConf) *API {
	return &API{
		log:            log,
		service:        s,
		requestTimeout: conf.RequestHandleTimeout,
	}
}

func (a API) CreateReservation(request *reservation.ReserveProductsReq, response *api.Empty) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()
	if err := a.service.ReserveProducts(ctx, *request); err != nil {
		return err
	}
	*response = api.Empty{}
	return nil
}

func (a API) UndoReservation(request *reservation.UndoReservationReq, response *api.Empty) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()
	if err := a.service.UndoReserve(ctx, *request); err != nil {
		return err
	}
	*response = api.Empty{}
	return nil
}
