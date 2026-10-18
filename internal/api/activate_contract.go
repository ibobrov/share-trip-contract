package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/service"
)

func (s *Server) ActivateContract(c *fiber.Ctx) error {
	logger := handlerLogger(c, "ActivateContract")
	var req gen.ActivateContractRequest
	if ok, err := ParseAndValidateRequest(&req, c, logger); !ok {
		return err
	}
	res, err := s.contractService.ActivateContract(c.UserContext(), service.ActivateContractRequest{
		ID: req.ID,
	})
	if err != nil {
		return HandleError(c, err)
	}
	logger.Info("contract activated")
	return SuccessWithHttpStatus(c, fiber.StatusCreated, gen.ActivateContractResponse{
		ID: res.ID,
	})
}
