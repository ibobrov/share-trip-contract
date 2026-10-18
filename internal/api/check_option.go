package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/service"
)

func (s *Server) CheckOption(c *fiber.Ctx) error {
	logger := handlerLogger(c, "CheckOption")
	var req gen.CheckOptionRequest
	if ok, err := ParseAndValidateRequest(&req, c, logger); !ok {
		return err
	}
	res, err := s.contractService.CheckOption(c.UserContext(), service.CheckOptionRequest{
		CompanyID:  req.CompanyID,
		OptionCode: string(req.OptionCode),
	})
	if err != nil {
		return HandleError(c, err)
	}
	response := gen.CheckOptionResponse{Allowed: res.Allowed}
	if res.Reason != nil {
		reason := gen.CheckOptionResponseReason(*res.Reason)
		response.Reason = &reason
	}
	logger.Info("option checked")
	return Success(c, response)
}
