package router

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"address-quality/internal/config"
	"address-quality/internal/handler"
	mw "address-quality/internal/middleware"
)

func Setup(h *handler.Handler, cfg *config.Config) *echo.Echo {
	e := echo.New()

	e.Server.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", h.HandleHealthCheck)

	api := e.Group("")
	api.Use(middleware.BodyLimit(cfg.MaxBodySize))
	api.Use(mw.APIKeyAuth(cfg.APIKey))
	api.Use(mw.RateLimiter(cfg.RateLimit, cfg.RateWindow))
	api.POST("/v1/validate", h.HandleAddressRequest)

	return e
}
