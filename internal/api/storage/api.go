package storage

import "go.uber.org/zap"

type API struct {
	log     *zap.SugaredLogger
	service UseCase
}

func NewAPI(log *zap.SugaredLogger, s UseCase) *API {
	return &API{
		log:     log,
		service: s,
	}
}
