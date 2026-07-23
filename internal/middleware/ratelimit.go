// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	ratelimit "github.com/samaita/go-http-ratelimit"
)

func RateLimiter(rate int, windowSec int) echo.MiddlewareFunc {
	window := time.Duration(windowSec) * time.Second
	rl := ratelimit.New(ratelimit.Config{
		KeyFunc: func(r *http.Request) string {
			ip := r.RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}
			return ip
		},
		Rules: []ratelimit.Rule{
			{Key: "* /v1/*", Rate: rate, Window: window},
		},
		DefaultRate:   rate,
		DefaultWindow: window,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(windowSec))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","retry_after":` + strconv.Itoa(windowSec) + `}`))
		},
	})

	return echo.WrapMiddleware(rl.Middleware)
}
