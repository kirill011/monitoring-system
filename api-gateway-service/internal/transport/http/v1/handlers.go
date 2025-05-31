package v1

import (
	authHandlers "api-gateway-service/internal/transport/http/v1/auth"
	devicesHandlers "api-gateway-service/internal/transport/http/v1/devices"
	reportsHandlers "api-gateway-service/internal/transport/http/v1/reports"
	tagsHandlers "api-gateway-service/internal/transport/http/v1/tags"

	"api-gateway-service/internal/transport/natshandlers/auth"
	"api-gateway-service/internal/transport/natshandlers/devices"
	"api-gateway-service/internal/transport/natshandlers/reports"
	"api-gateway-service/internal/transport/natshandlers/tags"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	jwtKey          string
	authHandlers    *auth.AuthHandlers
	devicesHandlers *devices.DevicesHandler
	tagsHandlers    *tags.TagsHandler
	reportsHandlers *reports.ReportsHandler
}

type Config struct {
	JwtKey          string
	AuthHandlers    *auth.AuthHandlers
	DevicesHandlers *devices.DevicesHandler
	TagsHandlers    *tags.TagsHandler
	ReportsHandlers *reports.ReportsHandler
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		jwtKey:          cfg.JwtKey,
		authHandlers:    cfg.AuthHandlers,
		devicesHandlers: cfg.DevicesHandlers,
		tagsHandlers:    cfg.TagsHandlers,
		reportsHandlers: cfg.ReportsHandlers,
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

	tagsHandlers.NewTagsHandler(&tagsHandlers.Config{
		NatsHandlers: h.tagsHandlers,
		JWTKey:       h.jwtKey,
	}).InitTagsRoutes(routeV1)

	reportsHandlers.NewReportsHandler(&reportsHandlers.Config{
		NatsHandlers: h.reportsHandlers,
		JWTKey:       h.jwtKey,
	}).InitReportsRoutes(routeV1)
}
