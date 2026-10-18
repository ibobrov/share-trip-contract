package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/jackc/pgx/v5"
)

type GetContractByIdRequest struct {
	ID uuid.UUID
}

type GetContractByIdResponse struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Status    string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}

func (s *ContractService) GetContractById(
	ctx context.Context,
	req GetContractByIdRequest,
) (GetContractByIdResponse, error) {
	result, err := tx(ctx, s.pool, func(tx pgx.Tx) (*GetContractByIdResponse, error) {
		response, err := s.contractUseCase.GetContractById(ctx, tx, domain.GetContractByIdRequest{
			ID: req.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("usecase.GetContractById: %w", err)
		}

		return &GetContractByIdResponse{
			ID:        response.ID,
			CompanyID: response.CompanyID,
			Status:    response.Status,
			StartsAt:  response.StartsAt,
			ExpiresAt: response.ExpiresAt,
			Options:   response.Options,
		}, nil
	})
	if err != nil {
		return GetContractByIdResponse{}, err
	}

	return *result, nil
}
