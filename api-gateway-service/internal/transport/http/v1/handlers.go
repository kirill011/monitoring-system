package v1

import (
	authHandlers "api-gateway-service/internal/transport/http/v1/auth"
	devicesHandlers "api-gateway-service/internal/transport/http/v1/devices"

	"api-gateway-service/internal/transport/natshandlers/auth"
	"api-gateway-service/internal/transport/natshandlers/devices"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	jwtKey          string
	authHandlers    *auth.AuthHandlers
	devicesHandlers *devices.DevicesHandler
}

type Config struct {
	JwtKey          string
	AuthHandlers    *auth.AuthHandlers
	DevicesHandlers *devices.DevicesHandler
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		jwtKey:          cfg.JwtKey,
		authHandlers:    cfg.AuthHandlers,
		devicesHandlers: cfg.DevicesHandlers,
	}
}

func (h *Handler) InitRouter(routeV1 fiber.Router) {
	authHandlers.NewAuthHandler(&authHandlers.Config{
		NatsHandlers: h.authHandlers,
		JWTKey:       h.jwtKey,
	}).InitAuthRoutes(routeV1)

	devicesHandlers.NewDevicesHandler(&devicesHandlers.Config{
		NatsHandlers: h.devicesHandlers,
		JWTKey:       h.jwtKey,
	}).InitDevicesRoutes(routeV1)
}
