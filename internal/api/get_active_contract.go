package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
)

func (s *Server) GetActiveContract(c *fiber.Ctx) error {
	var req gen.GetActiveContractRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(gen.Response{
			Data:   nil,
			Errors: []string{"Некорректный запрос."},
		})
	}

	// Временные данные для проверки API; чтение договора из БД пока не реализовано.
	return c.Status(fiber.StatusOK).JSON(gen.Response{
		Data: gen.GetActiveContractResponse{
			ID:        uuid.New(),
			CompanyID: req.CompanyID,
			Status:    "active",
			StartsAt:  time.Now(),
			ExpiresAt: nil,
			Options: gen.ContractOptions{
				{Code: gen.TripCreation},
				{Code: gen.TripParticipants},
				{Code: gen.Notifications},
			},
		},
		Errors: []string{},
	})
}
