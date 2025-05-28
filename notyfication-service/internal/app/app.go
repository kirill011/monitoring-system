package app

import (
	"context"
	"fmt"

	"notification-service/config"
	natslisteners "notification-service/internal/transport/nats"
	smtpsender "notification-service/internal/transport/smtp"
	"notification-service/pkg/closer"
	"notification-service/pkg/logger"
	"notification-service/pkg/nats"

	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, stop context.CancelFunc) {
	log := logger.New(logger.Config{
		LogLevel:    cfg.Logger.LogLevel,
		ServiceName: cfg.Logger.ServiceName,
		LogPath:     cfg.Logger.LogPath,
	})

	nats, err := nats.NewNats(cfg.Nats.URL)
	if err != nil {
		log.Fatal(fmt.Errorf("nats.NewNats: %w", err).Error())
	}

	smtpsender := smtpsender.New(smtpsender.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		User:     cfg.SMTP.User,
		Password: cfg.SMTP.Password,
	})

	listeners := natslisteners.NewListener(natslisteners.Config{
		NatsConn:   nats.NatsConn,
		Log:        log,
		Timeout:    cfg.Nats.Timeout,
		SMTPSender: smtpsender,
	})

	go func() {
		if err = listeners.Run(); err != nil {
			log.Error(fmt.Errorf("error occurred while running NATs listeners: %w", err).Error())
			stop()
		}
	}()

	log.Info("Running NATs listener")

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
