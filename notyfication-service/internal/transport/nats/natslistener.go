package natslisteners

import (
	smtpsender "notification-service/internal/transport/smtp"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsListeners struct {
	natsConn *nats.Conn
	log      *zap.Logger
	timeout  time.Duration

	smtpSender smtpsender.SMTPSender
}

type Config struct {
	NatsConn *nats.Conn
	Log      *zap.Logger
	Timeout  time.Duration

	SMTPSender smtpsender.SMTPSender
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:   cfg.NatsConn,
		log:        cfg.Log,
		timeout:    cfg.Timeout,
		smtpSender: cfg.SMTPSender,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
