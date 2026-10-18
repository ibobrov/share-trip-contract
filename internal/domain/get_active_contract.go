package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type GetActiveContractRequest struct {
	CompanyID uuid.UUID
}
type GetActiveContractResponse struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Status    string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}

func (u *ContractUseCase) GetActiveContract(ctx context.Context, tx pgx.Tx, req GetActiveContractRequest) (*GetActiveContractResponse, error) {
	contract, err := u.contractRepo.Active(ctx, tx, req.CompanyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoActiveContract
	}
	if err != nil {
		return nil, fmt.Errorf("get active contract: %w", err)
	}
	now := time.Now().UTC()
	if contract.ExpiresAt != nil && !now.Before(*contract.ExpiresAt) {
		return nil, ErrNoActiveContract
	}
	responseOptions := make([]string, 0, len(contract.Options))
	for _, option := range contract.Options {
		responseOptions = append(responseOptions, option.Code)
	}

	return &GetActiveContractResponse{
		ID:        contract.ID,
		CompanyID: contract.CompanyID,
		Status:    contract.Status,
		StartsAt:  contract.StartsAt,
		ExpiresAt: contract.ExpiresAt,
		Options:   responseOptions,
	}, nil
}
