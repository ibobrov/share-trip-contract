package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type GetContractByIdRequest struct {
	ID uuid.UUID
}
type GetContractByIdResponse struct {
	ID        uuid.UUID
	CompanyID string
	Status    string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}

func (u *ContractUseCase) GetContractById(ctx context.Context, tx pgx.Tx, req GetContractByIdRequest) (*GetContractByIdResponse, error) {
	contract, err := u.contractRepo.Get(ctx, tx, req.ID, false)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get contract: %w", err)
	}
	responseOptions := make([]string, 0, len(contract.Options))
	for _, option := range contract.Options {
		responseOptions = append(responseOptions, option.Code)
	}

	return &GetContractByIdResponse{
		ID:        contract.ID,
		CompanyID: contract.CompanyID,
		Status:    contract.Status,
		StartsAt:  contract.StartsAt,
		ExpiresAt: contract.ExpiresAt,
		Options:   responseOptions,
	}, nil
}
