package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ActivateContractRequest struct {
	ID uuid.UUID
}
type ActivateContractResponse struct {
	ID uuid.UUID
}

func (u *ContractUseCase) ActivateContract(ctx context.Context, tx pgx.Tx, req ActivateContractRequest) (*ActivateContractResponse, error) {
	contract, err := u.contractRepo.Get(ctx, tx, req.ID, true)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get contract for activation: %w", err)
	}
	if contract.Status != StatusDraft {
		return nil, ErrInvalidStatus
	}
	if err := u.contractRepo.Activate(ctx, tx, req.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "contracts_one_active_per_company_idx" {
			return nil, ErrActiveContractExists
		}
		return nil, fmt.Errorf("activate contract: %w", err)
	}
	return &ActivateContractResponse{ID: req.ID}, nil
}
