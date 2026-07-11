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

	repo, err := database.New("address.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	s := sanitizer.New(sanitizer.DefaultPolicy())
	svc := service.New(repo, s, cfg.MaxAddressLength)
	h := handler.New(svc)

	e := router.Setup(h, cfg)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("server starting on %s", addr)
	e.Logger.Fatal(e.Start(addr))
}
