package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/jackc/pgx/v5"
)

type GetActiveContractRequest struct {
	CompanyID string
}

type GetActiveContractResponse struct {
	ID        uuid.UUID
	CompanyID string
	Status    string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}

func (s *ContractService) GetActiveContract(
	ctx context.Context,
	req GetActiveContractRequest,
) (GetActiveContractResponse, error) {
	result, err := tx(ctx, s.pool, func(tx pgx.Tx) (*GetActiveContractResponse, error) {
		response, err := s.contractUseCase.GetActiveContract(ctx, tx, domain.GetActiveContractRequest{
			CompanyID: req.CompanyID,
		})
		if err != nil {
			return nil, fmt.Errorf("usecase.GetActiveContract: %w", err)
		}

		return &GetActiveContractResponse{
			ID:        response.ID,
			CompanyID: response.CompanyID,
			Status:    response.Status,
			StartsAt:  response.StartsAt,
			ExpiresAt: response.ExpiresAt,
			Options:   response.Options,
		}, nil
	})
	if err != nil {
		return GetActiveContractResponse{}, err
	}

	return *result, nil
}
