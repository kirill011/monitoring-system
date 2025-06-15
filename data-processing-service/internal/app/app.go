package app

import (
	"context"
	"fmt"

	"data-processing-service/config"
	"data-processing-service/internal/repo/pg"
	"data-processing-service/internal/services"
	"data-processing-service/internal/transport/http"
	messagelisteners "data-processing-service/internal/transport/nats/messages"
	tagslistener "data-processing-service/internal/transport/nats/tags"
	"data-processing-service/pkg/closer"
	"data-processing-service/pkg/logger"
	"data-processing-service/pkg/nats"
	"data-processing-service/pkg/postgres"

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

	tagsRepo := pg.NewTagsRepo(postgresDB, log)
	messagesRepo := pg.NewMessagesRepo(postgresDB, log)

	messagesService := services.NewMessagesService(services.Config{
		MessageRepo:        messagesRepo,
		TagRepo:            tagsRepo,
		Log:                log,
		NotificationPeriod: cfg.Service.NotificationPeriod,
	})

	tagsService := services.NewTagsService(tagsRepo, messagesService)

	messagesListeners := messagelisteners.NewListener(messagelisteners.Config{
		NatsConn:        nats.NatsConn,
		Js:              nats.Js,
		MessagesService: messagesService,
		Timeout:         cfg.Nats.Timeout,
		Log:             log,
	})

	tagsListener := tagslistener.NewListener(tagslistener.Config{
		NatsConn:    nats.NatsConn,
		TagsService: tagsService,
		Log:         log,
	})

	httpServer := http.NewServer(http.Config{
		Log:            log,
		JwtKey:         cfg.Server.JwtKey,
		Addr:           cfg.Server.Addr,
		LogQuerys:      cfg.Server.LogQuerys,
		ReportsHandler: messagesService,
	})

	go func() {
		if err = httpServer.Run(); err != nil {
			log.Error(fmt.Sprintf("error occurred while running HTTP server: %v", err))
			stop()
		}
	}()

	log.Info("start http server", zap.String("listen_on", cfg.Server.Addr))

	go func() {
		if err = messagesListeners.Listen(); err != nil {
			log.Error(fmt.Errorf("error occurred while running messagesListeners: %w", err).Error())
			stop()
		}
	}()

	go func() {
		if err = tagsListener.Listen(); err != nil {
			log.Error(fmt.Errorf("error occurred while running tagsListener: %w", err).Error())
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
	closer.Add(httpServer.Stop)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := closer.Close(shutdownCtx); err != nil {
		log.Error("Closer", zap.Error(err))
		return
	}

	log.Info("Shutdown success")
}
