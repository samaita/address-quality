// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package router

import (
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoswagger "github.com/swaggo/echo-swagger"

	"address-quality/internal/config"
	"address-quality/internal/handler"
	"address-quality/internal/logger"
	"address-quality/internal/metrics"
	mw "address-quality/internal/middleware"

	_ "address-quality/docs"
)

func Setup(h *handler.Handler, cfg *config.Config) *echo.Echo {
	e := echo.New()

	e.Logger = logger.NewEchoLogger()

	e.Server.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second

	e.Use(logger.EchoMiddleware())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderContentType,
			"X-API-Key",
			echo.HeaderAuthorization,
		},
		MaxAge: 3600,
	}))

	e.GET("/health", h.HandleHealthCheck)
	e.GET("/swagger/*", echoswagger.WrapHandler)

	api := e.Group("")
	api.Use(middleware.BodyLimit(cfg.MaxBodySize))
	api.Use(mw.APIKeyAuth(cfg.APIKey))
	api.Use(mw.RateLimiter(cfg.RateLimit, cfg.RateWindow))
	api.Use(mw.RequestID())
	api.POST("/v0/validate", h.HandleGeocodeRequest)
	api.POST("/v1/validate", h.HandleAddressRequest)

	// Benchmark-only application/runtime snapshot. Registered only when
	// explicitly enabled and only reachable from loopback, so it never leaks
	// to public callers. Proxy headers are refused as well: behind a reverse
	// proxy on localhost every client would otherwise look loopback.
	if cfg.EnablePerfSnapshot {
		e.GET("/internal/perf-snapshot", func(c echo.Context) error {
			req := c.Request()
			if req.Header.Get(echo.HeaderXForwardedFor) != "" || req.Header.Get(echo.HeaderXRealIP) != "" {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			if !isLoopback(req.RemoteAddr) {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			return c.JSON(http.StatusOK, metrics.Collect())
		})
	}

	return e
}

func isLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
