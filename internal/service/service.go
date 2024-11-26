package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"rates/internal/entity"
	"rates/internal/infrastructure/metrics"
	"rates/internal/repository"
	"rates/pkg/logger"
	"time"

	"go.opentelemetry.io/otel"
)

var (
	log = logger.Logger().Named("service").Sugar()
)

type Service struct {
	rep repository.Repositer
}

func NewService(rep repository.Repositer) *Service {
	return &Service{rep: rep}
}

func (s Service) GetRates(ctx context.Context) (entity.Depth, error) {
	log.Debug("Starting GetRates request")
	// Создание трассера для ослеживания времени получения данных от сервиса
	tracer := otel.Tracer("service.GetRacer")
	ctx, span := tracer.Start(ctx, "Service")
	defer span.End()

	// Метрика начала запроса к Garantex
	startTotal := time.Now()

	var dept entity.Depth
	// Get запрос к Garantex для получения свеч пары USDT-RUB
	resp, err := http.Get("https://garantex.org/api/v2/depth?market=usdtrub")
	log.Info("call a resp")
	if err != nil {
		// Метрика Prometheus неудачных запросов к Garantex
		metrics.StatusRequestToGarantex("error")
		log.Errorf("Error during HTTP request: %v", err)
		return entity.Depth{}, err
	}
	// Фиксация времени запроса к Garantex
	metrics.TimeRequestToGarantex("http_request", time.Since(startTotal).Seconds())
	//  Метрика Prometheus удачных запросов к Garantex
	metrics.StatusRequestToGarantex("success")

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body: %v", err)
		return entity.Depth{}, err
	}

	var data entity.DepthRequest
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("Error unmarshalling response data: %v", err)
		return entity.Depth{}, err
	}
	if len(data.Asks) > 0 && len(data.Bids) > 0 && data.Timestamp != 0 {
		dept = entity.Depth{
			Asks: entity.Order{
				Price:  data.Asks[0].Price,
				Volume: data.Asks[0].Volume,
				Amount: data.Asks[0].Amount,
				Factor: data.Asks[0].Factor,
				Type:   data.Asks[0].Type,
			},
			Bids: entity.Order{
				Price:  data.Bids[0].Price,
				Volume: data.Bids[0].Volume,
				Amount: data.Bids[0].Amount,
				Factor: data.Bids[0].Factor,
				Type:   data.Bids[0].Type,
			},
			Timestamp: data.Timestamp,
		}
	}
	// Метрика начала выполненеия запросов к репозиторию
	startTotalDB := time.Now()
	err = s.rep.InsertAsks(ctx, dept)
	if err != nil {
		return entity.Depth{}, err
	}

	err = s.rep.InsertBids(ctx, dept)
	if err != nil {
		return entity.Depth{}, err
	}
	// Фиксация времени запроса к репозиторию
	metrics.TimeRequestToDB("insert_to_db", time.Since(startTotalDB).Seconds())
	return dept, nil
}
