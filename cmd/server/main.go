package main

import (
	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
)

func main() {
	log := logger.New()
	defer log.Close()

	conf := config.Initialize()
	repo := repository.New()
	stor := storage.New(repo, log)

	serv := server.NewServer(stor, log)
	serv.Start(*conf)
}
