package app

import (
	"context"
	"fmt"

	"data-ingestion-service/config"
	natslisteners "data-ingestion-service/internal/transport/natslistener"
	"data-ingestion-service/pkg/closer"
	"data-ingestion-service/pkg/logger"
	"data-ingestion-service/pkg/nats"

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

	listeners := natslisteners.NewListener(natslisteners.Config{
		NatsConn: nats.NatsConn,
		Log:      log,
		Timeout:  cfg.Nats.Timeout,
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
