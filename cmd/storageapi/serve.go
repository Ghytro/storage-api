package main

import (
	"log"
	"net/rpc"
	"storageapi/internal/api"
	"storageapi/internal/api/reservation"
	"storageapi/internal/api/storage"
	"storageapi/internal/config"
	"storageapi/internal/database"
	"storageapi/internal/repository"
	reservationService "storageapi/internal/usecase/reservation"
	storageService "storageapi/internal/usecase/storage"

	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func serve() *rpc.Server {
	db, err := database.NewDBWithPgx(config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	migrateDB, err := goose.OpenDBWithDriver(config.MigrationDialect, config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := goose.Run("up", migrateDB, config.FixturesPath); err != nil {
		log.Fatal(err)
	}
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zapcore.FatalLevel))
	if err != nil {
		log.Fatalf("Can't initialize zap logger: %v", err)
	}
	sugar := logger.Sugar()

	repo := repository.NewRepository(db, sugar)

	storageService := storageService.NewService(repo, sugar)
	reservationService := reservationService.NewService(repo, sugar)

	apiConf := api.ApiConf{
		RequestHandleTimeout: config.RequestHandleTimeout,
	}
	storageApi := storage.NewAPI(sugar, storageService, apiConf)
	reservationApi := reservation.NewAPI(sugar, reservationService, apiConf)
	server := newServer(storageApi, reservationApi)
	return server
}

func newServer(api ...interface{}) *rpc.Server {
	server := rpc.NewServer()
	for _, a := range api {
		server.Register(a)
	}
	return server
}
