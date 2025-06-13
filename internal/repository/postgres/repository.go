// Пакет repository предоставляет конкретную реализацию репозитория
// для доступа к postres-хранилищу.
package repository

import (
	"context"
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/Mr-Filatik/go-metrics-collector/internal/repeater"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrConnectionStart = errors.New("start connection error")
	ErrQueryRun        = errors.New("query run error")
	ErrScanData        = errors.New("scan data error")
)

type PostgresRepository struct {
	log    logger.Logger
	conn   *pgxpool.Pool
	dbConn string
}

func New(dbConn string, l logger.Logger) (*PostgresRepository, error) {
	conn, err := repeater.New[string, *pgxpool.Pool](l).
		SetFunc(func(c string) (*pgxpool.Pool, error) {
			poolConfig, err := pgxpool.ParseConfig(dbConn)
			if err != nil {
				l.Error("Error parsing connection string", err)
				return nil, ErrConnectionStart
			}

			poolConfig.MaxConns = 10
			poolConfig.MinConns = 2

			conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
			if err != nil {
				l.Error("Error creating connection pool", err)
				return nil, ErrConnectionStart
			}

			err = conn.Ping(context.Background())
			if err != nil {
				l.Error("Error during ping", err)
				return nil, ErrConnectionStart
			}

			// Create table
			query := `
    		CREATE TABLE IF NOT EXISTS metrics (
        		id TEXT PRIMARY KEY,
        		mtype TEXT NOT NULL,
        		value DOUBLE PRECISION,
        		delta BIGINT
    		);
    		`
			_, eerr := conn.Exec(context.Background(), query)
			if eerr != nil {
				l.Error("Error during table creation", eerr)
				return nil, ErrQueryRun
			}
			return conn, nil
		}).
		SetCondition(func(err error) bool {
			return !errors.Is(err, ErrConnectionStart) // I didn't pack the bugs due to time constraints.
		}).
		Run(dbConn)

	if err != nil {
		l.Error("Error when connecting to the database", err)
		return nil, ErrConnectionStart
	}

	l.Info("Create PostgresRepository")

	return &PostgresRepository{
		log:    l,
		dbConn: dbConn,
		conn:   conn,
	}, nil
}

func (r *PostgresRepository) Ping() error {
	err := r.conn.Ping(context.Background())
	if err != nil {
		r.log.Error("Error during ping", err)
		return ErrQueryRun
	}

	r.log.Info("Successful ping")
	return nil
}

func (r *PostgresRepository) GetAll() ([]entity.Metrics, error) {
	rows, err := r.conn.Query(context.Background(),
		"SELECT id, mtype, value, delta FROM metrics")
	if err != nil {
		r.log.Error("Error during query execution", err)
		return nil, ErrQueryRun
	}
	defer rows.Close()

	var errs []error
	var metrics []entity.Metrics
	for rows.Next() {
		var m entity.Metrics
		err := rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
		if err != nil {
			r.log.Error("Error scanning row", err)
			errs = append(errs, ErrScanData)
			metrics = append(metrics, m)
		}
	}
	if errs != nil {
		return nil, errors.Join(errs...)
	}

	r.log.Debug(
		"Query all metrics from PostgresRepository",
		"count", len(metrics),
	)
	return metrics, nil
}

func (r *PostgresRepository) GetByID(id string) (entity.Metrics, error) {
	var m entity.Metrics
	err := r.conn.QueryRow(context.Background(),
		"SELECT id, mtype, value, delta FROM metrics WHERE id = $1", id).Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Debug("Metric not found in PostgresRepository", "id", id)
			return entity.Metrics{}, errors.New("metric not found")
		}
		r.log.Error("Error during query execution", err)
		return entity.Metrics{}, ErrQueryRun
	}

	r.log.Debug(
		"Getting metric from PostgresRepository",
		"id", m.ID,
		"type", m.MType,
		"value", m.Value,
		"delta", m.Delta,
	)
	return m, nil
}

func (r *PostgresRepository) Create(e entity.Metrics) (string, error) {
	_, err := r.conn.Exec(context.Background(),
		"INSERT INTO metrics (id, mtype, value, delta) VALUES ($1, $2, $3, $4)", e.ID, e.MType, e.Value, e.Delta)
	if err != nil {
		r.log.Error("Error during insert execution", err)
		return "", errors.New("insert error")
	}

	r.log.Debug(
		"Creating a new metric in PostgresRepository",
		"id", e.ID,
		"type", e.MType,
		"value", e.Value,
		"delta", e.Delta,
	)
	return e.ID, nil
}

func (r *PostgresRepository) Update(e entity.Metrics) (float64, int64, error) {
	_, err := r.conn.Exec(context.Background(),
		"UPDATE metrics SET mtype = $1, value = $2, delta = $3 WHERE id = $4", e.MType, e.Value, e.Delta, e.ID)
	if err != nil {
		r.log.Error("Error during update execution", err)
		return 0, 0, errors.New("update error")
	}

	r.log.Debug(
		"Updating metric data in PostgresRepository",
		"id", e.ID,
		"type", e.MType,
		"value", e.Value,
		"delta", e.Delta,
	)
	value := float64(0)
	if e.Value != nil {
		value = *e.Value
	}
	delta := int64(0)
	if e.Delta != nil {
		delta = *e.Delta
	}
	return value, delta, nil
}

func (r *PostgresRepository) Remove(e entity.Metrics) (string, error) {
	_, err := r.conn.Exec(context.Background(), "DELETE FROM metrics WHERE id = $1", e.ID)
	if err != nil {
		r.log.Error("Error during delete execution", err)
		return "", errors.New("delete error")
	}

	r.log.Debug(
		"Deleting a metric in PostgresRepository",
		"id", e.ID,
	)
	return e.ID, nil
}

func (r *PostgresRepository) Close() {
	r.conn.Close()
}
