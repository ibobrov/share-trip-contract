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
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/stretchr/testify/require"
)

type getContractByIdTestResponse struct {
	Data   *gen.GetContractByIdResponse `json:"data"`
	Errors []string                     `json:"errors"`
}

func sendGetContractByIdRequest(
	t *testing.T,
	expectedStatus int,
	payload gen.GetContractByIdRequest,
) getContractByIdTestResponse {
	t.Helper()

	return sendJSONRequest[getContractByIdTestResponse](
		t,
		http.MethodPost,
		contractTestURL+"get-by-id",
		payload,
		expectedStatus,
	)
}

func TestServer_GetContractById(t *testing.T) {
	t.Parallel()

	startsAt := time.Now().UTC().Truncate(time.Second)
	expiresAt := startsAt.AddDate(1, 0, 0)
	created := sendCreateContractRequest(t, fiber.StatusCreated, gen.CreateContractRequest{
		CompanyID: "company-" + uuid.NewString(),
		StartsAt:  startsAt,
		ExpiresAt: &expiresAt,
		Options: gen.ContractOptions{
			{Code: gen.TripCreation},
			{Code: gen.TripParticipants},
		},
	})
	require.Empty(t, created.Errors)
	require.NotNil(t, created.Data)

	response := sendGetContractByIdRequest(t, fiber.StatusOK, gen.GetContractByIdRequest{
		ID: created.Data.ID,
	})

	require.Equal(t, []string{}, response.Errors)
	require.NotNil(t, response.Data)
	require.Equal(t, created.Data.ID, response.Data.ID)
	require.Equal(t, created.Data.CompanyID, response.Data.CompanyID)
	require.Equal(t, string(created.Data.Status), string(response.Data.Status))
	require.True(t, created.Data.StartsAt.Equal(response.Data.StartsAt))
	require.NotNil(t, response.Data.ExpiresAt)
	require.True(t, expiresAt.Equal(*response.Data.ExpiresAt))
	require.ElementsMatch(t, created.Data.Options, response.Data.Options)
}

func TestServer_GetContractById_NotFound(t *testing.T) {
	t.Parallel()

	response := sendGetContractByIdRequest(t, fiber.StatusNotFound, gen.GetContractByIdRequest{
		ID: uuid.New(),
	})

	require.Nil(t, response.Data)
	require.NotEmpty(t, response.Errors)
	require.Contains(t, strings.Join(response.Errors, "; "), domain.ErrNotFound.Error())
}

func TestServer_GetContractById_Validation(t *testing.T) {
	t.Parallel()

	response := sendGetContractByIdRequest(t, fiber.StatusBadRequest, gen.GetContractByIdRequest{})

	require.Nil(t, response.Data)
	require.Equal(t, []string{"id обязателен"}, response.Errors)
}

func TestServer_GetContractById_IndefiniteReturnsNull(t *testing.T) {
	t.Parallel()

	contract := createDraftContract(t, "company-"+uuid.NewString())

	// RawMessage отличает явный null от отсутствующего поля.
	type expirationTestResponse struct {
		Data *struct {
			ExpiresAt json.RawMessage `json:"expires_at"`
		} `json:"data"`
		Errors []string `json:"errors"`
	}

	response := sendJSONRequest[expirationTestResponse](
		t,
		http.MethodPost,
		contractTestURL+"get-by-id",
		gen.GetContractByIdRequest{ID: contract.ID},
		fiber.StatusOK,
	)

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.JSONEq(t, "null", string(response.Data.ExpiresAt))

	var savedExpiresAt *time.Time
	err := testPool.QueryRow(
		context.Background(),
		`SELECT expires_at FROM contracts WHERE id = $1`,
		contract.ID,
	).Scan(&savedExpiresAt)
	require.NoError(t, err)
	require.Nil(t, savedExpiresAt)
}
