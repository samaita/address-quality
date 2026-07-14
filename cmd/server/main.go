package main

import (
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

	s := sanitizer.New(sanitizer.DefaultPolicy())
	svc := service.New(repo, locationRepo, s, cfg.MaxAddressLength, cfg.LocationSourceCode)
	h := handler.New(svc)

	e := router.Setup(h, cfg)

	logger.Info().Int("port", cfg.Port).Msg("server starting")
	if err := e.StartServer(e.Server); err != nil {
		logger.Fatal().Err(err).Msg("server failed to start")
	}
}
