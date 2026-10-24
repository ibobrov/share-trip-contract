package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/internal/api"
	"github.com/ibobrov/share-trip-contract/internal/app"
	"github.com/ibobrov/share-trip-contract/internal/config"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/ibobrov/share-trip-contract/internal/observability/middleware"
	"github.com/ibobrov/share-trip-contract/internal/repository"
	"github.com/ibobrov/share-trip-contract/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: invalid or unreadable file")
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	logger, logFile, err := app.NewLogger(cfg.LogFile, cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() {
		if logFile != nil {
			if err := logFile.Close(); err != nil {
				log.Printf("close log file: %v", err)
			}
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := repository.NewPool(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()
	repo := repository.NewContractRepository()
	useCase := domain.NewContractUseCase(repo)
	contractService := service.NewContractService(pool, useCase)
	server := api.NewServer(contractService)
	application := fiber.New(fiber.Config{EnablePrintRoutes: true, ErrorHandler: api.ErrorHandler})
	application.Use(middleware.Correlation(logger))
	application.Use(swagger.New(swagger.Config{
		BasePath: "/", FilePath: "./api/contract.yaml", Path: "docs", Title: "ShareTrip Contract API",
	}))
	server.Route(application.Group(""), cfg.RequestTimeout)
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		if err := application.ShutdownWithTimeout(5 * time.Second); err != nil {
			logger.Error("shutdown HTTP server", "error", err)
		}
	}()
	err = application.Listen(":" + cfg.HTTPPort)
	stop()
	<-shutdownDone
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}
