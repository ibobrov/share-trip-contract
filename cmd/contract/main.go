package main

import (
	"context"
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
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}
	logger, logFile, err := app.NewLogger()
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Printf("close log file: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dbConfig := repository.Config{
		Host: config.Env("DB_HOST", "localhost"), Port: config.EnvInt("POSTGRES_PORT", 6545),
		User: config.Env("POSTGRES_USER", "postgres"), Password: config.Env("POSTGRES_PASSWORD", "password"),
		DBName: config.Env("POSTGRES_DB", "sharetrip-contract"), SSLMode: config.Env("DB_SSLMODE", "disable"),
	}
	pool, err := repository.NewPool(ctx, dbConfig.DSN())
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
	server.Route(application.Group(""))
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		if err := application.ShutdownWithTimeout(5 * time.Second); err != nil {
			logger.Error("shutdown HTTP server", "error", err)
		}
	}()
	err = application.Listen(":" + config.Env("HTTP_PORT", "4002"))
	stop()
	<-shutdownDone
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}
