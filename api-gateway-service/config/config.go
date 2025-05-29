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
	Server ServerConfig
	Nats   NatsConfig
}

type Logger struct {
	LogLevel    string `env:"LOG_LEVEL,required"`
	ServiceName string `env:"LOG_SERVICE_NAME,required"`
	LogPath     string `env:"LOG_PATH"`
}

type AppConfig struct {
	ShutdownTimeout time.Duration `env:"APP_SHUTDOWN_TIMEOUT,required"`
}

type ServerConfig struct {
	JwtKey        string        `env:"SERVER_JWT_KEY,required"`
	Addr          string        `env:"SERVER_ADDR,required"`
	TokenLifeTime time.Duration `env:"SERVER_TOKEN_LIFE_TIME,required"`
	LogQuerys     bool          `env:"SERVER_LOG_QUERYS"`
}

type NatsConfig struct {
	URL     string        `env:"NATS_URL"`
	Timeout time.Duration `env:"NATS_TIMEOUT,required"`
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
