package service

import (
	"context"
	"errors"
	"rates/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepositer struct {
	mock.Mock
}

func (m *MockRepositer) InsertAsks(ctx context.Context, dept entity.Depth) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *MockRepositer) InsertBids(ctx context.Context, dept entity.Depth) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func TestGetRates(t *testing.T) {
	// Мокаем репозиторий
	mockRepo := new(MockRepositer)

	// Создаем сервис
	service := NewService(mockRepo)

	// Ожидаем, что InsertAsks и InsertBids будут вызваны один раз
	mockRepo.On("InsertAsks", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("InsertBids", mock.Anything, mock.Anything).Return(nil)

	// Запускаем тест
	dept, err := service.GetRates(context.Background())

	// Проверяем что ошибок не было
	assert.NoError(t, err)
	assert.NotNil(t, dept)

	// Проверяем, что InsertAsks и InsertBids были вызваны
	mockRepo.AssertExpectations(t)
}

func TestGetRates_DBError(t *testing.T) {
	// Мокаем репозиторий
	mockRepo := new(MockRepositer)

	// Ожидаем, что InsertAsks вернет ошибку
	mockRepo.On("InsertAsks", mock.Anything, mock.Anything).Return(errors.New("DB error"))
	// Ожидаем, что InsertBids НЕ будет вызван, потому что ошибка произошла до его вызова
	mockRepo.On("InsertBids", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Создаем сервис
	service := NewService(mockRepo)

	// Запускаем тест
	_, err := service.GetRates(context.Background())

	// Проверяем, что ошибка из-за работы с базой данных
	assert.Error(t, err)

	// Проверяем, что InsertAsks был вызван с ошибкой, а InsertBids не был вызван
	mockRepo.AssertExpectations(t)
}
