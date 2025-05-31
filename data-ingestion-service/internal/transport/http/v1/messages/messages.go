package messages

import (
	"data-ingestion-service/internal/models"
	"data-ingestion-service/internal/transport/natslistener"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"github.com/gofiber/fiber/v3"
)

const (
	localID = "localID"
)

type messagesHandler struct {
	natsHandlers *natslistener.NatsListeners
}

type Config struct {
	NatsHandlers *natslistener.NatsListeners
}

func NewMessagesHandler(cfg *Config) *messagesHandler {
	return &messagesHandler{
		natsHandlers: cfg.NatsHandlers,
	}
}

func (h *messagesHandler) InitMessagesRoutes(api fiber.Router) {
	servicesRoute := api.Group("/messages")
	servicesRoute.Post("/send_msg", h.sendMsg)
}

type (
	sendMsgReq struct {
		Message     string `form:"message" 		json:"message" 		validate:"required" 	xml:"message"`
		MessageType string `form:"message_type" json:"message_type" validate:"required" 	xml:"message_type"`
		Component   string `form:"component" 	json:"component" 	validate:"required" 	xml:"component"`
		Address     string `form:"address" 		json:"address" 		validate:"required,ip" 	xml:"address"`
	}
)

func (h *messagesHandler) sendMsg(ctx fiber.Ctx) error {
	body := sendMsgReq{
		Message:     "",
		MessageType: "",
		Component:   "",
		Address:     "",
	}

	if err := ctx.Bind().Body(&body); err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity,
			fmt.Errorf("ctx.Bind().Body: %w", err).Error(),
		)
	}

	err := h.natsHandlers.PublishSaveMessage(
		models.Message{
			Message:     body.Message,
			MessageType: body.MessageType,
			Component:   body.Component,
			DeviceIP:    body.Address,
		},
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("h.nats.PublishGetAllByDeviceId: %w", err).Error(),
		)
	}

	jsonResponse, err := jsoniter.Marshal(
		fiber.StatusAccepted,
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("json.Marshal: %w", err).Error(),
		)
	}

	if err = ctx.Status(fiber.StatusOK).Send(jsonResponse); err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			fmt.Errorf("ctx.Send: %w", err).Error(),
		)
	}

	return nil
}
