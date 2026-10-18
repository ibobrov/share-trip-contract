package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/stretchr/testify/require"
)

type createContractTestResponse struct {
	Data   *gen.CreateContractResponse `json:"data"`
	Errors []string                    `json:"errors"`
}

func sendCreateContractRequest(
	t *testing.T,
	expectedStatus int,
	payload gen.CreateContractRequest,
) createContractTestResponse {
	t.Helper()

	return sendJSONRequest[createContractTestResponse](
		t,
		http.MethodPost,
		contractTestURL+"create",
		payload,
		expectedStatus,
	)
}

func TestServer_CreateContract(t *testing.T) {
	t.Parallel()

	companyID := uuid.New()
	startsAt := time.Now().UTC().Truncate(time.Second)
	expiresAt := startsAt.AddDate(1, 0, 0)
	options := gen.ContractOptions{
		{Code: gen.TripCreation},
		{Code: gen.Notifications},
	}

	response := sendCreateContractRequest(t, fiber.StatusCreated, gen.CreateContractRequest{
		CompanyID: companyID,
		StartsAt:  startsAt,
		ExpiresAt: &expiresAt,
		Options:   options,
	})

	require.Equal(t, []string{}, response.Errors)
	require.NotNil(t, response.Data)
	require.NotEqual(t, uuid.Nil, response.Data.ID)
	require.Equal(t, companyID, response.Data.CompanyID)
	require.Equal(t, "draft", string(response.Data.Status))
	require.Equal(t, startsAt, response.Data.StartsAt)
	require.NotNil(t, response.Data.ExpiresAt)
	require.Equal(t, expiresAt, *response.Data.ExpiresAt)
	require.ElementsMatch(t, options, response.Data.Options)

	// Проверяем сохранённый договор и включённые услуги.
	var savedCompanyID uuid.UUID
	var savedStatus string
	var savedStartsAt time.Time
	var savedExpiresAt *time.Time

	err := testPool.QueryRow(
		context.Background(),
		`SELECT company_id, status, starts_at, expires_at FROM contracts WHERE id = $1`,
		response.Data.ID,
	).Scan(&savedCompanyID, &savedStatus, &savedStartsAt, &savedExpiresAt)
	require.NoError(t, err)
	require.Equal(t, companyID, savedCompanyID)
	require.Equal(t, "draft", savedStatus)
	require.True(t, startsAt.Equal(savedStartsAt))
	require.NotNil(t, savedExpiresAt)
	require.True(t, expiresAt.Equal(*savedExpiresAt))

	var optionCount int
	var allOptionsEnabled bool

	err = testPool.QueryRow(
		context.Background(),
		`SELECT count(*), bool_and(enabled) FROM contract_options WHERE contract_id = $1`,
		response.Data.ID,
	).Scan(&optionCount, &allOptionsEnabled)
	require.NoError(t, err)
	require.Equal(t, len(options), optionCount)
	require.True(t, allOptionsEnabled)
}

func TestServer_CreateContract_Indefinite(t *testing.T) {
	t.Parallel()

	response := sendCreateContractRequest(t, fiber.StatusCreated, gen.CreateContractRequest{
		CompanyID: uuid.New(),
		StartsAt:  time.Now().UTC(),
		Options: gen.ContractOptions{
			{Code: gen.TripCreation},
		},
	})

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.Nil(t, response.Data.ExpiresAt)

	var savedExpiresAt *time.Time
	err := testPool.QueryRow(
		context.Background(),
		`SELECT expires_at FROM contracts WHERE id = $1`,
		response.Data.ID,
	).Scan(&savedExpiresAt)
	require.NoError(t, err)
	require.Nil(t, savedExpiresAt)
}

