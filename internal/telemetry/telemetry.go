// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

// Package telemetry measures server-side work with OpenTelemetry and exposes
// aggregates to Prometheus.
//
// Spans are the single timing source: the SDK stamps start and end on each
// span, and the span-to-metric processor turns every completed span into a
// duration histogram so Prometheus and Grafana can chart time spent per span.
// This package never calls time.Now to time work.
//
// It explains server-side time only. k6 remains authoritative for
// client-observed latency, throughput, and errors.
package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "address-quality/telemetry"

// Fixed span names. Spans are the dimension of app.span.duration_ms, so the
// set is closed: anything else collapses to "other" at export time.
const (
	SpanHTTPRequest         = "http.request"
	SpanSourceCacheReady    = "stage.source_and_cache_ready"
	SpanEvidenceResolution  = "stage.evidence_resolution"
	SpanCandidateBuild      = "stage.candidate_build"
	SpanContextualRecovery  = "stage.contextual_recovery"
	SpanCandidateEvaluation = "stage.candidate_evaluation"
	SpanPostalFallback      = "stage.postal_code_fallback"
)

var knownSpans = map[string]bool{
	SpanHTTPRequest:         true,
	SpanSourceCacheReady:    true,
	SpanEvidenceResolution:  true,
	SpanCandidateBuild:      true,
	SpanContextualRecovery:  true,
	SpanCandidateEvaluation: true,
	SpanPostalFallback:      true,
}

// knownMethods bounds the http.request.method attribute.
var knownMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "DELETE": true,
	"PATCH": true, "HEAD": true, "OPTIONS": true,
}

// knownRoutes bounds the http.route attribute to Echo route templates from
// internal/router. Add new templates here; a growing "other" series is the
// signal that one was missed.
var knownRoutes = map[string]bool{
	"/health":      true,
	"/metrics":     true,
	"/swagger/*":   true,
	"/v0/validate": true,
	"/v1/validate": true,
}

var registry *prometheus.Registry

// Init wires the global OpenTelemetry providers, bridges them to Prometheus,
// and starts Go runtime instrumentation. Call once at startup; the returned
// function shuts the providers down.
func Init() (func(context.Context) error, error) {
	reg := prometheus.NewRegistry()
	exp, err := otelprom.New(
		otelprom.WithRegisterer(reg),
		otelprom.WithoutUnits(),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp))
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(newSpanMetrics(mp.Meter(instrumentationName))),
	)

	otel.SetMeterProvider(mp)
	otel.SetTracerProvider(tp)

	// Go runtime facts (goroutines, heap, GC) as OTel instruments. One second
	// is the sampling floor so a short benchmark phase is not left unobserved.
	if err := runtime.Start(
		runtime.WithMeterProvider(mp),
		runtime.WithMinimumReadMemStatsInterval(time.Second),
	); err != nil {
		return nil, err
	}

	registry = reg

	return func(ctx context.Context) error {
		_ = tp.Shutdown(ctx)
		return mp.Shutdown(ctx)
	}, nil
}

// Start begins a span. Its elapsed time becomes app.span.duration_ms once the
// span ends, so callers only have to End it. The context is deliberately not
// returned: no call site changes behaviour, and request flow stays untouched.
func Start(ctx context.Context, name string, attrs ...attribute.KeyValue) trace.Span {
	opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindInternal)}
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	_, span := otel.Tracer(instrumentationName).Start(ctx, name, opts...)
	return span
}

// Handler serves Prometheus exposition format. Returns 503 until Init runs.
func Handler() http.Handler {
	if registry == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "telemetry not initialized", http.StatusServiceUnavailable)
		})
	}
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// BoundMethod reduces a request method to a fixed set.
func BoundMethod(method string) string {
	if knownMethods[method] {
		return method
	}
	return "OTHER"
}

// BoundRoute reduces a route to a registered Echo template so a raw path,
// query, or address can never become a metric label.
func BoundRoute(route string) string {
	if knownRoutes[route] {
		return route
	}
	return "other"
}

func boundedSpan(name string) string {
	if knownSpans[name] {
		return name
	}
	return "other"
}

// statusClass reduces the span's status-code attribute to a fixed class so
// 4xx and 5xx stay separable without exploding cardinality. Non-HTTP spans
// report "none".
func statusClass(attrs []attribute.KeyValue) string {
	for _, a := range attrs {
		if a.Key != "http.response.status_code" || a.Value.Type() != attribute.INT64 {
			continue
		}
		code := a.Value.AsInt64()
		switch {
		case code >= 200 && code < 300:
			return "2xx"
		case code >= 300 && code < 400:
			return "3xx"
		case code >= 400 && code < 500:
			return "4xx"
		case code >= 500:
			return "5xx"
		}
		return "other"
	}
	return "none"
}

// spanMetrics converts every completed span into a duration histogram. Elapsed
// time comes from the SDK-stamped span times, not from this package.
type spanMetrics struct {
	duration metric.Float64Histogram
}

func newSpanMetrics(m metric.Meter) *spanMetrics {
	d, err := m.Float64Histogram("app.span.duration_ms",
		metric.WithUnit("ms"),
		metric.WithDescription("Elapsed time of completed spans by span name and status"),
	)
	if err != nil {
		// A failed instrument must not break request handling.
		return &spanMetrics{}
	}
	return &spanMetrics{duration: d}
}

func (s *spanMetrics) OnStart(context.Context, sdktrace.ReadWriteSpan) {}

func (s *spanMetrics) OnEnd(r sdktrace.ReadOnlySpan) {
	if s.duration == nil {
		return
	}
	status := "unset"
	switch r.Status().Code {
	case codes.Ok:
		status = "ok"
	case codes.Error:
		status = "error"
	}
	s.duration.Record(context.Background(),
		float64(r.EndTime().Sub(r.StartTime()))/float64(time.Millisecond),
		metric.WithAttributes(
			attribute.String("span.name", boundedSpan(r.Name())),
			attribute.String("span.status", status),
			attribute.String("http.status_class", statusClass(r.Attributes())),
		),
	)
}

func (s *spanMetrics) Shutdown(context.Context) error   { return nil }
func (s *spanMetrics) ForceFlush(context.Context) error { return nil }
