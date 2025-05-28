package natslisteners

import (
	"auth-service/internal/services"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsListeners struct {
	natsConn    *nats.Conn
	authService services.Auth
	log         *zap.Logger
}

type Config struct {
	NatsConn    *nats.Conn
	AuthService services.Auth
	Log         *zap.Logger
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:    cfg.NatsConn,
		authService: cfg.AuthService,
		log:         cfg.Log,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
