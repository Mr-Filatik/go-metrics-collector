package main

import (
	"fmt"

	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
	repositoryMemory "github.com/Mr-Filatik/go-metrics-collector/internal/repository/memory"
	repositoryPostgres "github.com/Mr-Filatik/go-metrics-collector/internal/repository/postgres"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server"
	config "github.com/Mr-Filatik/go-metrics-collector/internal/server/config"
	"github.com/Mr-Filatik/go-metrics-collector/internal/service"
	storage "github.com/Mr-Filatik/go-metrics-collector/internal/storage/file"
)

// go run -ldflags "-X main.buildVersion=v2.0.0 -X main.buildDate=2025-07-07 -X main.buildCommit=98d1d98".
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	log := logger.New(logger.LevelDebug)
	defer log.Close()

	log.Info(fmt.Sprintf("Build version: %v", buildVersion))
	log.Info(fmt.Sprintf("Build date: %v", buildDate))
	log.Info(fmt.Sprintf("Build commit: %v", buildCommit))

	conf := config.Initialize()

	var srvc *service.Service
	if conf.ConnectionString != "" {
		repo, err := repositoryPostgres.New(conf.ConnectionString, log)
		if err != nil {
			panic(err.Error())
		}
		defer repo.Close()
		srvc = service.New(repo, nil, 0, log)
	} else {
		repo := repositoryMemory.New(conf.ConnectionString, log)
		stor := storage.New(conf.FileStoragePath, log)
		srvc = service.New(repo, stor, conf.StoreInterval, log)
	}

	serv := server.NewServer(srvc, conf.HashKey, log)
	serv.Start(conf.ServerAddress, conf.Restore)
}
