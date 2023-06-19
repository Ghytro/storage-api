package storage

import (
	"context"
	"storageapi/internal/entity"
	"storageapi/internal/usecase/storage"
)

type UseCase interface {
	DefineStorageSchema(ctx context.Context, req storage.StorageSchemaReq) (storage.StorageSchemaResp, error)
	GetStorageSchema(ctx context.Context) (storage.StorageSchemaResp, error)
	GetUnreservedStorage(ctx context.Context, storageID entity.PK) (*storage.StorageSchemaRespItem, error)
}

var _ UseCase = (*storage.Service)(nil)
