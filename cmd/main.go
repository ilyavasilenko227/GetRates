package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"rates/cmd/config"
	"rates/internal/controller"
	"rates/internal/infrastructure/metrics"
	"rates/internal/infrastructure/optel.go"
	"rates/internal/infrastructure/server"
	"rates/internal/repository"
	"rates/internal/service"
	"rates/pkg/logger"
	"sync"
	"syscall"
	"time"

	"github.com/pressly/goose"
	"go.uber.org/zap"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
	log *zap.SugaredLogger
)

func main() {
	var wg sync.WaitGroup

	configs, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	logger.BuildLogger(configs.LogLevel)
	log = logger.Logger().Named("main").Sugar()

	ctx := context.Background()

	db, err := repository.NewPostgresClient(configs.DbHost, configs.DbPort, configs.DbUser,
		configs.DbPassword, configs.DbName)
	if err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("error db migrate: %s", err)
	}

	repo := repository.NewRepository(db)
	service := service.NewService(repo)
	contrll := controller.NewController(service)
	server := server.NewServer(contrll)

	grpcServer := server.RunApp(configs.AppHost, configs.AppPort)

	// Проверка здоровья gRPC сервера
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("GetRatesUSDT", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		err := metrics.Listen(fmt.Sprintf("%s:%s", configs.PrometheusHost, configs.PrometheusPort))
		if err != nil {
			log.Error("error listen metrics: %w", err)
		}
	}()

	otelShutdown, err := optel.SetUpOTelSDK(ctx)
	if err != nil {
		log.Fatalf("error start tracer: %s", err)
	}

	defer func() {
		err := errors.Join(err, otelShutdown(ctx))
		log.Error("error stop service: %w", err)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-sigs
		log.Info("Received termination signal, shutting down...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		end := make(chan struct{})

		go func() {
			grpcServer.GracefulStop()
			end <- struct{}{}
			close(end)
		}()

		select {
		case <-end:
			log.Infof("Shutting down gracefully...")
			return
		case <-shutdownCtx.Done():
			grpcServer.Stop()
			log.Infof("Server stop whith time limit")
			return
		}
	}()

	wg.Wait()

}
