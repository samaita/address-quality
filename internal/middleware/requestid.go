// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := context.WithValue(req.Context(), requestIDKey, uuid.Must(uuid.NewV7()).String())
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
