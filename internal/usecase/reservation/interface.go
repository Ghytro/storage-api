package reservation

import "storageapi/internal/repository"

type Repository interface {
	repository.IRepoMixin
	// todo: methods for repo provider
}
