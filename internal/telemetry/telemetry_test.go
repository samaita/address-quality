// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package telemetry

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func setup(t *testing.T) {
	t.Helper()
	shutdown, err := Init()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })
}

func scrape(t *testing.T) string {
	t.Helper()
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if rec.Code != 200 {
		t.Fatalf("metrics status = %d", rec.Code)
	}
	return rec.Body.String()
}

func TestSpanBecomesPrometheusMetric(t *testing.T) {
	setup(t)
	_, span := Start(context.Background(), SpanCandidateBuild)
	time.Sleep(2 * time.Millisecond)
	span.End()

	body := scrape(t)
	if !strings.Contains(body, "app_span_duration_ms") {
		t.Fatalf("span duration not exported:\n%s", body)
	}
	if !strings.Contains(body, "stage.candidate_build") {
		t.Fatalf("span name not attributed:\n%s", body)
	}
}

func TestDimensionsAreBounded(t *testing.T) {
	if got := BoundMethod("weird-method"); got != "OTHER" {
		t.Fatalf("method %q", got)
	}
	if got := BoundMethod("GET"); got != "GET" {
		t.Fatalf("method %q", got)
	}
	if got := BoundRoute("/raw/12345"); got != "other" {
		t.Fatalf("route %q", got)
	}
	if got := BoundRoute(""); got != "other" {
		t.Fatalf("route %q", got)
	}
	if got := BoundRoute("/v1/validate"); got != "/v1/validate" {
		t.Fatalf("route %q", got)
	}
	if got := boundedSpan("raw_address_label_stage"); got != "other" {
		t.Fatalf("span %q", got)
	}
	if got := boundedSpan(SpanEvidenceResolution); got != SpanEvidenceResolution {
		t.Fatalf("span %q", got)
	}
}

func TestStatusClass(t *testing.T) {
	for code, want := range map[int64]string{
		200: "2xx", 204: "2xx", 301: "3xx",
		404: "4xx", 429: "4xx", 500: "5xx", 503: "5xx",
	} {
		got := statusClass([]attribute.KeyValue{attribute.Int64("http.response.status_code", code)})
		if got != want {
			t.Fatalf("status %d class = %q, want %q", code, got, want)
		}
	}
	if got := statusClass(nil); got != "none" {
		t.Fatalf("non-HTTP span class = %q", got)
	}
	// A raw address must never ride in as a status code.
	if got := statusClass([]attribute.KeyValue{attribute.String("http.response.status_code", "Jl. Aceh")}); got != "none" {
		t.Fatalf("untyped attr class = %q", got)
	}
}

func TestNoLabelLeakage(t *testing.T) {
	setup(t)
	secret := "Jl. Gatot Subroto No.86 Bandung"

	// Both as the span name and as an extra attribute: the exporter must
	// carry neither, so no caller can smuggle an address into a label.
	_, span := Start(context.Background(), secret,
		attribute.String("http.route", secret),
		attribute.String("raw_address", secret),
	)
	span.End()

	if body := scrape(t); strings.Contains(body, "Gatot") {
		t.Fatalf("exporter leaked raw input:\n%s", body)
	}
}

func TestConcurrentSpans(t *testing.T) {
	setup(t)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, s := Start(context.Background(), SpanCandidateEvaluation)
				s.End()
			}
		}()
	}
	wg.Wait()

	if body := scrape(t); !strings.Contains(body, "stage.candidate_evaluation") {
		t.Fatalf("concurrent spans missing:\n%s", body)
	}
}

type endedSpan struct {
	name   string
	parent trace.SpanContext
}

// recorder captures completed spans so parentage can be asserted without a
// trace backend.
type recorder struct {
	mu    sync.Mutex
	spans []endedSpan
}

func (r *recorder) OnStart(context.Context, sdktrace.ReadWriteSpan) {}
func (r *recorder) OnEnd(s sdktrace.ReadOnlySpan) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spans = append(r.spans, endedSpan{name: s.Name(), parent: s.Parent()})
}
func (*recorder) Shutdown(context.Context) error   { return nil }
func (*recorder) ForceFlush(context.Context) error { return nil }

func TestContextCarriesSpanAndNests(t *testing.T) {
	rec := &recorder{}
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	t.Cleanup(func() { _, _ = Init() })

	ctx, parent := Start(context.Background(), SpanHTTPRequest)
	parentSC := parent.SpanContext()
	if !trace.SpanFromContext(ctx).SpanContext().IsValid() {
		t.Fatal("context does not carry the span")
	}
	_, child := Start(ctx, SpanEvidenceResolution)
	child.End()
	parent.End()

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.spans) != 2 {
		t.Fatalf("ended spans = %d, want 2", len(rec.spans))
	}
	childRec, parentRec := rec.spans[0], rec.spans[1]
	if childRec.name != SpanEvidenceResolution {
		t.Fatalf("child name = %q", childRec.name)
	}
	if childRec.parent.TraceID() != parentSC.TraceID() || childRec.parent.SpanID() != parentSC.SpanID() {
		t.Fatalf("child parent = %v, want trace %s span %s", childRec.parent, parentSC.TraceID(), parentSC.SpanID())
	}
	if parentRec.parent.IsValid() {
		t.Fatalf("http.request span should be a root, got parent %v", parentRec.parent)
	}
}
