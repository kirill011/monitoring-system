package app

import (
	"context"
	"fmt"

	"auth-service/config"
	"auth-service/internal/repo/pg"
	"auth-service/internal/services"
	natslisteners "auth-service/internal/transport/nats"
	"auth-service/pkg/closer"
	"auth-service/pkg/logger"
	"auth-service/pkg/nats"
	"auth-service/pkg/postgres"

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

	nats, err := nats.NewNats(cfg.Nats.URL)
	if err != nil {
		log.Fatal(fmt.Errorf("nats.NewNats: %w", err).Error())
	}

	_, err = postgresDB.RunMigrations(cfg.Postgres.PathToMigrations, cfg.Postgres.ApplicationSchema)
	if err != nil {
		log.Fatal("migration error", zap.Error(err))
	}

	authRepo := pg.NewAuthRepo(postgresDB, log)

	authService := services.NewAuthService(authRepo)

	listeners := natslisteners.NewListener(natslisteners.Config{
		NatsConn:      nats.NatsConn,
		AuthService:   authService,
		TokenLifeTime: cfg.Token.TokenLifeTime,
		JWTKey:        cfg.Token.JWTKey,
		Log:           log,
	})

	go func() {
		if err = listeners.Run(); err != nil {
			log.Error(fmt.Errorf("error occurred while running NATs listeners: %w", err).Error())
			stop()
		}
	}()

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
