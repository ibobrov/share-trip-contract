package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ibobrov/share-trip-contract/gen"
)

func (s *Server) CreateContract(c *fiber.Ctx) error {
	var req gen.CreateContractRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(gen.Response{
			Data:   nil,
			Errors: []string{"Некорректный запрос."},
		})
	}

	// Заглушка: сохранение договора в БД пока не реализовано.
	return c.Status(fiber.StatusCreated).JSON(gen.Response{
		Data: gen.CreateContractResponse{
			ID:        uuid.New(),
			CompanyID: req.CompanyID,
			Status:    "draft",
			StartsAt:  req.StartsAt,
			ExpiresAt: req.ExpiresAt,
			Options:   req.Options,
		},
		Errors: []string{},
	})
}
