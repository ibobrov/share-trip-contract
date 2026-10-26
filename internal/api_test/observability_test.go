package api_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/api"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/ibobrov/share-trip-contract/internal/observability/middleware"
	"github.com/ibobrov/share-trip-contract/internal/repository"
	"github.com/ibobrov/share-trip-contract/internal/service"
	"github.com/stretchr/testify/require"
)

func TestServer_CheckOption_LogsCheckPermissionAtInfo(t *testing.T) {
	t.Parallel()
	useCase := domain.NewContractUseCase(repository.NewContractRepository())
	contractService := service.NewContractService(testPool, useCase, nil)
	server := api.NewServer(contractService)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil)).With("service", "contract")
	app := fiber.New()
	app.Use(middleware.Correlation(logger))
	app.Post("/api/option/check", server.CheckOption)
	companyID := "company-" + uuid.NewString()
	tripID, clientID := uuid.NewString(), uuid.NewString()
	request := httptest.NewRequest(http.MethodPost, "/api/option/check", strings.NewReader(`{"company_id":"`+companyID+`","option_code":"trip_publication"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "request-123")
	request.Header.Set("X-Correlation-ID", "publication-456")
	request.Header.Set("X-Trip-ID", tripID)
	request.Header.Set("X-User-ID", clientID)

	response, err := app.Test(request)

	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, 1, strings.Count(logs.String(), "\n"), "check should have one final Info log")
	var entry map[string]any
	require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
	require.Equal(t, "INFO", entry["level"])
	require.Equal(t, "contract", entry["service"])
	require.Equal(t, "CheckPermission", entry["operation"])
	require.Equal(t, "denied", entry["result"])
	require.Equal(t, "request-123", entry["request_id"])
	require.Equal(t, "publication-456", entry["correlation_id"])
	require.Equal(t, tripID, entry["trip_id"])
	require.Equal(t, clientID, entry["client_id"])
}
