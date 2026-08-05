// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package config

import (
	"strings"

	"github.com/spf13/viper"

	"address-quality/internal/logger"
)

type Config struct {
	Port               int
	APIKey             string
	RateLimit          int
	RateWindow         int
	ReadTimeout        int
	WriteTimeout       int
	MaxBodySize        string
	MaxAddressLength   int
	AddressDBPath      string
	LocationDBPath     string
	DBMaxOpenConns     int
	LocationSourceCode string
	LogLevel           string
	AllowedOrigins     []string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("PORT", 7300)
	viper.SetDefault("API_KEY", "")
	viper.SetDefault("RATE_LIMIT", 100)
	viper.SetDefault("RATE_WINDOW", 60)
	viper.SetDefault("READ_TIMEOUT", 5)
	viper.SetDefault("WRITE_TIMEOUT", 10)
	viper.SetDefault("MAX_BODY_SIZE", "1M")
	viper.SetDefault("MAX_ADDRESS_LENGTH", 1000)
	viper.SetDefault("ADDRESS_DB_PATH", "db/address.db")
	viper.SetDefault("LOCATION_DB_PATH", "db/location.db")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 10)
	viper.SetDefault("LOCATION_SOURCE_CODE", "kemendagri")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "https://samaita.com")

	if err := viper.ReadInConfig(); err != nil {
		logger.Warn().Err(err).Msg(".env not found, using defaults")
	}

	cfg := &Config{
		Port:               viper.GetInt("PORT"),
		APIKey:             viper.GetString("API_KEY"),
		RateLimit:          viper.GetInt("RATE_LIMIT"),
		RateWindow:         viper.GetInt("RATE_WINDOW"),
		ReadTimeout:        viper.GetInt("READ_TIMEOUT"),
		WriteTimeout:       viper.GetInt("WRITE_TIMEOUT"),
		MaxBodySize:        viper.GetString("MAX_BODY_SIZE"),
		MaxAddressLength:   viper.GetInt("MAX_ADDRESS_LENGTH"),
		AddressDBPath:      viper.GetString("ADDRESS_DB_PATH"),
		LocationDBPath:     viper.GetString("LOCATION_DB_PATH"),
		DBMaxOpenConns:     viper.GetInt("DB_MAX_OPEN_CONNS"),
		LocationSourceCode: viper.GetString("LOCATION_SOURCE_CODE"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		AllowedOrigins:     strings.Split(viper.GetString("CORS_ALLOWED_ORIGINS"), ","),
	}

	logger.Info().
		Int("port", cfg.Port).
		Int("rate_limit", cfg.RateLimit).
		Int("rate_window", cfg.RateWindow).
		Msg("config loaded")
	return cfg
}
