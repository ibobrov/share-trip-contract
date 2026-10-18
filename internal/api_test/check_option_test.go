package api_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/stretchr/testify/require"
)

type checkOptionTestResponse struct {
	Data   *gen.CheckOptionResponse `json:"data"`
	Errors []string                 `json:"errors"`
}

func sendCheckOptionRequest(
	t *testing.T,
	expectedStatus int,
	payload gen.CheckOptionRequest,
) checkOptionTestResponse {
	t.Helper()

	return sendJSONRequest[checkOptionTestResponse](
		t,
		http.MethodPost,
		optionTestURL+"check",
		payload,
		expectedStatus,
	)
}

func TestServer_CheckOption_Allowed(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	contract := createDraftContract(t, companyID)
	activated := sendActivateContractRequest(t, fiber.StatusCreated, gen.ActivateContractRequest{
		ID: contract.ID,
	})
	require.Empty(t, activated.Errors)

	response := sendCheckOptionRequest(t, fiber.StatusOK, gen.CheckOptionRequest{
		CompanyID:  companyID,
		OptionCode: gen.TripCreation,
	})

	require.Equal(t, []string{}, response.Errors)
	require.NotNil(t, response.Data)
	require.True(t, response.Data.Allowed)
	require.Nil(t, response.Data.Reason)
}

func TestServer_CheckOption_ContractNotFound(t *testing.T) {
	t.Parallel()

	response := sendCheckOptionRequest(t, fiber.StatusOK, gen.CheckOptionRequest{
		CompanyID:  uuid.New(),
		OptionCode: gen.TripCreation,
	})

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.False(t, response.Data.Allowed)
	require.NotNil(t, response.Data.Reason)
	require.Equal(t, gen.CONTRACTNOTFOUND, *response.Data.Reason)
}

func TestServer_CheckOption_NoActiveContract(t *testing.T) {
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

			response := sendCheckOptionRequest(t, fiber.StatusOK, gen.CheckOptionRequest{
				CompanyID:  companyID,
				OptionCode: gen.TripCreation,
			})

			require.Empty(t, response.Errors)
			require.NotNil(t, response.Data)
			require.False(t, response.Data.Allowed)
			require.NotNil(t, response.Data.Reason)
			require.Equal(t, gen.NOACTIVECONTRACT, *response.Data.Reason)
		})
	}
}

func TestServer_CheckOption_NotStarted(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	created := sendCreateContractRequest(t, fiber.StatusCreated, gen.CreateContractRequest{
		CompanyID: companyID,
		StartsAt:  time.Now().UTC().Add(time.Hour),
		Options:   gen.ContractOptions{{Code: gen.TripCreation}},
	})
	require.NotNil(t, created.Data)
	setContractStatus(t, created.Data.ID, "active")

	response := sendCheckOptionRequest(t, fiber.StatusOK, gen.CheckOptionRequest{
		CompanyID:  companyID,
		OptionCode: gen.TripCreation,
	})

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.False(t, response.Data.Allowed)
	require.NotNil(t, response.Data.Reason)
	require.Equal(t, gen.CONTRACTNOTSTARTED, *response.Data.Reason)
}

func TestServer_CheckOption_Expired(t *testing.T) {
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

	response := sendCheckOptionRequest(t, fiber.StatusOK, gen.CheckOptionRequest{
		CompanyID:  companyID,
		OptionCode: gen.TripCreation,
	})

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.False(t, response.Data.Allowed)
	require.NotNil(t, response.Data.Reason)
	require.Equal(t, gen.CONTRACTEXPIRED, *response.Data.Reason)
}

func TestServer_CheckOption_MissingOption(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	contract := createDraftContract(t, companyID)
	setContractStatus(t, contract.ID, "active")

	response := sendCheckOptionRequest(t, fiber.StatusOK, gen.CheckOptionRequest{
		CompanyID:  companyID,
		OptionCode: gen.Notifications,
	})

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.False(t, response.Data.Allowed)
	require.NotNil(t, response.Data.Reason)
	require.Equal(t, gen.SERVICENOTALLOWED, *response.Data.Reason)
}

func TestServer_CheckOption_DisabledOption(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	contract := createDraftContract(t, companyID)
	setContractStatus(t, contract.ID, "active")

	_, err := testPool.Exec(
		context.Background(),
		`UPDATE contract_options SET enabled = false WHERE contract_id = $1 AND code = $2`,
		contract.ID,
		string(gen.TripCreation),
	)
	require.NoError(t, err)

	response := sendCheckOptionRequest(t, fiber.StatusOK, gen.CheckOptionRequest{
		CompanyID:  companyID,
		OptionCode: gen.TripCreation,
	})

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.False(t, response.Data.Allowed)
	require.NotNil(t, response.Data.Reason)
	require.Equal(t, gen.SERVICENOTALLOWED, *response.Data.Reason)
}

func TestServer_CheckOption_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		change        func(*gen.CheckOptionRequest)
		expectedError string
	}{
		{
			name: "компания не указана",
			change: func(request *gen.CheckOptionRequest) {
				request.CompanyID = uuid.Nil
			},
			expectedError: "company_id обязателен",
		},
		{
			name: "услуга не указана",
			change: func(request *gen.CheckOptionRequest) {
				request.OptionCode = ""
			},
			expectedError: "option_code обязателен",
		},
		{
			name: "услуга неизвестна",
			change: func(request *gen.CheckOptionRequest) {
				request.OptionCode = "unknown"
			},
			expectedError: "UNKNOWN_SERVICE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := gen.CheckOptionRequest{
				CompanyID:  uuid.New(),
				OptionCode: gen.TripCreation,
			}
			tt.change(&request)

			response := sendCheckOptionRequest(t, fiber.StatusBadRequest, request)

			require.Nil(t, response.Data)
			require.Equal(t, []string{tt.expectedError}, response.Errors)
		})
	}
}
