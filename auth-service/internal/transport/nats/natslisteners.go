package natslisteners

import (
	"auth-service/internal/services"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsListeners struct {
	natsConn      *nats.Conn
	authService   services.Auth
	tokenLifeTime time.Duration
	jwtKey        string
	log           *zap.Logger
}

type Config struct {
	NatsConn      *nats.Conn
	AuthService   services.Auth
	TokenLifeTime time.Duration
	JWTKey        string
	Log           *zap.Logger
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:      cfg.NatsConn,
		authService:   cfg.AuthService,
		tokenLifeTime: cfg.TokenLifeTime,
		jwtKey:        cfg.JWTKey,
		log:           cfg.Log,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
