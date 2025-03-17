package main

import (
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
)

func main() {
	log := logger.New(logger.LevelDebug)
	defer log.Close()

	conf := config.Initialize()
	repo := repository.New()
	stor := storage.New(repo, log, conf.FileStoragePath)

	serv := server.NewServer(stor, log)
	log.Debug("AAAAA")
	log.Info("BBBB")

	serv.Start(*conf)
}
