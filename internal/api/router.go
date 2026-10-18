package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
	"github.com/ibobrov/share-trip-contract/internal/observability/middleware"
)

func (s *Server) Route(router fiber.Router) {
	gen.RegisterHandlers(router.Group("", middleware.TimeoutMiddleware(2*time.Second)), s)
}
