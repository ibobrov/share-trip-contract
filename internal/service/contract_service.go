package service

import (
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/ibobrov/share-trip-contract/internal/observability/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type ContractService struct {
	metrics         *metrics.Metrics
	pool            *pgxpool.Pool
	contractUseCase *domain.ContractUseCase
}

func NewContractService(pool *pgxpool.Pool, useCase *domain.ContractUseCase, observer *metrics.Metrics) *ContractService {
	if observer == nil {
		observer = metrics.New(prometheus.NewRegistry())
	}
	return &ContractService{pool: pool, contractUseCase: useCase, metrics: observer}
}
