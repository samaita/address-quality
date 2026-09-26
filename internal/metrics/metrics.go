// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

// Package metrics collects bounded in-process performance measurements.
//
// It is diagnostic context for controlled benchmark runs, not a client
// performance authority: k6 owns client-observed latency and throughput.
// Aggregates are cumulative so a runner can snapshot before and after a
// measured phase and compute deltas. No external exporter is required and
// nothing here is exposed to public callers.
package metrics

import (
	"runtime"
	"sort"
	"sync"
	"time"
)

// SchemaVersion identifies the snapshot layout for artifact assembly.
const SchemaVersion = "1.0"

// Stage is a fixed validation stage name. Cardinality is bounded to these values.
type Stage string

// Stable V1 stage boundaries. Do not add per-function or per-candidate stages.
const (
	StageSourceAndCacheReady Stage = "source_and_cache_ready"
	StageEvidenceResolution  Stage = "evidence_resolution"
	StageCandidateBuild      Stage = "candidate_build"
	StageContextualRecovery  Stage = "contextual_recovery"
	StageCandidateEvaluation Stage = "candidate_evaluation"
	StagePostalCodeFallback  Stage = "postal_code_fallback"
)

var knownStages = map[Stage]bool{
	StageSourceAndCacheReady: true,
	StageEvidenceResolution:  true,
	StageCandidateBuild:      true,
	StageContextualRecovery:  true,
	StageCandidateEvaluation: true,
	StagePostalCodeFallback:  true,
}

// DurationBucketsMs are inclusive upper bounds in milliseconds for the fixed
// duration histograms. The implicit overflow bucket holds everything above the
// last bound. Bounds are shared by every aggregate in a snapshot.
var DurationBucketsMs = []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000}

type durationAgg struct {
	count   uint64
	sumMs   float64
	buckets []uint64 // len(DurationBucketsMs)+1; last index is the overflow bucket
}

func newDurationAgg() *durationAgg {
	return &durationAgg{buckets: make([]uint64, len(DurationBucketsMs)+1)}
}

func (a *durationAgg) observe(d time.Duration) {
	ms := float64(d) / float64(time.Millisecond)
	a.count++
	a.sumMs += ms
	a.buckets[sort.SearchFloat64s(DurationBucketsMs, ms)]++
}

// DurationSnapshot is a cumulative duration aggregate in milliseconds. A zero
// count means no observation was recorded; consumers must not read that as a
// measured zero duration.
type DurationSnapshot struct {
	Count        uint64   `json:"count"`
	SumMs        float64  `json:"sum_ms"`
	BucketCounts []uint64 `json:"bucket_counts"`
}

