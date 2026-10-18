package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ibobrov/share-trip-contract/internal/repository/entity"
	"github.com/jackc/pgx/v5"
)

type CheckOptionRequest struct {
	CompanyID  string
	OptionCode string
}
type CheckOptionResponse struct {
	Allowed bool
	Reason  *string
}

func (u *ContractUseCase) CheckOption(ctx context.Context, tx pgx.Tx, req CheckOptionRequest) (*CheckOptionResponse, error) {
	contract, err := u.contractRepo.Active(ctx, tx, req.CompanyID)
	if errors.Is(err, pgx.ErrNoRows) {
		exists, err := u.contractRepo.CompanyHasContracts(ctx, tx, req.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("check company contracts: %w", err)
		}
		reason := ContractNotFound
		if exists {
			reason = NoActiveContract
		}
		return &CheckOptionResponse{Reason: &reason}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get contract for option: %w", err)
	}
	now := time.Now().UTC()
	result := evaluateOption(contract, req.OptionCode, now)
	return &result, nil
}

func evaluateOption(contract entity.Contract, code string, now time.Time) CheckOptionResponse {
	reason := ""
	switch {
	case contract.Status != StatusActive:
		reason = NoActiveContract
	case now.Before(contract.StartsAt):
		reason = ContractNotStarted
	case contract.ExpiresAt != nil && !now.Before(*contract.ExpiresAt):
		reason = ContractExpired
	default:
		for _, option := range contract.Options {
			if option.Code == code && option.Enabled {
				return CheckOptionResponse{Allowed: true}
			}
		}
		reason = ServiceNotAllowed
	}
	return CheckOptionResponse{Reason: &reason}
}
