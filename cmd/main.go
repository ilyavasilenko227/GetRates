package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
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

	ctx := context.Background()

	logger.BuildLogger("DEBUG")
	log = logger.Logger().Named("main").Sugar()

	db, err := repository.NewPostgresClient()
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
	appHost := os.Getenv("APP_HOST")
	appPort := os.Getenv("APP_PORT")
	if appHost == "" || appPort == "" {
		appHost = "localhost"
		appPort = "8080"
	}
	grpcServer := server.RunApp(appHost, appPort)

	// Проверка здоровья gRPC сервера
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("GetRatesUSDT", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		err := metrics.Listen(fmt.Sprintf("%s:%s",
			os.Getenv("PROMETHEUS_HOST"), os.Getenv("PROMETHEUS_PORT")))
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
