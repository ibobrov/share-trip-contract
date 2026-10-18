package service

import (
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContractService struct {
	pool            *pgxpool.Pool
	contractUseCase *domain.ContractUseCase
}

func NewContractService(pool *pgxpool.Pool, useCase *domain.ContractUseCase) *ContractService {
	return &ContractService{pool: pool, contractUseCase: useCase}
}
