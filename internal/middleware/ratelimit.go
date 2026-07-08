package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	ratelimit "github.com/samaita/go-http-ratelimit"
)

func RateLimiter(rate int, windowSec int) echo.MiddlewareFunc {
	rl := ratelimit.New(ratelimit.Config{
		Rate:   rate,
		Window: time.Duration(windowSec) * time.Second,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","retry_after":60}`))
		},
	})

	return echo.WrapMiddleware(rl.Middleware)
}
