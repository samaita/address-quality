package main

import (
	"fmt"
	"log"

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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(mw.RateLimiter(config.Cfg.RateLimit, config.Cfg.RateWindow))

	e.POST("/v1/address", handler.HandleAddressRequest)

	addr := fmt.Sprintf(":%d", config.Cfg.Port)
	log.Printf("server starting on %s", addr)
	e.Logger.Fatal(e.Start(addr))
}
