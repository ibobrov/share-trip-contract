package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"sort"

	"github.com/asaskevich/govalidator/v12"
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/observability/logctx"
	"github.com/ibobrov/share-trip-contract/internal/service"
)

type Server struct{ contractService *service.ContractService }

var _ gen.ServerInterface = (*Server)(nil)

func NewServer(contractService *service.ContractService) *Server {
	return &Server{contractService: contractService}
}

// Транспортная валидация: JSON и valid-теги, сгенерированные из OpenAPI.
// Бизнес-условия проверяются в domain.
func ParseAndValidateRequest[T any](request *T, c *fiber.Ctx, logger *slog.Logger) (bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(request)
	if err == nil {
		var extra any
		if trailing := decoder.Decode(&extra); !errors.Is(trailing, io.EOF) {
			err = errors.New("request must contain one JSON object")
		}
	}
	if err != nil {
		logger.Warn("invalid request body", slog.String("reason", "invalid_json"), slog.String("result", "validation_error"))
		return false, FailureWithHttpStatus(c, fiber.StatusBadRequest, "Некорректное тело запроса.")
	}
	if _, err := govalidator.ValidateStruct(*request); err != nil {
		fields := make([]string, 0)
		for field := range govalidator.ErrorsByField(err) {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		fieldErrors := govalidator.ErrorsByField(err)
		validationErrors := make([]string, 0, len(fieldErrors))
		for _, message := range fieldErrors {
			validationErrors = append(validationErrors, message)
		}
		if len(validationErrors) == 0 {
			validationErrors = append(validationErrors, err.Error())
		}
		sort.Strings(validationErrors)

		logger.Warn("request validation failed", slog.Any("fields", fields), slog.String("result", "validation_error"))
		return false, FailureWithHttpStatus(c, fiber.StatusBadRequest, validationErrors...)
	}

	return true, nil
}

func handlerLogger(c *fiber.Ctx, operation string) *slog.Logger {
	return logctx.Logger(c.UserContext()).With(slog.String("component", "ContractServer"), slog.String("operation", operation))
}
