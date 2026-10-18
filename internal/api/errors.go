package api

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/internal/domain"
	"github.com/ibobrov/share-trip-contract/internal/observability/logctx"
)

func HandleError(c *fiber.Ctx, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrBusinessValidation):
		var validationError *domain.BusinessValidationError
		if errors.As(err, &validationError) {
			return FailureWithHttpStatus(c, fiber.StatusBadRequest, validationError.Message)
		}
		return FailureWithHttpStatus(c, fiber.StatusBadRequest, domain.ErrBusinessValidation.Error())
	case errors.Is(err, domain.ErrNotFound):
		return FailureWithHttpStatus(c, fiber.StatusNotFound, domain.ErrNotFound.Error())
	case errors.Is(err, domain.ErrNoActiveContract):
		return FailureWithHttpStatus(c, fiber.StatusNotFound, domain.ErrNoActiveContract.Error())
	case errors.Is(err, domain.ErrInvalidStatus):
		return FailureWithHttpStatus(c, fiber.StatusConflict, domain.ErrInvalidStatus.Error())
	case errors.Is(err, domain.ErrActiveContractExists):
		return FailureWithHttpStatus(c, fiber.StatusConflict, domain.ErrActiveContractExists.Error())
	default:
		logctx.Logger(c.UserContext()).Error("request failed", slog.Any("error", err))
		return FailureWithHttpStatus(c, fiber.StatusInternalServerError, "Внутренняя ошибка сервиса.")
	}
}

// ErrorHandler keeps Fiber and generated-router errors inside the same envelope.
func ErrorHandler(c *fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		if fiberErr.Code >= 500 {
			return HandleError(c, err)
		}
		return FailureWithHttpStatus(c, fiberErr.Code, fiberErr.Message)
	}
	return HandleError(c, err)
}
