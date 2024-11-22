package repository

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"rates/internal/entity"
	"rates/internal/infrastructure/metrics"
	"rates/pkg/logger"

	_ "github.com/lib/pq"
)

var (
	log = logger.Logger().Named("repository").Sugar()
)

func NewPostgresClient() (*sql.DB, error) {
	dbHost := flag.String("host", os.Getenv("POSTGRES_HOST"), "Postgres host")
	dbPort := flag.String("port", os.Getenv("POSTGRES_PORT"), "Postgres port")
	dbUser := flag.String("user", os.Getenv("POSTGRES_USER"), "Postgres user")
	dbName := flag.String("dbname", os.Getenv("POSTGRES_DB"), "Postgres database name")
	dbPassword := flag.String("password", os.Getenv("POSTGRES_PASSWORD"), "Postgres password")
	flag.Parse()

	// // Проверка, если флаги пустые, берем значения из переменных окружения
	// if *dbHost == "" || *dbPort == "" || *dbUser == "" || *dbName == "" || *dbPassword == "" {
	// 	*dbHost = os.Getenv("POSTGRES_HOST")
	// 	*dbPort = os.Getenv("POSTGRES_PORT")
	// 	*dbUser = os.Getenv("POSTGRES_USER")
	// 	*dbName = os.Getenv("POSTGRES_DB")
	// 	*dbPassword = os.Getenv("POSTGRES_PASSWORD")
	// }

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		*dbHost, *dbPort, *dbUser, *dbPassword, *dbName)

	db, err := sql.Open("postgres", dns)

	if err != nil {
		log.Errorf("failed to open database conection: %w", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Errorf("failed to ping database: %w", err)
		return nil, err
	}

	return db, nil
}

type Repositer interface {
	InsertAsks(ctx context.Context, dept entity.Depth) error
	InsertBids(ctx context.Context, dept entity.Depth) error
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) InsertAsks(ctx context.Context, dept entity.Depth) error {
	tx, err := r.db.Begin()
	if err != nil {
		metrics.StatusRequestToDB("begin_transaction", "error")
		log.Errorf("Failed to begin transaction for InsertAsks: %v", err)
		return err
	}
	metrics.StatusRequestToDB("begin_transaction", "success")

	log.Infof("Transaction started for InsertAsks")

	err = insertOrder(ctx, tx, dept.Asks, dept.Timestamp, "asks")
	if err != nil {
		_ = tx.Rollback()
		metrics.StatusRequestToDB("insert_order", "error")
		log.Errorf("Failed to insert asks data: %v", err)
		return err
	}
	metrics.StatusRequestToDB("insert_order", "success")

	if err := tx.Commit(); err != nil {
		metrics.StatusRequestToDB("commit_transaction", "error")
		log.Errorf("Failed to commit transaction for InsertAsks: %v", err)
		return err
	}
	metrics.StatusRequestToDB("commit_transaction", "success")
	log.Info("InsertAsks transaction committed successfully")
	return nil
}

func (r *Repository) InsertBids(ctx context.Context, dept entity.Depth) error {
	tx, err := r.db.Begin()
	if err != nil {
		metrics.StatusRequestToDB("begin_transaction", "error")
		log.Errorf("Failed to begin transaction for InsertBids: %v", err)
		return err
	}

	metrics.StatusRequestToDB("begin_transaction", "success")

	err = insertOrder(ctx, tx, dept.Bids, dept.Timestamp, "bids")

	if err != nil {
		_ = tx.Rollback()
		metrics.StatusRequestToDB("insert_order", "error")
		log.Errorf("Failed to insert bids data: %v", err)
		return err
	}
	metrics.StatusRequestToDB("insert_order", "success")

	if err := tx.Commit(); err != nil {
		metrics.StatusRequestToDB("commit_transaction", "error")
		log.Errorf("Failed to commit transaction for InsertBids: %v", err)
		return err
	}
	metrics.StatusRequestToDB("commit_transaction", "succses")
	log.Info("InsertBids transaction committed successfully")
	return nil
}

func insertOrder(ctx context.Context, tx *sql.Tx, order entity.Order, timestamp int64, typeOrder string) error {
	query := `INSERT INTO history (type_price, price, volume, amount, time_stamp_order, transcription_type) 
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.ExecContext(ctx, query, order.Type, order.Price, order.Volume,
		order.Amount, timestamp, typeOrder)

	if err != nil {
		log.Errorf("failed to insert order data: %v", err)
		return err
	}
	log.Infof("Successfully inserted order data for type=%s", typeOrder)
	return nil
}
