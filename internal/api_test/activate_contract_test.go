package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/stretchr/testify/require"
)

type activateContractTestResponse struct {
	Data   *gen.ActivateContractResponse `json:"data"`
	Errors []string                      `json:"errors"`
}

func sendActivateContractRequest(
	t *testing.T,
	expectedStatus int,
	payload gen.ActivateContractRequest,
) activateContractTestResponse {
	t.Helper()

	return sendJSONRequest[activateContractTestResponse](
		t,
		http.MethodPost,
		contractTestURL+"activate",
		payload,
		expectedStatus,
	)
}

func TestServer_ActivateContract(t *testing.T) {
	t.Parallel()

	contract := createDraftContract(t, uuid.New())

	response := sendActivateContractRequest(t, fiber.StatusCreated, gen.ActivateContractRequest{
		ID: contract.ID,
	})

	require.Equal(t, []string{}, response.Errors)
	require.NotNil(t, response.Data)
	require.Equal(t, contract.ID, response.Data.ID)

	var savedStatus string
	err := testPool.QueryRow(
		context.Background(),
		`SELECT status FROM contracts WHERE id = $1`,
		contract.ID,
	).Scan(&savedStatus)
	require.NoError(t, err)
	require.Equal(t, "active", savedStatus)
}

func TestServer_ActivateContract_NotFound(t *testing.T) {
	t.Parallel()

	response := sendActivateContractRequest(t, fiber.StatusNotFound, gen.ActivateContractRequest{
		ID: uuid.New(),
	})

	require.Nil(t, response.Data)
	require.Equal(t, []string{domain.ErrNotFound.Error()}, response.Errors)
}

func TestServer_ActivateContract_InvalidStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status string
	}{
		{name: "уже активен", status: "active"},
		{name: "приостановлен", status: "suspended"},
		{name: "расторгнут", status: "terminated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			contract := createDraftContract(t, uuid.New())
			setContractStatus(t, contract.ID, tt.status)

			response := sendActivateContractRequest(t, fiber.StatusConflict, gen.ActivateContractRequest{
				ID: contract.ID,
			})

			require.Nil(t, response.Data)
			require.Contains(t, strings.Join(response.Errors, "; "), domain.ErrInvalidStatus.Error())

			var savedStatus string
			err := testPool.QueryRow(
				context.Background(),
				`SELECT status FROM contracts WHERE id = $1`,
				contract.ID,
			).Scan(&savedStatus)
			require.NoError(t, err)
			require.Equal(t, tt.status, savedStatus, "отказ не должен менять статус договора")
		})
	}
}

func TestServer_ActivateContract_AnotherActiveContract(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	activeContract := createDraftContract(t, companyID)
	activated := sendActivateContractRequest(t, fiber.StatusCreated, gen.ActivateContractRequest{
		ID: activeContract.ID,
	})
	require.Empty(t, activated.Errors)

	draft := createDraftContract(t, companyID)

	response := sendActivateContractRequest(t, fiber.StatusConflict, gen.ActivateContractRequest{
		ID: draft.ID,
	})

	require.Nil(t, response.Data)
	require.Equal(t, []string{domain.ErrActiveContractExists.Error()}, response.Errors)

	var savedStatus string
	err := testPool.QueryRow(
		context.Background(),
		`SELECT status FROM contracts WHERE id = $1`,
		draft.ID,
	).Scan(&savedStatus)
	require.NoError(t, err)
	require.Equal(t, "draft", savedStatus)
}

func TestServer_ActivateContract_Validation(t *testing.T) {
	t.Parallel()

	response := sendActivateContractRequest(t, fiber.StatusBadRequest, gen.ActivateContractRequest{})

	require.Nil(t, response.Data)
	require.Equal(t, []string{"id обязателен"}, response.Errors)
}

func TestServer_ActivateContract_ConcurrentContracts(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	first := createDraftContract(t, companyID)
	second := createDraftContract(t, companyID)

	statuses := sendConcurrentActivationRequests(t, []uuid.UUID{first.ID, second.ID})

	require.ElementsMatch(t, []int{fiber.StatusCreated, fiber.StatusConflict}, statuses)

	var activeCount int
	err := testPool.QueryRow(
		context.Background(),
		`SELECT count(*) FROM contracts WHERE company_id = $1 AND status = 'active'`,
		companyID,
	).Scan(&activeCount)
	require.NoError(t, err)
	require.Equal(t, 1, activeCount, "у компании должен быть только один активный договор")
}

func TestServer_ActivateContract_ConcurrentSameContract(t *testing.T) {
	t.Parallel()

	contract := createDraftContract(t, uuid.New())

	statuses := sendConcurrentActivationRequests(t, []uuid.UUID{contract.ID, contract.ID})

	require.ElementsMatch(t, []int{fiber.StatusCreated, fiber.StatusConflict}, statuses)
}

func sendConcurrentActivationRequests(t *testing.T, contractIDs []uuid.UUID) []int {
	t.Helper()

	// Ошибки проверяются в основном потоке теста, после выполнения запросов.
	type requestResult struct {
		status int
		err    error
	}

	results := make(chan requestResult, len(contractIDs))
	start := make(chan struct{})

	for _, contractID := range contractIDs {
		payload := gen.ActivateContractRequest{ID: contractID}
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, contractTestURL+"activate", bytes.NewReader(body))
		require.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		go func() {
			<-start
			response, err := testApp.Test(request, 10000)
			if err != nil {
				results <- requestResult{err: err}
				return
			}

			err = response.Body.Close()
			results <- requestResult{status: response.StatusCode, err: err}
		}()
	}

	close(start)

	statuses := make([]int, 0, len(contractIDs))
	for range contractIDs {
		result := <-results
		require.NoError(t, result.err)
		statuses = append(statuses, result.status)
	}

	return statuses
}
