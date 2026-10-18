package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ibobrov/share-trip-contract/gen"
)

type Server struct{}

var _ gen.ServerInterface = (*Server)(nil)

func NewServer() *Server {
	return &Server{}
}

func RegisterRoutes(router fiber.Router) {
	gen.RegisterHandlers(router, NewServer())
}
