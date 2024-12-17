package main

import (
	"log"
	"net/http"

	config "github.com/Mr-Filatik/go-metrics-collector/internal/config/server"
	"github.com/Mr-Filatik/go-metrics-collector/internal/handler"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/Mr-Filatik/go-metrics-collector/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	conf := config.Initialize()
	repo := repository.New()
	stor := storage.New(repo)

	r := chi.NewRouter()
	r.Get("/", handler.GetAllMetricsHandle(*stor))
	r.Get("/value/{type}/{name}", handler.GetMetricHandle(*stor))
	r.Post("/update/{type}/{name}/{value}", handler.UpdateMetricHandle(*stor))

	log.Printf("Start server on endpoint %v.", conf.ServerAddress)
	err := http.ListenAndServe(conf.ServerAddress, r)
	if err != nil {
		log.Fatal(err)
	}
}
