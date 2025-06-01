package natslisteners

import (
	httpsender "notification-service/internal/transport/http"
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
	httpSender httpsender.HTTPSender
}

type Config struct {
	NatsConn *nats.Conn
	Log      *zap.Logger
	Timeout  time.Duration

	SMTPSender smtpsender.SMTPSender
	HTTPSender httpsender.HTTPSender
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:   cfg.NatsConn,
		log:        cfg.Log,
		timeout:    cfg.Timeout,
		smtpSender: cfg.SMTPSender,
		httpSender: cfg.HTTPSender,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
