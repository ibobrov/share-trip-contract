package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/service"
)

func (s *Server) GetContractById(c *fiber.Ctx) error {
	logger := handlerLogger(c, "GetContractById")
	var req gen.GetContractByIdRequest
	if ok, err := ParseAndValidateRequest(&req, c, logger); !ok {
		return err
	}
	res, err := s.contractService.GetContractById(c.UserContext(), service.GetContractByIdRequest{
		ID: req.ID,
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

	logger.Info("contract retrieved", "result", "success", "contract_id", res.ID, "company_id", res.CompanyID)
	return Success(c, gen.GetContractByIdResponse{
		ID:        res.ID,
		CompanyID: res.CompanyID,
		Status:    gen.GetContractByIdResponseStatus(res.Status),
		StartsAt:  res.StartsAt,
		ExpiresAt: res.ExpiresAt,
		Options:   responseOptions,
	})
}
