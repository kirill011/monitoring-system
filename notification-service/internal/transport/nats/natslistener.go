package natslisteners

import (
	"notification-service/internal/services"
	httpsender "notification-service/internal/transport/http"
	smtpsender "notification-service/internal/transport/smtp"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
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
	metrics             prometheus.Counter
}

type Config struct {
	NatsConn *nats.Conn
	Log      *zap.Logger
	Timeout  time.Duration
	Js       nats.JetStreamContext

	SMTPSender smtpsender.SMTPSender
	HTTPSender httpsender.HTTPSender

	NotificationService *services.NotificationService

	Metrics prometheus.Counter
}

func NewListener(cfg Config) *NatsListeners {
	egressResponses := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "egress_responses_total",
			Help: "Total async responses sent",
		},
	)

	prometheus.MustRegister(egressResponses)

	return &NatsListeners{
		natsConn:            cfg.NatsConn,
		js:                  cfg.Js,
		log:                 cfg.Log,
		timeout:             cfg.Timeout,
		smtpSender:          cfg.SMTPSender,
		httpSender:          cfg.HTTPSender,
		notificationService: cfg.NotificationService,
		metrics:             egressResponses,
	}
}

func (n *NatsListeners) Run() error {
	return n.listen()
}
