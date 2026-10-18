package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/jackc/pgx/v5"
)

type ActivateContractRequest struct {
	ID uuid.UUID
}

type ActivateContractResponse struct {
	ID uuid.UUID
}

func (s *ContractService) ActivateContract(
	ctx context.Context,
	req ActivateContractRequest,
) (ActivateContractResponse, error) {
	result, err := tx(ctx, s.pool, func(tx pgx.Tx) (*ActivateContractResponse, error) {
		response, err := s.contractUseCase.ActivateContract(ctx, tx, domain.ActivateContractRequest{
			ID: req.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("usecase.ActivateContract: %w", err)
		}

		return &ActivateContractResponse{
			ID: response.ID,
		}, nil
	})
	if err != nil {
		return ActivateContractResponse{}, err
	}

	return *result, nil
}
