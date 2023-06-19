package storage

import (
	"context"
	"storageapi/internal/api"
	"storageapi/internal/entity"
	"storageapi/internal/usecase/storage"
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

func (a API) DefineStorageSchema(request *storage.StorageSchemaReq, response *storage.StorageSchemaResp) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()
	resp, err := a.service.DefineStorageSchema(ctx, *request)
	if err != nil {
		return err
	}
	*response = resp
	return nil
}

func (a API) GetUnreservedStorage(request *GetUnreservedStorageReq, response *storage.StorageSchemaRespItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.requestTimeout)
	defer cancel()
	resp, err := a.service.GetUnreservedStorage(ctx, entity.PK(request.StorageID))
	if err != nil {
		return err
	}
	*response = *resp
	return nil
}
