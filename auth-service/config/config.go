package config

import (
	"log"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Logger   Logger
	Postgres PostgresConfig
	App      AppConfig
	Token    Token
	Nats     NatsConfig
}

type Logger struct {
	LogLevel    string `env:"LOG_LEVEL,required"`
	ServiceName string `env:"LOG_SERVICE_NAME,required"`
	LogPath     string `env:"LOG_PATH"`
}

type PostgresConfig struct {
	DataSource        string `env:"DB_DATA_SOURCE,required"`
	PathToMigrations  string `env:"DB_PATH_TO_MIGRATION,required"`
	ApplicationSchema string `env:"DB_APPLICATION_SCHEMA,required"`
}

type AppConfig struct {
	ShutdownTimeout time.Duration `env:"APP_SHUTDOWN_TIMEOUT,required"`
}

type Token struct {
	JWTKey        string        `env:"TOKEN_JWT_KEY,required"`
	TokenLifeTime time.Duration `env:"TOKEN_LIFE_TIME"`
}

type NatsConfig struct {
	URL string `env:"NATS_URL"`
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
