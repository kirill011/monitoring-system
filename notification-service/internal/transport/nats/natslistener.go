package natslisteners

import (
	"notification-service/internal/services"
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
	js       nats.JetStreamContext

	smtpSender smtpsender.SMTPSender
	httpSender httpsender.HTTPSender

	notificationService *services.NotificationService
}

type Config struct {
	NatsConn *nats.Conn
	Log      *zap.Logger
	Timeout  time.Duration
	Js       nats.JetStreamContext

	SMTPSender smtpsender.SMTPSender
	HTTPSender httpsender.HTTPSender

	NotificationService *services.NotificationService
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:            cfg.NatsConn,
		js:                  cfg.Js,
		log:                 cfg.Log,
		timeout:             cfg.Timeout,
		smtpSender:          cfg.SMTPSender,
		httpSender:          cfg.HTTPSender,
		notificationService: cfg.NotificationService,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
