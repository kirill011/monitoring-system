package natslistener

import (
	"data-ingestion-service/internal/services"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsListeners struct {
	natsConn *nats.Conn
	log      *zap.Logger
	timeout  time.Duration

	devicesService services.DeviceService
}

type Config struct {
	NatsConn *nats.Conn
	Log      *zap.Logger
	Timeout  time.Duration

	DevicesService services.DeviceService
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:       cfg.NatsConn,
		log:            cfg.Log,
		timeout:        cfg.Timeout,
		devicesService: cfg.DevicesService,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
