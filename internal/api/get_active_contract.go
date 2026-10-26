package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/service"
)

func (s *Server) GetActiveContract(c *fiber.Ctx) error {
	logger := handlerLogger(c, "GetActiveContract")
	var req gen.GetActiveContractRequest
	if ok, err := ParseAndValidateRequest(&req, c, logger); !ok {
		return err
	}
	res, err := s.contractService.GetActiveContract(c.UserContext(), service.GetActiveContractRequest{
		CompanyID: req.CompanyID,
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

	logger.Info("active contract retrieved", "result", "success", "contract_id", res.ID, "company_id", res.CompanyID)
	return Success(c, gen.GetActiveContractResponse{
		ID:        res.ID,
		CompanyID: res.CompanyID,
		Status:    gen.GetActiveContractResponseStatus(res.Status),
		StartsAt:  res.StartsAt,
		ExpiresAt: res.ExpiresAt,
		Options:   responseOptions,
	})
}
