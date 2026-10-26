package api

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/internal/domain"
)

func HandleError(c *fiber.Ctx, logger *slog.Logger, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrBusinessValidation):
		logger.Info("request rejected", "result", "validation_error")
		var validationError *domain.BusinessValidationError
		if errors.As(err, &validationError) {
			return FailureWithHttpStatus(c, fiber.StatusBadRequest, validationError.Message)
		}
		return FailureWithHttpStatus(c, fiber.StatusBadRequest, domain.ErrBusinessValidation.Error())
	case errors.Is(err, domain.ErrNotFound):
		logger.Info("request rejected", "result", "not_found")
		return FailureWithHttpStatus(c, fiber.StatusNotFound, domain.ErrNotFound.Error())
	case errors.Is(err, domain.ErrNoActiveContract):
		logger.Info("request rejected", "result", "not_found")
		return FailureWithHttpStatus(c, fiber.StatusNotFound, domain.ErrNoActiveContract.Error())
	case errors.Is(err, domain.ErrInvalidStatus):
		logger.Info("request rejected", "result", "conflict")
		return FailureWithHttpStatus(c, fiber.StatusConflict, domain.ErrInvalidStatus.Error())
	case errors.Is(err, domain.ErrActiveContractExists):
		logger.Info("request rejected", "result", "conflict")
		return FailureWithHttpStatus(c, fiber.StatusConflict, domain.ErrActiveContractExists.Error())
	default:
		logger.Error("request failed", "result", "internal_error", slog.Any("error", err))
		return FailureWithHttpStatus(c, fiber.StatusInternalServerError, "Внутренняя ошибка сервиса.")
	}
}

// ErrorHandler keeps Fiber and generated-router errors inside the same envelope.
func ErrorHandler(c *fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		if fiberErr.Code >= 500 {
			return HandleError(c, handlerLogger(c, "HTTP"), err)
		}
		return FailureWithHttpStatus(c, fiberErr.Code, fiberErr.Message)
	}
	return HandleError(c, handlerLogger(c, "HTTP"), err)
}
