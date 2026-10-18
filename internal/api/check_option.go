package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
)

func (s *Server) CheckOption(c *fiber.Ctx) error {
	var req gen.CheckOptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(gen.Response{
			Data:   nil,
			Errors: []string{"Некорректный запрос."},
		})
	}

	// Заглушка: известные услуги разрешены, проверка договора в БД пока не реализована.
	res := gen.CheckOptionResponse{Allowed: req.OptionCode.Valid()}
	if !res.Allowed {
		reason := gen.SERVICENOTALLOWED
		res.Reason = &reason
	}

	return c.Status(fiber.StatusOK).JSON(gen.Response{
		Data:   res,
		Errors: []string{},
	})
}
