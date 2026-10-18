package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CheckOptionRequest struct {
	CompanyID  uuid.UUID
	OptionCode string
}

type CheckOptionResponse struct {
	Allowed bool
	Reason  *string
}

func (s *ContractService) CheckOption(
	ctx context.Context,
	req CheckOptionRequest,
) (CheckOptionResponse, error) {
	result, err := tx(ctx, s.pool, func(tx pgx.Tx) (*CheckOptionResponse, error) {
		response, err := s.contractUseCase.CheckOption(ctx, tx, domain.CheckOptionRequest{
			CompanyID:  req.CompanyID,
			OptionCode: req.OptionCode,
		})
		if err != nil {
			return nil, fmt.Errorf("usecase.CheckOption: %w", err)
		}

		return &CheckOptionResponse{
			Allowed: response.Allowed,
			Reason:  response.Reason,
		}, nil
	})
	if err != nil {
		return CheckOptionResponse{}, err
	}

	return *result, nil
}
