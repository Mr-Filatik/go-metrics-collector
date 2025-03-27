package repository

import (
	"context"
	"errors"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	log    logger.Logger
	conn   *pgx.Conn
	dbConn string
}

func New(dbConn string, l logger.Logger) (*PostgresRepository, error) {
	conn, err := pgx.Connect(context.Background(), dbConn)
	if err != nil {
		l.Error("Error when connecting to the database", err)
		return nil, errors.New("start connection error")
	}

	// create table
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
		return nil, errors.New("start connection error")
	}

	l.Info("Create PostgresRepository")

	return &PostgresRepository{
		log:    l,
		dbConn: dbConn,
		conn:   conn,
	}, nil
}

func (r *PostgresRepository) Ping() error {
	var version string
	err := r.conn.QueryRow(context.Background(), "SELECT version();").Scan(&version)
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

func (r *PostgresRepository) GetAll() ([]entity.Metrics, error) {
	rows, err := r.conn.Query(context.Background(),
		"SELECT id, mtype, value, delta FROM metrics")
	if err != nil {
		r.log.Error("Error during query execution", err)
		return nil, errors.New("query error")
	}
	defer rows.Close()

	var metrics []entity.Metrics
	for rows.Next() {
		var m entity.Metrics
		err := rows.Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
		if err != nil {
			r.log.Error("Error scanning row", err)
			return nil, errors.New("scan error")
		}
		metrics = append(metrics, m)
	}

	r.log.Debug(
		"Query all metrics from PostgresRepository",
		"count", len(metrics),
	)
	return metrics, nil
}

func (r *PostgresRepository) Get(id string) (entity.Metrics, error) {
	var m entity.Metrics
	err := r.conn.QueryRow(context.Background(),
		"SELECT id, mtype, value, delta FROM metrics WHERE id = $1", id).Scan(&m.ID, &m.MType, &m.Value, &m.Delta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Debug("Metric not found in PostgresRepository", "id", id)
			return entity.Metrics{}, errors.New("metric not found")
		}
		r.log.Error("Error during query execution", err)
		return entity.Metrics{}, errors.New("query error")
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

func (r *PostgresRepository) Create(e entity.Metrics) (entity.Metrics, error) {
	_, err := r.conn.Exec(context.Background(),
		"INSERT INTO metrics (id, mtype, value, delta) VALUES ($1, $2, $3, $4)", e.ID, e.MType, e.Value, e.Delta)
	if err != nil {
		r.log.Error("Error during insert execution", err)
		return entity.Metrics{}, errors.New("insert error")
	}

	r.log.Debug(
		"Creating a new metric in PostgresRepository",
		"id", e.ID,
		"type", e.MType,
		"value", e.Value,
		"delta", e.Delta,
	)
	return e, nil
}

func (r *PostgresRepository) Update(e entity.Metrics) (entity.Metrics, error) {
	_, err := r.conn.Exec(context.Background(),
		"UPDATE metrics SET mtype = $1, value = $2, delta = $3 WHERE id = $4", e.MType, e.Value, e.Delta, e.ID)
	if err != nil {
		r.log.Error("Error during update execution", err)
		return entity.Metrics{}, errors.New("update error")
	}

	r.log.Debug(
		"Updating metric data in PostgresRepository",
		"id", e.ID,
		"type", e.MType,
		"value", e.Value,
		"delta", e.Delta,
	)
	return e, nil
}

func (r *PostgresRepository) Remove(e entity.Metrics) (entity.Metrics, error) {
	_, err := r.conn.Exec(context.Background(), "DELETE FROM metrics WHERE id = $1", e.ID)
	if err != nil {
		r.log.Error("Error during delete execution", err)
		return entity.Metrics{}, errors.New("delete error")
	}

	r.log.Debug(
		"Deleting a metric in PostgresRepository",
		"id", e.ID,
	)
	return e, nil
}

func (r *PostgresRepository) Close() error {
	err := r.conn.Close(context.Background())
	if err != nil {
		r.log.Error("Error when closing the database connection", err)
		return errors.New("close connection error")
	}
	return nil
}
