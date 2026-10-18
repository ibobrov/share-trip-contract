package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CreateContractRequest struct {
	CompanyID string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}

type CreateContractResponse struct {
	ID        uuid.UUID
	CompanyID string
	Status    string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}

func (s *ContractService) CreateContract(
	ctx context.Context,
	req CreateContractRequest,
) (CreateContractResponse, error) {
	result, err := tx(ctx, s.pool, func(tx pgx.Tx) (*CreateContractResponse, error) {
		response, err := s.contractUseCase.CreateContract(ctx, tx, domain.CreateContractRequest{
			CompanyID: req.CompanyID,
			StartsAt:  req.StartsAt,
			ExpiresAt: req.ExpiresAt,
			Options:   req.Options,
		})
		if err != nil {
			return nil, fmt.Errorf("usecase.CreateContract: %w", err)
		}

		return &CreateContractResponse{
			ID:        response.ID,
			CompanyID: response.CompanyID,
			Status:    response.Status,
			StartsAt:  response.StartsAt,
			ExpiresAt: response.ExpiresAt,
			Options:   response.Options,
		}, nil
	})
	if err != nil {
		return CreateContractResponse{}, err
	}

	return *result, nil
}
