// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2025 Samaita

package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func APIKeyAuth(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if key == "" {
				return next(c)
			}
			if c.Request().Header.Get("X-API-Key") != key {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid or missing API key",
				})
			}
			return next(c)
		}
	}
}
