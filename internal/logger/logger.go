// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"

	mw "address-quality/internal/middleware"
)

var L zerolog.Logger

func Init(level string) {
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339

	L = zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()

	zerolog.DefaultContextLogger = &L
}

func Info() *zerolog.Event  { return L.Info() }
func Debug() *zerolog.Event { return L.Debug() }
func Warn() *zerolog.Event  { return L.Warn() }
func Error() *zerolog.Event { return L.Error() }
func Fatal() *zerolog.Event { return L.Fatal() }

type EchoLogger struct{}

func NewEchoLogger() EchoLogger { return EchoLogger{} }

func (l EchoLogger) Output() io.Writer                         { return os.Stdout }
func (l EchoLogger) SetOutput(w io.Writer)                     {}
func (l EchoLogger) Prefix() string                            { return "" }
func (l EchoLogger) SetPrefix(p string)                        {}
func (l EchoLogger) Level() log.Lvl                            { return 0 }
func (l EchoLogger) SetLevel(v log.Lvl)                        {}
func (l EchoLogger) SetHeader(h string)                        {}
func (l EchoLogger) Print(i ...interface{})                    { L.Info().Msg(fmt.Sprint(i...)) }
func (l EchoLogger) Printf(format string, args ...interface{}) { L.Info().Msgf(format, args...) }
func (l EchoLogger) Printj(j log.JSON)                         { L.Info().Interface("json", j).Msg("") }
func (l EchoLogger) Debug(i ...interface{})                    { L.Debug().Msg(fmt.Sprint(i...)) }
func (l EchoLogger) Debugf(format string, args ...interface{}) { L.Debug().Msgf(format, args...) }
func (l EchoLogger) Debugj(j log.JSON)                         { L.Debug().Interface("json", j).Msg("") }
func (l EchoLogger) Info(i ...interface{})                     { L.Info().Msg(fmt.Sprint(i...)) }
func (l EchoLogger) Infof(format string, args ...interface{})  { L.Info().Msgf(format, args...) }
func (l EchoLogger) Infoj(j log.JSON)                          { L.Info().Interface("json", j).Msg("") }
func (l EchoLogger) Warn(i ...interface{})                     { L.Warn().Msg(fmt.Sprint(i...)) }
func (l EchoLogger) Warnf(format string, args ...interface{})  { L.Warn().Msgf(format, args...) }
func (l EchoLogger) Warnj(j log.JSON)                          { L.Warn().Interface("json", j).Msg("") }
func (l EchoLogger) Error(i ...interface{})                    { L.Error().Msg(fmt.Sprint(i...)) }
func (l EchoLogger) Errorf(format string, args ...interface{}) { L.Error().Msgf(format, args...) }
func (l EchoLogger) Errorj(j log.JSON)                         { L.Error().Interface("json", j).Msg("") }
func (l EchoLogger) Fatal(i ...interface{})                    { L.Fatal().Msg(fmt.Sprint(i...)) }
func (l EchoLogger) Fatalf(format string, args ...interface{}) { L.Fatal().Msgf(format, args...) }
func (l EchoLogger) Fatalj(j log.JSON)                         { L.Fatal().Interface("json", j).Msg("") }
func (l EchoLogger) Panic(i ...interface{})                    { L.Panic().Msg(fmt.Sprint(i...)) }
func (l EchoLogger) Panicf(format string, args ...interface{}) { L.Panic().Msgf(format, args...) }
func (l EchoLogger) Panicj(j log.JSON)                         { L.Panic().Interface("json", j).Msg("") }

func EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Now()

			req := c.Request()
			res := c.Response()
			reqID := mw.GetRequestID(req.Context())

			event := L.Info()
			if res.Status >= 500 {
				event = L.Error()
			} else if res.Status >= 400 {
				event = L.Warn()
			}

			event.
				Str("request_id", reqID).
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("remote_ip", c.RealIP()).
				Str("host", req.Host).
				Int("status", res.Status).
				Int64("bytes_in", req.ContentLength).
				Int64("bytes_out", res.Size).
				Dur("latency", stop.Sub(start)).
				Msg("")

			return err
		}
	}
}
