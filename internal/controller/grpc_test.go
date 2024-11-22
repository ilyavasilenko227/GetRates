package controller_test

import (
	"context"
	"errors"
	"testing"

	"rates/internal/controller"
	"rates/internal/entity"
	pb "rates/internal/infrastructure/pb"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockServicer - мок для интерфейса Servicer
type MockServicer struct {
	mock.Mock
}

func (m *MockServicer) GetRates(ctx context.Context) (entity.Depth, error) {
	args := m.Called(ctx)
	return args.Get(0).(entity.Depth), args.Error(1)
}

func TestController_GetRates(t *testing.T) {
	// Создаем mock сервиса
	mockService := new(MockServicer)

	// Создаем контроллер с mock сервисом
	ctrl := controller.NewController(mockService)

	// Контекст вызова
	ctx := context.Background()

	// Подготовленные данные
	expectedDepth := entity.Depth{
		Asks: entity.Order{
			Price:  "100",
			Volume: "1",
			Amount: "100",
			Factor: "2",
			Type:   "ask",
		},
		Bids: entity.Order{
			Price:  "90",
			Volume: "1",
			Amount: "90",
			Factor: "2",
			Type:   "bid",
		},
		Timestamp: 1234567890,
	}

	// Настройка мока для успешного вызова
	mockService.On("GetRates", ctx).Return(expectedDepth, nil)

	// Вызов тестируемой функции
	req := &pb.RatesRequest{}
	resp, err := ctrl.GetRates(ctx, req)

	// Проверяем отсутствие ошибок
	require.NoError(t, err)

	// Проверяем, что результат корректно преобразован в ответ gRPC
	require.Equal(t, "100", resp.Ask.Price)
	require.Equal(t, "90", resp.Bid.Price)
	require.Equal(t, int64(1234567890), resp.Timestamp)

	// Убедимся, что метод сервиса был вызван один раз
	mockService.AssertExpectations(t)
}

func TestController_GetRates_Error(t *testing.T) {
	// Создаем mock сервиса
	mockService := new(MockServicer)

	// Создаем контроллер с mock сервисом
	ctrl := controller.NewController(mockService)

	// Контекст вызова
	ctx := context.Background()

	// Настройка мока для вызова с ошибкой
	mockService.On("GetRates", ctx).Return(entity.Depth{}, errors.New("service error"))

	// Вызов тестируемой функции
	req := &pb.RatesRequest{}
	resp, err := ctrl.GetRates(ctx, req)

	// Проверяем, что ошибка возвращена
	require.Error(t, err)
	require.Equal(t, &pb.RatesResponse{}, resp)

	// Убедимся, что метод сервиса был вызван один раз
	mockService.AssertExpectations(t)
}
