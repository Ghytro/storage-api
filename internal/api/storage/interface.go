package storage

import (
	"context"
	"storageapi/internal/usecase/storage"
)

type UseCase interface {
	DefineStorageSchema(ctx context.Context, req storage.StorageSchemaReq) error
	GetStorageSchema(ctx context.Context) (storage.StorageSchemaResp, error)
}
