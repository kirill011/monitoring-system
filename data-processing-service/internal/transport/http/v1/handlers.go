package v1

import (
	"data-processing-service/internal/services"
	reportsHandlers "data-processing-service/internal/transport/http/v1/reports"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	jwtKey          string
	reportsHandlers services.Messages
}

type Config struct {
	JwtKey          string
	ReportsHandlers services.Messages
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		jwtKey:          cfg.JwtKey,
		reportsHandlers: cfg.ReportsHandlers,
	}
}

func (h *Handler) InitRouter(routeV1 fiber.Router) {
	reportsHandlers.NewReportsHandler(&reportsHandlers.Config{
		MessageService: h.reportsHandlers,
		JWTKey:         h.jwtKey,
	}).InitReportsRoutes(routeV1)
}
