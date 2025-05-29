package natslisteners

import (
	"device-management-service/internal/services"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsListeners struct {
	natsConn       *nats.Conn
	devicesService services.Devices
	log            *zap.Logger
}

type Config struct {
	NatsConn       *nats.Conn
	DevicesService services.Devices
	Log            *zap.Logger
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:       cfg.NatsConn,
		devicesService: cfg.DevicesService,
		log:            cfg.Log,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
