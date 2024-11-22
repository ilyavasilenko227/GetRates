package controller

import (
	"context"
	"rates/internal/entity"
	"rates/internal/infrastructure/metrics"
	pb "rates/internal/infrastructure/pb"
	"rates/pkg/logger"
)

var (
	log = logger.Logger().Named("controller").Sugar()
)

type Servicer interface {
	GetRates(ctx context.Context) (entity.Depth, error)
}

type Controller struct {
	pb.UnimplementedGetRateserServer
	service Servicer
}

func NewController(service Servicer) *Controller {
	return &Controller{service: service}
}

func (c Controller) GetRates(ctx context.Context, _ *pb.RatesRequest) (*pb.RatesResponse, error) {
	log.Infof("Received GetRates request")

	// метрика Prometheus общее количество запросов
	metrics.CountRequestToService()

	orders, err := c.service.GetRates(ctx)
	if err != nil {
		return &pb.RatesResponse{}, err
	}
	// Метрика Prometheus количества успешных ответов
	metrics.CountSuccessRequestToService()

	asResp := &pb.Order{
		Price:  orders.Asks.Price,
		Volume: orders.Asks.Volume,
		Amount: orders.Asks.Amount,
		Factor: orders.Asks.Factor,
		Type:   orders.Asks.Type,
	}
	bitResp := &pb.Order{
		Price:  orders.Bids.Price,
		Volume: orders.Bids.Volume,
		Amount: orders.Bids.Amount,
		Factor: orders.Bids.Factor,
		Type:   orders.Bids.Type,
	}
	depReq := &pb.RatesResponse{
		Ask:       asResp,
		Bid:       bitResp,
		Timestamp: orders.Timestamp,
	}

	log.Infof("Returning rates response with Ask Price: %s, Bid Price: %s",
		orders.Asks.Price, orders.Bids.Price)

	return depReq, nil
}
