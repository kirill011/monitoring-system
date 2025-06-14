package app

import (
	"context"
	"fmt"
	"net/http"

	"notification-service/config"
	"notification-service/internal/services"
	httpsender "notification-service/internal/transport/http"
	natslisteners "notification-service/internal/transport/nats"
	smtpsender "notification-service/internal/transport/smtp"
	"notification-service/pkg/closer"
	"notification-service/pkg/logger"
	"notification-service/pkg/nats"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, stop context.CancelFunc) {
	log := logger.New(logger.Config{
		LogLevel:    cfg.Logger.LogLevel,
		ServiceName: cfg.Logger.ServiceName,
		LogPath:     cfg.Logger.LogPath,
	})

	nats, err := nats.NewNats(cfg.Nats.URL, log)
	if err != nil {
		log.Fatal(fmt.Errorf("nats.NewNats: %w", err).Error())
	}

	notificationService := services.NewNotificationService()

	smtpsender := smtpsender.New(smtpsender.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		User:     cfg.SMTP.User,
		Password: cfg.SMTP.Password,
	})

	httpsender := httpsender.New(httpsender.Config{
		Host:     cfg.HTTP.Host,
		Port:     cfg.HTTP.Port,
		Endpoint: cfg.HTTP.Endpoint,
	})

	listeners := natslisteners.NewListener(natslisteners.Config{
		NatsConn:            nats.NatsConn,
		Js:                  nats.Js,
		Log:                 log,
		Timeout:             cfg.Nats.Timeout,
		SMTPSender:          smtpsender,
		HTTPSender:          httpsender,
		NotificationService: notificationService,
	})

	go func() {
		if err = listeners.Run(); err != nil {
			log.Error(fmt.Errorf("error occurred while running NATs listeners: %w", err).Error())
			stop()
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9081", nil)
	}()

	log.Info("Running NATs listener")

	// req := pbnotification.SendNotifyReq{
	// 	DeviceID: 1,
	// 	Text:     "Hello, World!",
	// 	Subject:  "Problem",
	// }
	// b, _ := proto.Marshal(&req)
	// nats.NatsConn.Publish("notify.send", b)
	// Shutdown
	<-ctx.Done()

	log.Info("Shutdown start")

	closer := closer.Closer{}

	closer.Add(nats.Close)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := closer.Close(shutdownCtx); err != nil {
		log.Error("Closer", zap.Error(err))
		return
	}

	log.Info("Shutdown success")
}
