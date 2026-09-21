// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

// @title           Address Quality API
// @version         1.0
// @description     Thai address validation and resolution API
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey ApiKeyAuth
// @in                         header
// @name                       X-API-Key

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"address-quality/internal/config"
	"address-quality/internal/database"
	"address-quality/internal/handler"
	"address-quality/internal/logger"
	"address-quality/internal/router"
	"address-quality/internal/sanitizer"
	"address-quality/internal/service"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.LogLevel)

	repo, err := database.New(cfg.AddressDBPath, cfg.DBMaxOpenConns)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize address database")
	}

	locationRepo, err := database.NewLocationDB(cfg.LocationDBPath, cfg.DBMaxOpenConns)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize location database")
	}

	// Optional second connection; the app runs on SQLite alone when Postgres
	// is not configured or unreachable.
	postgresRepo, err := database.NewPostgresDB(cfg.PostgresDSN, cfg.DBMaxOpenConns)
	if err != nil {
		logger.Warn().Err(err).Msg("postgres unavailable, continuing without it")
	}
	// Avoid the typed-nil-in-interface trap: only set the interface when a real
	// connection exists.
	var postgresSvc service.PostgresRepository
	if postgresRepo != nil {
		postgresSvc = postgresRepo
	}

	s := sanitizer.New(sanitizer.DefaultPolicy())
	svc := service.New(repo, locationRepo, postgresSvc, s, cfg.MaxAddressLength, cfg.LocationSourceCode, cfg.EnableStoreRequest, cfg.GoogleMapsAPIMock, cfg.GoogleMapsAPIKey, cfg.GoogleMapsBaseURL)
	h := handler.New(svc)

	e := router.Setup(h, cfg)

	e.Server.Addr = fmt.Sprintf(":%d", cfg.Port)
	logger.Info().Int("port", cfg.Port).Msg("server starting")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.StartServer(e.Server); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("http server shutdown failed")
	}
	if err := svc.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("store queue drain failed")
	}
}
