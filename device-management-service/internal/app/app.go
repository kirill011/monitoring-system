package app

import (
	"context"
	"fmt"

	"device-management-service/config"
	"device-management-service/internal/repo/pg"
	"device-management-service/internal/services"
	natslisteners "device-management-service/internal/transport/nats"
	"device-management-service/pkg/closer"
	"device-management-service/pkg/logger"
	"device-management-service/pkg/nats"
	"device-management-service/pkg/postgres"

	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg *config.Config, stop context.CancelFunc) {
	log := logger.New(logger.Config{
		LogLevel:    cfg.Logger.LogLevel,
		ServiceName: cfg.Logger.ServiceName,
		LogPath:     cfg.Logger.LogPath,
	})

	postgresDB, err := postgres.NewPostgres(postgres.Config{
		DataSource:        cfg.Postgres.DataSource,
		ApplicationSchema: cfg.Postgres.ApplicationSchema,
	})
	if err != nil {
		log.Fatal("init postgresDB error", zap.Error(err))
	}

	nats, err := nats.NewNats(cfg.Nats.URL, log)
	if err != nil {
		log.Fatal(fmt.Errorf("nats.NewNats: %w", err).Error())
	}

	_, err = postgresDB.RunMigrations(cfg.Postgres.PathToMigrations, cfg.Postgres.ApplicationSchema)
	if err != nil {
		log.Fatal("migration error", zap.Error(err))
	}

	postgresDB, err = postgres.NewPostgres(postgres.Config{
		DataSource:        cfg.Postgres.DataSource,
		ApplicationSchema: cfg.Postgres.ApplicationSchema,
	})
	if err != nil {
		log.Fatal("init postgresDB error", zap.Error(err))
	}

	devicesRepo := pg.NewDevicesRepo(postgresDB, log)

	devicesService := services.NewDevicesService(devicesRepo)

	listeners := natslisteners.NewListener(natslisteners.Config{
		NatsConn:       nats.NatsConn,
		DevicesService: devicesService,
		Log:            log,
	})

	go func() {
		if err = listeners.Run(); err != nil {
			log.Error(fmt.Errorf("error occurred while running NATs listeners: %w", err).Error())
			stop()
		}
	}()

	log.Info("start nats listeners", zap.String("listen_on", cfg.Nats.URL))

	// Shutdown
	<-ctx.Done()

	log.Info("Shutdown start")

	closer := closer.Closer{}

	closer.Add(nats.Close)
	closer.Add(postgresDB.Close)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := closer.Close(shutdownCtx); err != nil {
		log.Error("Closer", zap.Error(err))
		return
	}

	log.Info("Shutdown success")
}
