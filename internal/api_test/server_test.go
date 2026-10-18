package api_test

import (
	"context"
	"database/sql"
	"io"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/internal/api"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/ibobrov/share-trip-contract/internal/observability/middleware"
	"github.com/ibobrov/share-trip-contract/internal/repository"
	"github.com/ibobrov/share-trip-contract/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const contractTestURL = "http://localhost/api/contract/"
const optionTestURL = "http://localhost/api/option/"

var testApp *fiber.App
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:16.6",
		postgres.WithDatabase("contract_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Printf("start test postgres: %v", err)
		return 1
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("terminate test postgres: %v", err)
		}
	}()
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Print(err)
		return 1
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Print(err)
		return 1
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Print(err)
		}
	}()
	if err := goose.SetDialect("postgres"); err != nil {
		log.Print(err)
		return 1
	}
	if err := goose.Up(db, "../../migrations"); err != nil {
		log.Print(err)
		return 1
	}
	// Проверяем применение, откат и повторное применение миграций.
	if err := goose.DownTo(db, "../../migrations", 0); err != nil {
		log.Print(err)
		return 1
	}
	if err := goose.Up(db, "../../migrations"); err != nil {
		log.Print(err)
		return 1
	}
	testPool, err = repository.NewPool(ctx, dsn)
	if err != nil {
		log.Print(err)
		return 1
	}
	defer testPool.Close()
	contractRepository := repository.NewContractRepository()
	contractUseCase := domain.NewContractUseCase(contractRepository)
	contractService := service.NewContractService(testPool, contractUseCase)
	server := api.NewServer(contractService)
	testApp = fiber.New(fiber.Config{ErrorHandler: api.ErrorHandler})
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	testApp.Use(middleware.Correlation(logger))
	server.Route(testApp.Group(""))
	return m.Run()
}
