package domain

import "github.com/ibobrov/share-trip-contract/internal/repository"

const (
	StatusDraft        = "draft"
	StatusActive       = "active"
	ContractNotFound   = "CONTRACT_NOT_FOUND"
	NoActiveContract   = "NO_ACTIVE_CONTRACT"
	ContractNotStarted = "CONTRACT_NOT_STARTED"
	ContractExpired    = "CONTRACT_EXPIRED"
	ServiceNotAllowed  = "SERVICE_NOT_ALLOWED"
)

type ContractUseCase struct {
	contractRepo *repository.ContractRepository
}

func NewContractUseCase(repo *repository.ContractRepository) *ContractUseCase {
	return &ContractUseCase{contractRepo: repo}
}
