package main

import (
	"fmt"
	"log"

	"address-quality/internal/config"
	"address-quality/internal/database"
	"address-quality/internal/handler"
	"address-quality/internal/router"
	"address-quality/internal/sanitizer"
	"address-quality/internal/service"
)

func main() {
	cfg := config.Load()

	repo, err := database.New(cfg.AddressDBPath, cfg.DBMaxOpenConns)
	if err != nil {
		log.Fatalf("failed to initialize address database: %v", err)
	}

	locationRepo, err := database.NewLocationDB(cfg.LocationDBPath, cfg.DBMaxOpenConns)
	if err != nil {
		log.Fatalf("failed to initialize location database: %v", err)
	}

	s := sanitizer.New(sanitizer.DefaultPolicy())
	svc := service.New(repo, locationRepo, s, cfg.MaxAddressLength, cfg.LocationSourceCode)
	h := handler.New(svc)

	e := router.Setup(h, cfg)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("server starting on %s", addr)
	e.Logger.Fatal(e.Start(addr))
}
