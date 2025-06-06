package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v1 "api-gateway-service/internal/transport/http/v1"
	"api-gateway-service/internal/transport/natshandlers/auth"
	"api-gateway-service/internal/transport/natshandlers/devices"
	"api-gateway-service/internal/transport/natshandlers/reports"
	"api-gateway-service/internal/transport/natshandlers/tags"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Server struct {
	app             *fiber.App
	addr            string
	jwtKey          string
	log             *zap.Logger
	authHandlers    *auth.AuthHandlers
	devicesHandlers *devices.DevicesHandler
	tagsHandler     *tags.TagsHandler
	reportsHandler  *reports.ReportsHandler
}

type Config struct {
	Log             *zap.Logger
	JwtKey          string
	TokenLifeTime   time.Duration
	Addr            string
	LogQuerys       bool
	AuthHandlers    *auth.AuthHandlers
	DevicesHandlers *devices.DevicesHandler
	TagsHandler     *tags.TagsHandler
	ReportsHandler  *reports.ReportsHandler
}

func NewServer(cfg Config) *Server {
	server := &Server{
		log:             cfg.Log,
		addr:            cfg.Addr,
		jwtKey:          cfg.JwtKey,
		authHandlers:    cfg.AuthHandlers,
		devicesHandlers: cfg.DevicesHandlers,
		tagsHandler:     cfg.TagsHandler,
		reportsHandler:  cfg.ReportsHandler,
		app:             nil,
	}

	server.app = fiber.New(
		fiber.Config{
			ServerHeader:                 "",
			StrictRouting:                false,
			CaseSensitive:                false,
			Immutable:                    false,
			UnescapePath:                 false,
			BodyLimit:                    fiber.DefaultBodyLimit,
			Concurrency:                  fiber.DefaultConcurrency,
			Views:                        nil,
			ViewsLayout:                  "",
			PassLocalsToViews:            false,
			ReadTimeout:                  0,
			WriteTimeout:                 0,
			IdleTimeout:                  0,
			ReadBufferSize:               fiber.DefaultReadBufferSize,
			WriteBufferSize:              0,
			CompressedFileSuffixes:       map[string]string{},
			ProxyHeader:                  "",
			GETOnly:                      false,
			ErrorHandler:                 server.errorHandler,
			DisableKeepalive:             false,
			DisableDefaultDate:           false,
			DisableDefaultContentType:    false,
			DisableHeaderNormalizing:     false,
			AppName:                      "",
			StreamRequestBody:            false,
			DisablePreParseMultipartForm: false,
			ReduceMemoryUsage:            false,
			JSONEncoder:                  nil,
			JSONDecoder:                  nil,
			XMLEncoder:                   nil,
			EnableIPValidation:           false,
			ColorScheme: fiber.Colors{
				Black:   "",
				Red:     "",
				Green:   "",
				Yellow:  "",
				Blue:    "",
				Magenta: "",
				Cyan:    "",
				White:   "",
				Reset:   "",
			},
			StructValidator:          v1.NewValidator(),
			RequestMethods:           []string{},
			EnableSplittingOnParsers: false,
			CBOREncoder:              nil,
			CBORDecoder:              nil,
			XMLDecoder:               nil,
			TrustProxy:               false,
			TrustProxyConfig:         fiber.TrustProxyConfig{},
		},
	)

	server.setMiddleware(cfg.LogQuerys)
	server.setHandlers()

	return server
}

func (s *Server) Run() error {
	if err := s.app.Listen(s.addr, fiber.ListenConfig{
		ListenerNetwork:       "",
		CertFile:              "",
		CertKeyFile:           "",
		CertClientFile:        "",
		GracefulContext:       nil,
		TLSConfigFunc:         nil,
		ListenerAddrFunc:      nil,
		BeforeServeFunc:       nil,
		DisableStartupMessage: true,
		EnablePrefork:         false,
		EnablePrintRoutes:     false,
		OnShutdownError:       nil,
		OnShutdownSuccess:     nil,
	}); err != nil {
		return fmt.Errorf("listening HTTP server: %w", err)
	}
	return nil
}

func (s *Server) setMiddleware(logQuerys bool) {
	// RequestD MW
	s.app.Use(func(ctx fiber.Ctx) error {
		ctx.Request().Header.Set(fiber.HeaderXRequestID, uuid.NewString())
		return ctx.Next()
	})

	// Request Log MW
	if logQuerys {
		s.app.Use(s.logMW)
	}

	// Response content-Type MW
	s.app.Use(func(ctx fiber.Ctx) error {
		ctx.Response().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return ctx.Next() //nolint:wrapcheck
	})

	// Response add RequestID MW
	s.app.Use(func(ctx fiber.Ctx) error {
		ctx.Response().Header.Set(fiber.HeaderXRequestID, string(ctx.Request().Header.Peek(fiber.HeaderXRequestID)))
		return ctx.Next()
	})
}

func (s *Server) setHandlers() {
	rootRoute := s.app.Group("/api-gateway-service")

	handlerV1 := v1.NewHandler(v1.Config{
		JwtKey:          s.jwtKey,
		AuthHandlers:    s.authHandlers,
		DevicesHandlers: s.devicesHandlers,
		TagsHandlers:    s.tagsHandler,
		ReportsHandlers: s.reportsHandler,
	})
	{
		apiV1 := rootRoute.Group("/v1")
		handlerV1.InitRouter(apiV1)
	}
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.app.ShutdownWithContext(ctx)
	if err != nil {
		return fmt.Errorf("s.app.ShutdownWithContext: %w", err)
	}
	return nil
}

type errResp struct {
	Error bool   `json:"error"`
	Data  string `json:"data"`
}

func (s *Server) errorHandler(ctx fiber.Ctx, err error) error {
	requestID := ctx.Get(fiber.HeaderXRequestID)
	statusCode := fiber.StatusInternalServerError

	resp := errResp{
		Error: true,
		Data:  err.Error(),
	}

	var fiberErr *fiber.Error

	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
		resp.Data = fiberErr.Message
	}

	s.log.Error(
		err.Error(),
		zap.String("request_id", requestID),
		zap.String("method", ctx.Method()),
		zap.String("path", ctx.Path()),
		zap.Int("status", statusCode),
	)

	body, err := json.Marshal(resp)
	if err != nil {
		s.log.Error(fmt.Errorf("json.Marshal: %w", err).Error())
	}

	if respondErr := ctx.Status(statusCode).Send(body); respondErr != nil {
		s.log.Error(
			"sending error response",
			zap.String("error", err.Error()),
			zap.String("request_id", requestID),
			zap.String("method", ctx.Method()),
			zap.String("path", ctx.Path()),
			zap.Int("status", statusCode),
		)
	}

	return nil
}

func (s *Server) logMW(ctx fiber.Ctx) error {
	s.log.Info(
		"Request",
		zap.String("req", ctx.Request().String()),
	)
	return ctx.Next() //nolint:wrapcheck
}
