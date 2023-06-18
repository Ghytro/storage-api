package storage

import "storageapi/internal/repository"

type Repository interface {
	repository.IRepoMixin
}
