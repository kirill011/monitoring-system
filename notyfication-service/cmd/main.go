package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"notification-service/config"
	"notification-service/internal/app"
)

func main() {
	defer os.Exit(1)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	app.Run(ctx, config.Get(), stop)
}
