package v1

import (
	"time"

	"auth-service/internal/services"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	jwtKey        string
	tokenLifeTime time.Duration

	authService services.Auth
}

type Config struct {
	AuthService   services.Auth
	JwtKey        string
	TokenLifeTime time.Duration
}

func NewHandler(cfg Config) *Handler {
	return &Handler{
		authService:   cfg.AuthService,
		jwtKey:        cfg.JwtKey,
		tokenLifeTime: cfg.TokenLifeTime,
	}
}

func (h *Handler) InitRouter(routeV1 fiber.Router) {
	h.initAuthRoutes(routeV1)
}
