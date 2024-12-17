package main

import (
	config "github.com/Mr-Filatik/go-metrics-collector/internal/config/server"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/server"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
)

func main() {
	conf := config.Initialize()
	repo := repository.New()
	stor := storage.New(repo)

	serv := server.NewServer(stor)
	serv.Start(*conf)
}
