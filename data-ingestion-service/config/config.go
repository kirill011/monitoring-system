package config

import (
	"log"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Logger Logger
	App    AppConfig
	Nats   NatsConfig
	Server ServerConfig
}

type Logger struct {
	LogLevel    string `env:"LOG_LEVEL,required"`
	ServiceName string `env:"LOG_SERVICE_NAME,required"`
	LogPath     string `env:"LOG_PATH,required"`
}

type AppConfig struct {
	ShutdownTimeout time.Duration `env:"APP_SHUTDOWN_TIMEOUT,required"`
}

type NatsConfig struct {
	URL     string        `env:"NATS_URL,required"`
	Timeout time.Duration `env:"NATS_TIMEOUT,required"`
}

type ServerConfig struct {
	Addr              string `env:"SERVER_ADDR,required"`
	DeviceCheckPeriod int    `env:"SERVER_DEVICE_CHECK_PERIOD"`
	LogQuerys         bool   `env:"SERVER_LOG_QUERYS"`
}

var (
	config Config
	once   sync.Once
)

func Get() *Config {
	once.Do(func() {
		_ = godotenv.Load()
		if err := env.Parse(&config); err != nil {
			log.Fatal(err)
		}
	})
	return &config
}
