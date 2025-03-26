package repository

import (
	"context"
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repository"
	"github.com/jackc/pgx/v5"
)

type MemoryRepository struct {
	log    logger.Logger
	dbConn string
	datas  []entity.Metrics
}

func New(dbConn string, l logger.Logger) *MemoryRepository {
	return &MemoryRepository{
		datas:  make([]entity.Metrics, 0),
		log:    l,
		dbConn: dbConn,
	}
}

func (r *MemoryRepository) Ping() error {
	conn, err := pgx.Connect(context.Background(), r.dbConn)
	if err != nil {
		r.log.Error("Error when connecting to the database", err)
		return errors.New("connect error")
	}
	// defer conn.Close(context.Background())
	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			r.log.Error("Error when close connecting to the database", err)
		}
	}()

	var version string
	err = conn.QueryRow(context.Background(), "SELECT version();").Scan(&version)
	if err != nil {
		r.log.Error("Error during query execution", err)
		return errors.New("query error")
	}

	r.log.Info(
		"Successful connection",
		"version", version,
	)
	return nil
}

func (r *MemoryRepository) GetAll() ([]entity.Metrics, error) {
	if r.datas != nil {
		r.log.Debug(
			"Query all metrics from MemRepository",
			"count", len(r.datas),
		)
		return r.datas, nil
	}
	r.log.Debug("Querying empty data in MemRepository")
	return make([]entity.Metrics, 0), nil
}

func (r *MemoryRepository) Get(id string) (entity.Metrics, error) {
	for _, v := range r.datas {
		if v.ID == id {
			r.log.Debug(
				"Getting metric from MemRepository",
				"id", v.ID,
				"type", v.MType,
				"value", v.Value,
				"delta", v.Delta,
			)

			return entity.Metrics{
				ID:    v.ID,
				MType: v.MType,
				Value: v.Value,
				Delta: v.Delta,
			}, nil
		}
	}
	return entity.Metrics{}, errors.New(repository.ErrorMetricNotFound)
}

func (r *MemoryRepository) Create(e entity.Metrics) (entity.Metrics, error) {
	r.datas = append(r.datas, e)

	r.log.Debug(
		"Creating a new metric in MemRepository",
		"id", e.ID,
		"type", e.MType,
		"value", e.Value,
		"delta", e.Delta,
	)

	return entity.Metrics{
		ID:    e.ID,
		MType: e.MType,
		Value: e.Value,
		Delta: e.Delta,
	}, nil
}

func (r *MemoryRepository) Update(e entity.Metrics) (entity.Metrics, error) {
	for i, v := range r.datas {
		if v.ID == e.ID {
			item := &r.datas[i]
			item.Value = e.Value
			item.MType = e.MType
			item.Delta = e.Delta

			r.log.Debug(
				"Updating metric data in MemRepository",
				"id", item.ID,
				"type", item.MType,
				"value", item.Value,
				"delta", item.Delta,
			)

			return entity.Metrics{
				ID:    item.ID,
				MType: item.MType,
				Value: item.Value,
				Delta: item.Delta,
			}, nil
		}
	}
	return entity.Metrics{}, errors.New(repository.ErrorMetricNotFound)
}

func (r *MemoryRepository) Remove(e entity.Metrics) (entity.Metrics, error) {
	newDatas := make([]entity.Metrics, (len(r.datas) - 1))
	index := 0
	for i, v := range r.datas {
		if v.ID != e.ID {
			newDatas[index] = r.datas[i]
			index++
		} else {
			r.log.Debug(
				"Deleting a metric in MemRepository",
				"id", e.ID,
			)

			index++
		}
	}
	r.datas = newDatas
	return e, nil
}
