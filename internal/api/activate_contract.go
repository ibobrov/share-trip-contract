package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
)

func (s *Server) ActivateContract(c *fiber.Ctx) error {
	var req gen.ActivateContractRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(gen.Response{
			Data:   nil,
			Errors: []string{"Некорректный запрос."},
		})
	}

	// Заглушка: изменение статуса договора в БД пока не реализовано.
	return c.Status(fiber.StatusCreated).JSON(gen.Response{
		Data:   gen.ActivateContractResponse(req),
		Errors: []string{},
	})
}
