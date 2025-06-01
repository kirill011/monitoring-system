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
	SMTP   SMTPConfig
	HTTP   HTTPConfig
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

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST,required"`
	Port     string `env:"SMTP_PORT,required"`
	User     string `env:"SMTP_USER,required"`
	Password string `env:"SMTP_PASSWORD,required"`
}

type HTTPConfig struct {
	Host     string `env:"HTTP_HOST,required"`
	Port     string `env:"HTTP_PORT,required"`
	Endpoint string `env:"HTTP_ENDPOINT,required"`
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
