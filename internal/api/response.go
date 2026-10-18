package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
)

func Success(c *fiber.Ctx, data any) error {
	return SuccessWithHttpStatus(c, fiber.StatusOK, data)
}

func SuccessWithHttpStatus(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(gen.Response{
		Data:   data,
		Errors: []string{},
	})
}

func FailureWithHttpStatus(c *fiber.Ctx, status int, errs ...string) error {
	return c.Status(status).JSON(gen.Response{
		Data:   nil,
		Errors: errs,
	})
}
