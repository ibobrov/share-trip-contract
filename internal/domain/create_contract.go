package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/internal/repository/entity"
	"github.com/jackc/pgx/v5"
)

type CreateContractRequest struct {
	CompanyID uuid.UUID
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}
type CreateContractResponse struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Status    string
	StartsAt  time.Time
	ExpiresAt *time.Time
	Options   []string
}

func (u *ContractUseCase) CreateContract(ctx context.Context, tx pgx.Tx, req CreateContractRequest) (*CreateContractResponse, error) {
	if err := validateContractTerms(req); err != nil {
		return nil, err
	}
	options := make([]entity.ContractOption, 0, len(req.Options))
	for _, code := range req.Options {
		options = append(options, entity.ContractOption{Code: code, Enabled: true})
	}
	contract := entity.Contract{
		ID:        uuid.New(),
		CompanyID: req.CompanyID,
		Status:    StatusDraft,
		StartsAt:  req.StartsAt.UTC(),
		ExpiresAt: req.ExpiresAt,
		Options:   options,
	}
	if contract.ExpiresAt != nil {
		t := contract.ExpiresAt.UTC()
		contract.ExpiresAt = &t
	}
	if err := u.contractRepo.Create(ctx, tx, contract); err != nil {
		return nil, fmt.Errorf("create contract: %w", err)
	}
	responseOptions := make([]string, 0, len(contract.Options))
	for _, option := range contract.Options {
		responseOptions = append(responseOptions, option.Code)
	}

	return &CreateContractResponse{
		ID:        contract.ID,
		CompanyID: contract.CompanyID,
		Status:    contract.Status,
		StartsAt:  contract.StartsAt,
		ExpiresAt: contract.ExpiresAt,
		Options:   responseOptions,
	}, nil
}

func validateContractTerms(req CreateContractRequest) error {
	if req.ExpiresAt != nil && !req.ExpiresAt.After(req.StartsAt) {
		return &BusinessValidationError{Message: "expires_at должен быть позже starts_at"}
	}

	seen := make(map[string]bool, len(req.Options))
	for _, code := range req.Options {
		if seen[code] {
			return &BusinessValidationError{Message: fmt.Sprintf("услуга %q повторяется", code)}
		}
		seen[code] = true
	}
	return nil
}