func TestServer_CreateContract_TransportValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		change        func(*gen.CreateContractRequest)
		errorContains string
	}{
		{
			name: "компания не указана",
			change: func(request *gen.CreateContractRequest) {
				request.CompanyID = uuid.Nil
			},
			errorContains: "company_id обязателен",
		},
		{
			name: "начало действия не указано",
			change: func(request *gen.CreateContractRequest) {
				request.StartsAt = time.Time{}
			},
			errorContains: "starts_at обязателен",
		},
		{
			name: "список услуг отсутствует",
			change: func(request *gen.CreateContractRequest) {
				request.Options = nil
			},
			errorContains: "список options не должен быть пустым",
		},
		{
			name: "список услуг пуст",
			change: func(request *gen.CreateContractRequest) {
				request.Options = gen.ContractOptions{}
			},
			errorContains: "список options не должен быть пустым",
		},
		{
			name: "неизвестная услуга",
			change: func(request *gen.CreateContractRequest) {
				request.Options = gen.ContractOptions{{Code: "unknown"}}
			},
			errorContains: "неизвестная услуга",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := gen.CreateContractRequest{
				CompanyID: uuid.New(),
				StartsAt:  time.Now().UTC(),
				Options:   gen.ContractOptions{{Code: gen.TripCreation}},
			}
			tt.change(&request)

			response := sendCreateContractRequest(t, fiber.StatusBadRequest, request)

			require.Nil(t, response.Data)
			require.NotEmpty(t, response.Errors)
			require.Contains(t, strings.Join(response.Errors, "; "), tt.errorContains)

			var contractCount int
			err := testPool.QueryRow(
				context.Background(),
				`SELECT count(*) FROM contracts WHERE company_id = $1`,
				request.CompanyID,
			).Scan(&contractCount)
			require.NoError(t, err)
			require.Zero(t, contractCount, "невалидный договор не должен сохраняться")
		})
	}
}

func TestServer_CreateContract_BusinessValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		change        func(*gen.CreateContractRequest)
		errorContains string
	}{
		{
			name: "окончание совпадает с началом",
			change: func(request *gen.CreateContractRequest) {
				request.ExpiresAt = &request.StartsAt
			},
			errorContains: "expires_at должен быть позже starts_at",
		},
		{
			name: "окончание раньше начала",
			change: func(request *gen.CreateContractRequest) {
				expiresAt := request.StartsAt.Add(-time.Hour)
				request.ExpiresAt = &expiresAt
			},
			errorContains: "expires_at должен быть позже starts_at",
		},
		{
			name: "услуга повторяется",
			change: func(request *gen.CreateContractRequest) {
				request.Options = gen.ContractOptions{
					{Code: gen.TripCreation},
					{Code: gen.TripCreation},
				}
			},
			errorContains: "повторяется",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := gen.CreateContractRequest{
				CompanyID: uuid.New(),
				StartsAt:  time.Now().UTC(),
				Options:   gen.ContractOptions{{Code: gen.TripCreation}},
			}
			tt.change(&request)

			response := sendCreateContractRequest(t, fiber.StatusBadRequest, request)

			require.Nil(t, response.Data)
			require.NotEmpty(t, response.Errors)
			require.Contains(t, strings.Join(response.Errors, "; "), tt.errorContains)

			var contractCount int
			err := testPool.QueryRow(
				context.Background(),
				`SELECT count(*) FROM contracts WHERE company_id = $1`,
				request.CompanyID,
			).Scan(&contractCount)
			require.NoError(t, err)
			require.Zero(t, contractCount, "невалидный договор не должен сохраняться")
		})
	}
}

func TestServer_CreateContract_InvalidJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body json.RawMessage
	}{
		{
			name: "UUID некорректен",
			body: json.RawMessage(`{"company_id":"not-a-uuid"}`),
		},
		{
			name: "дата некорректна",
			body: json.RawMessage(`{"starts_at":"завтра"}`),
		},
		{
			name: "неизвестное поле",
			body: json.RawMessage(`{"unknown_field":true}`),
		},
		{
			name: "массив вместо объекта",
			body: json.RawMessage(`[]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			response := sendJSONRequest[createContractTestResponse](
				t,
				http.MethodPost,
				contractTestURL+"create",
				tt.body,
				fiber.StatusBadRequest,
			)

			require.Nil(t, response.Data)
			require.Equal(t, []string{"Некорректное тело запроса."}, response.Errors)
		})
	}
}
