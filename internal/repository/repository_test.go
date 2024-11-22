package repository

import (
	"context"
	"rates/internal/entity"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestInsertOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin() // Начало транзакции

	// Создаем mock транзакцию
	mockTx, err := db.Begin()
	require.NoError(t, err)

	// Параметры для вставки
	ctx := context.Background()
	order := entity.Order{
		Type:   "limit",
		Price:  "100.50",
		Volume: "2",
		Amount: "201.0",
	}
	timestamp := time.Now().Unix()
	typeOrder := "buy"

	// Ожидаем вызов SQL-запроса на вставку
	mock.ExpectExec(`INSERT INTO history \(type_price, price, volume, amount, time_stamp_order, transcription_type\)`).
		WithArgs(order.Type, order.Price, order.Volume, order.Amount, timestamp, typeOrder).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Успешный результат

	// Вызываем тестируемую функцию
	err = insertOrder(ctx, mockTx, order, timestamp, typeOrder)
	require.NoError(t, err)

	// Ожидаем завершения транзакции (Commit)
	mock.ExpectCommit()

	// Выполняем commit на mockTx
	err = mockTx.Commit()
	require.NoError(t, err)

	// Убеждаемся, что все ожидания выполнены
	require.NoError(t, mock.ExpectationsWereMet())
}
