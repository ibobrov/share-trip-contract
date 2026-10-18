package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/stretchr/testify/require"
)

func createDraftContract(t *testing.T, companyID string) *gen.CreateContractResponse {
	t.Helper()

	response := sendCreateContractRequest(t, fiber.StatusCreated, gen.CreateContractRequest{
		CompanyID: companyID,
		StartsAt:  time.Now().UTC().Add(-time.Hour).Truncate(time.Second),
		Options: gen.ContractOptions{
			{Code: gen.TripCreation},
		},
	})

	require.Empty(t, response.Errors)
	require.NotNil(t, response.Data)
	require.Equal(t, "draft", string(response.Data.Status))

	return response.Data
}

func setContractStatus(t *testing.T, contractID uuid.UUID, status string) {
	t.Helper()

	// Подготовка состояния, для которого пока нет отдельной ручки API.
	result, err := testPool.Exec(
		context.Background(),
		`UPDATE contracts SET status = $1 WHERE id = $2`,
		status,
		contractID,
	)
	require.NoError(t, err)
	require.EqualValues(t, 1, result.RowsAffected())
}
