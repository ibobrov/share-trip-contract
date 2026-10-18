package api_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/stretchr/testify/require"
)

type getActiveContractTestResponse struct {
	Data   *gen.GetActiveContractResponse `json:"data"`
	Errors []string                       `json:"errors"`
}

func sendGetActiveContractRequest(
	t *testing.T,
	expectedStatus int,
	payload gen.GetActiveContractRequest,
) getActiveContractTestResponse {
	t.Helper()

	return sendJSONRequest[getActiveContractTestResponse](
		t,
		http.MethodPost,
		contractTestURL+"get-active",
		payload,
		expectedStatus,
	)
}

func TestServer_GetActiveContract(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	contract := createDraftContract(t, companyID)
	activated := sendActivateContractRequest(t, fiber.StatusCreated, gen.ActivateContractRequest{
		ID: contract.ID,
	})
	require.Empty(t, activated.Errors)

	response := sendGetActiveContractRequest(t, fiber.StatusOK, gen.GetActiveContractRequest{
		CompanyID: companyID,
	})

	require.Equal(t, []string{}, response.Errors)
	require.NotNil(t, response.Data)
	require.Equal(t, contract.ID, response.Data.ID)
	require.Equal(t, companyID, response.Data.CompanyID)
	require.Equal(t, "active", string(response.Data.Status))
	require.True(t, contract.StartsAt.Equal(response.Data.StartsAt))
	require.Nil(t, response.Data.ExpiresAt)
	require.ElementsMatch(t, contract.Options, response.Data.Options)
}

func TestServer_GetActiveContract_NotFound(t *testing.T) {
	t.Parallel()

	response := sendGetActiveContractRequest(t, fiber.StatusNotFound, gen.GetActiveContractRequest{
		CompanyID: uuid.New(),
	})

	require.Nil(t, response.Data)
	require.Equal(t, []string{domain.ErrNoActiveContract.Error()}, response.Errors)
}

func TestServer_GetActiveContract_InactiveStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status string
	}{
		{name: "черновик", status: "draft"},
		{name: "приостановлен", status: "suspended"},
		{name: "расторгнут", status: "terminated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			companyID := uuid.New()
			contract := createDraftContract(t, companyID)
			setContractStatus(t, contract.ID, tt.status)

			response := sendGetActiveContractRequest(t, fiber.StatusNotFound, gen.GetActiveContractRequest{
				CompanyID: companyID,
			})

			require.Nil(t, response.Data)
			require.Equal(t, []string{domain.ErrNoActiveContract.Error()}, response.Errors)
		})
	}
}

func TestServer_GetActiveContract_Expired(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	startsAt := time.Now().UTC().Add(-2 * time.Hour)
	expiresAt := startsAt.Add(time.Hour)
	created := sendCreateContractRequest(t, fiber.StatusCreated, gen.CreateContractRequest{
		CompanyID: companyID,
		StartsAt:  startsAt,
		ExpiresAt: &expiresAt,
		Options:   gen.ContractOptions{{Code: gen.TripCreation}},
	})
	require.NotNil(t, created.Data)
	setContractStatus(t, created.Data.ID, "active")

	response := sendGetActiveContractRequest(t, fiber.StatusNotFound, gen.GetActiveContractRequest{
		CompanyID: companyID,
	})

	require.Nil(t, response.Data)
	require.Contains(t, strings.Join(response.Errors, "; "), domain.ErrNoActiveContract.Error())
}

func TestServer_GetActiveContract_FutureStart(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	startsAt := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	created := sendCreateContractRequest(t, fiber.StatusCreated, gen.CreateContractRequest{
		CompanyID: companyID,
		StartsAt:  startsAt,
		Options:   gen.ContractOptions{{Code: gen.TripCreation}},
	})
	require.NotNil(t, created.Data)
	setContractStatus(t, created.Data.ID, "active")

	response := sendGetActiveContractRequest(t, fiber.StatusOK, gen.GetActiveContractRequest{
		CompanyID: companyID,
	})

	// Эта ручка выбирает active без проверки начала; доступность проверяет /option/check.
	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.Equal(t, created.Data.ID, response.Data.ID)
	require.True(t, startsAt.Equal(response.Data.StartsAt))
}

func TestServer_GetActiveContract_Validation(t *testing.T) {
	t.Parallel()

	response := sendGetActiveContractRequest(t, fiber.StatusBadRequest, gen.GetActiveContractRequest{})

	require.Nil(t, response.Data)
	require.Equal(t, []string{"company_id обязателен"}, response.Errors)
}
