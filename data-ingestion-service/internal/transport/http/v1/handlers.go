package v1

import (
	"data-ingestion-service/internal/services"
	"data-ingestion-service/internal/transport/http/v1/devicechecker"
	"data-ingestion-service/internal/transport/http/v1/messages"
	"data-ingestion-service/internal/transport/natslistener"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type Handler struct {
	messageHandlers *natslistener.NatsListeners
	devicesService  services.DeviceService

	deviceCheckPeriod int
	log               *zap.Logger
}

type Config struct {
	MessageHandler *natslistener.NatsListeners
	DevicesService services.DeviceService

	DeviceCheckPeriod int
	Log               *zap.Logger
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		messageHandlers:   cfg.MessageHandler,
		devicesService:    cfg.DevicesService,
		deviceCheckPeriod: cfg.DeviceCheckPeriod,
		log:               cfg.Log,
	}
}

func (h *Handler) InitRouter(routeV1 fiber.Router) {
	messages.NewMessagesHandler(
		&messages.Config{
			NatsHandlers: h.messageHandlers,
		},
	).InitMessagesRoutes(routeV1)

	devicechecker.NewDeviceCheckerHandler(
		&devicechecker.Config{
			DeviceService: h.devicesService,
			NatsHandlers:  h.messageHandlers,

			Logger:            h.log,
			DeviceCheckPeriod: h.deviceCheckPeriod,
		},
	).Start()
}
