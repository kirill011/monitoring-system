package app

import (
	"context"
	"fmt"

	"api-gateway-service/config"

	"api-gateway-service/internal/transport/http"
	"api-gateway-service/internal/transport/natshandlers/auth"
	"api-gateway-service/internal/transport/natshandlers/devices"
	"api-gateway-service/pkg/closer"
	"api-gateway-service/pkg/logger"
	"api-gateway-service/pkg/nats"

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

	authHandlers := auth.NewAuthHandlers(auth.Config{
		NatsConn: nats.NatsConn,
		Timeout:  cfg.Nats.Timeout,
	})

	devicesHandlers := devices.NewDevicesHandlers(devices.Config{
		NatsConn: nats.NatsConn,
		Timeout:  cfg.Nats.Timeout,
	})

	httpServer := http.NewServer(http.Config{
		Log:             log,
		JwtKey:          cfg.Server.JwtKey,
		Addr:            cfg.Server.Addr,
		LogQuerys:       cfg.Server.LogQuerys,
		AuthHandlers:    authHandlers,
		DevicesHandlers: devicesHandlers,
	})

	go func() {
		if err = httpServer.Run(); err != nil {
			log.Error(fmt.Sprintf("error occurred while running HTTP server: %v", err))
			stop()
		}
	}()

	log.Info("start http server", zap.String("listen_on", cfg.Server.Addr))

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
