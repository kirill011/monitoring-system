package app

import (
	"context"
	"fmt"

	"data-ingestion-service/config"
	"data-ingestion-service/internal/services"
	"data-ingestion-service/internal/transport/http"
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

	devicesService := services.NewDeviceService(services.Config{})

	listeners := natslisteners.NewListener(natslisteners.Config{
		NatsConn:       nats.NatsConn,
		Log:            log,
		Timeout:        cfg.Nats.Timeout,
		DevicesService: devicesService,
	})

	go func() {
		if err = listeners.Run(); err != nil {
			log.Error(fmt.Errorf("error occurred while running NATs listeners: %w", err).Error())
			stop()
		}
	}()

	log.Info("Running NATs listener")

	httpServer := http.NewServer(http.Config{
		Log:               log,
		Addr:              cfg.Server.Addr,
		LogQuerys:         cfg.Server.LogQuerys,
		MessageHandler:    listeners,
		DevicesService:    devicesService,
		DeviceCheckPeriod: cfg.Server.DeviceCheckPeriod,
	})

	go func() {
		if err = httpServer.Run(); err != nil {
			log.Error(fmt.Errorf("error occurred while running http server: %w", err).Error())
			stop()
		}
	}()

	log.Info("Running HTTP server")

	// Shutdown
	<-ctx.Done()

	log.Info("Shutdown start")

	closer := closer.Closer{}

	closer.Add(nats.Close)
	closer.Add(httpServer.Stop)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := closer.Close(shutdownCtx); err != nil {
		log.Error("Closer", zap.Error(err))
		return
	}

	log.Info("Shutdown success")
}
