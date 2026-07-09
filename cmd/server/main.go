package main

import (
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"address-quality/internal/config"
	"address-quality/internal/database"
	"address-quality/internal/handler"
	mw "address-quality/internal/middleware"
)

func main() {
	config.Init()

	database.Init("address.db")

	e := echo.New()

	e.Server.ReadTimeout = time.Duration(config.Cfg.ReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(config.Cfg.WriteTimeout) * time.Second

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", handler.HandleHealthCheck)

	api := e.Group("")
	api.Use(mw.RateLimiter(config.Cfg.RateLimit, config.Cfg.RateWindow))
	api.POST("/v1/validate", handler.HandleAddressRequest)

	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	log.Printf("server starting on %s", addr)
	e.Logger.Fatal(e.Start(addr))
}
