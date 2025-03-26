package main

import (
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	repository "github.com/Mr-Filatik/go-metrics-collector/internal/repository/memory"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	storage "github.com/Mr-Filatik/go-metrics-collector/internal/storage/file"
)

func main() {
	log := logger.New(logger.LevelInfo)
	defer log.Close()

	conf := config.Initialize()
	repo := repository.New(conf.ConnectionString, log)
	stor := storage.New(conf.FileStoragePath, log)
	srvc := service.New(repo, stor, conf.StoreInterval, log)

	serv := server.NewServer(srvc, log)
	serv.Start(*conf)
}