// RuntimeSnapshot describes captured Go runtime facts at one instant.
type RuntimeSnapshot struct {
	GoVersion       string `json:"go_version"`
	Goroutines      int    `json:"goroutines"`
	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"`
	HeapInuseBytes  uint64 `json:"heap_inuse_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	GCCycles        uint32 `json:"gc_cycles"`
	GCPauseNs       uint64 `json:"gc_pause_ns"`
	NumCPU          int    `json:"cpu_count"`
	GOOS            string `json:"os"`
	GOARCH          string `json:"arch"`
}

// HTTPRequestSnapshot is one bounded (method, route, status) aggregate.
type HTTPRequestSnapshot struct {
	Method     string           `json:"method"`
	Route      string           `json:"route"`
	Status     int              `json:"status"`
	DurationMs DurationSnapshot `json:"duration_ms"`
}

// HTTPSnapshot is the request-level aggregate at the Echo middleware boundary.
// 4xx and 5xx counts stay separable and are derived from the same outcomes.
type HTTPSnapshot struct {
	RequestsTotal  uint64                `json:"requests_total"`
	DurationMs     DurationSnapshot      `json:"duration_ms"`
	Errors4xxTotal uint64                `json:"errors_4xx_total"`
	Errors5xxTotal uint64                `json:"errors_5xx_total"`
	ByRoute        []HTTPRequestSnapshot `json:"by_route"`
}

// Snapshot is the whole in-process measurement state at one instant.
type Snapshot struct {
	SchemaVersion     string                      `json:"schema_version"`
	StartedAt         time.Time                   `json:"started_at"`
	SnapshotAt        time.Time                   `json:"snapshot_at"`
	ProcessUptimeSecs float64                     `json:"process_uptime_seconds"`
	DurationBucketsMs []float64                   `json:"duration_buckets_ms"`
	HTTP              HTTPSnapshot                `json:"http"`
	ValidationStages  map[string]DurationSnapshot `json:"validation_stages"`
	Runtime           RuntimeSnapshot             `json:"runtime"`
}

type requestKey struct {
	method string
	route  string
	status int
}

type collector struct {
	mu       sync.Mutex
	started  time.Time
	requests map[requestKey]*durationAgg
	stages   map[Stage]*durationAgg
}

var defaultCollector = &collector{
	started:  time.Now(),
	requests: map[requestKey]*durationAgg{},
	stages:   map[Stage]*durationAgg{},
}

// knownMethods bounds the method dimension. Unknown methods collapse to OTHER.
var knownMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "DELETE": true,
	"PATCH": true, "HEAD": true, "OPTIONS": true,
}

// knownRoutes bounds the route dimension to Echo route templates registered in
// internal/router. Anything else collapses to "other" so a raw path, query, or
// address can never become a metric label. Add new templates here when routes
// are added; a growing "other" series is the signal that one was missed.
var knownRoutes = map[string]bool{
	"/health":                 true,
	"/swagger/*":              true,
	"/v0/validate":            true,
	"/v1/validate":            true,
	"/internal/perf-snapshot": true,
}

func boundedMethod(method string) string {
	if knownMethods[method] {
		return method
	}
	return "OTHER"
}

func boundedRoute(route string) string {
	if knownRoutes[route] {
		return route
	}
	return "other"
}

// ObserveHTTP records one server-side request. Route must be an Echo route
// template (bounded), never a raw URI with parameters.
func ObserveHTTP(method, route string, status int, d time.Duration) {
	key := requestKey{method: boundedMethod(method), route: boundedRoute(route), status: status}
	c := defaultCollector
	c.mu.Lock()
	agg, ok := c.requests[key]
	if !ok {
		agg = newDurationAgg()
		c.requests[key] = agg
	}
	agg.observe(d)
	c.mu.Unlock()
}

// ObserveStage records one completed validation stage. Unknown stages are
// ignored so metric cardinality cannot grow past the fixed stage enum.
func ObserveStage(stage Stage, d time.Duration) {
	if !knownStages[stage] {
		return
	}
	c := defaultCollector
	c.mu.Lock()
	agg, ok := c.stages[stage]
	if !ok {
		agg = newDurationAgg()
		c.stages[stage] = agg
	}
	agg.observe(d)
	c.mu.Unlock()
}

func (a *durationAgg) snapshot() DurationSnapshot {
	buckets := make([]uint64, len(a.buckets))
	copy(buckets, a.buckets)
	return DurationSnapshot{Count: a.count, SumMs: a.sumMs, BucketCounts: buckets}
}

// Collect returns a copy of the cumulative aggregates plus one Go runtime
// sample. Values absent from the snapshot were not measured and must not be
// filled in as zero by consumers.
func Collect() Snapshot {
	c := defaultCollector
	now := time.Now()

	c.mu.Lock()
	out := Snapshot{
		SchemaVersion:     SchemaVersion,
		StartedAt:         c.started.UTC(),
		SnapshotAt:        now.UTC(),
		ProcessUptimeSecs: now.Sub(c.started).Seconds(),
		DurationBucketsMs: append([]float64(nil), DurationBucketsMs...),
		ValidationStages:  map[string]DurationSnapshot{},
	}
	total := newDurationAgg()
	keys := make([]requestKey, 0, len(c.requests))
	for k := range c.requests {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].route != keys[j].route {
			return keys[i].route < keys[j].route
		}
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].status < keys[j].status
	})
	for _, k := range keys {
		agg := c.requests[k]
		out.HTTP.ByRoute = append(out.HTTP.ByRoute, HTTPRequestSnapshot{
			Method:     k.method,
			Route:      k.route,
			Status:     k.status,
			DurationMs: agg.snapshot(),
		})
		total.count += agg.count
		total.sumMs += agg.sumMs
		for i, b := range agg.buckets {
			total.buckets[i] += b
		}
		switch {
		case k.status >= 500:
			out.HTTP.Errors5xxTotal += agg.count
		case k.status >= 400:
			out.HTTP.Errors4xxTotal += agg.count
		}
	}
	out.HTTP.RequestsTotal = total.count
	out.HTTP.DurationMs = total.snapshot()
	for stage, agg := range c.stages {
		out.ValidationStages[string(stage)] = agg.snapshot()
	}
	c.mu.Unlock()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	out.Runtime = RuntimeSnapshot{
		GoVersion:       runtime.Version(),
		Goroutines:      runtime.NumGoroutine(),
		HeapAllocBytes:  ms.HeapAlloc,
		HeapInuseBytes:  ms.HeapInuse,
		TotalAllocBytes: ms.TotalAlloc,
		GCCycles:        ms.NumGC,
		GCPauseNs:       ms.PauseTotalNs,
		NumCPU:          runtime.NumCPU(),
		GOOS:            runtime.GOOS,
		GOARCH:          runtime.GOARCH,
	}
	return out
}
