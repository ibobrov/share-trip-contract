package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/ibobrov/share-trip-contract/internal/observability/logctx"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type CheckOptionRequest struct {
	CompanyID  string
	OptionCode string
}

type CheckOptionResponse struct {
	Allowed bool
	Reason  *string
}

func (s *ContractService) CheckOption(
	ctx context.Context,
	req CheckOptionRequest,
) (response CheckOptionResponse, err error) {
	ctx, span := otel.Tracer("contract").Start(ctx, "CheckPermission")
	defer span.End()
	started := time.Now()
	defer func() {
		result := "allowed"
		if err != nil {
			result = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, "operation failed")
		} else if !response.Allowed {
			result = "denied"
		}
		s.metrics.PermissionCheckTotal.WithLabelValues(result).Inc()
		s.metrics.PermissionCheckDuration.WithLabelValues(result).Observe(time.Since(started).Seconds())
		logctx.Logger(ctx).Debug("permission check completed", "component", "ContractService", "operation", "CheckPermission", "company_id", req.CompanyID, "option_code", req.OptionCode, "result", result)
	}()
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
