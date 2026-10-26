package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/service"
)

func (s *Server) CreateContract(c *fiber.Ctx) error {
	logger := handlerLogger(c, "CreateContract")
	var req gen.CreateContractRequest
	if ok, err := ParseAndValidateRequest(&req, c, logger); !ok {
		return err
	}
	options := make([]string, 0, len(req.Options))
	for _, option := range req.Options {
		options = append(options, string(option.Code))
	}
	res, err := s.contractService.CreateContract(c.UserContext(), service.CreateContractRequest{
		CompanyID: req.CompanyID,
		StartsAt:  req.StartsAt,
		ExpiresAt: req.ExpiresAt,
		Options:   options,
	})
	if err != nil {
		return HandleError(c, logger, err)
	}
	responseOptions := make(gen.ContractOptions, 0, len(res.Options))
	for _, code := range res.Options {
		responseOptions = append(responseOptions, gen.ContractOption{
			Code: gen.OptionCode(code),
		})
	}

	logger.Info("contract created", "result", "success", "contract_id", res.ID, "company_id", res.CompanyID)
	return SuccessWithHttpStatus(c, fiber.StatusCreated, gen.CreateContractResponse{
		ID:        res.ID,
		CompanyID: res.CompanyID,
		Status:    gen.CreateContractResponseStatus(res.Status),
		StartsAt:  res.StartsAt,
		ExpiresAt: res.ExpiresAt,
		Options:   responseOptions,
	})
}
