// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package metrics

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func stageCount(s Snapshot, stage Stage) uint64 {
	return s.ValidationStages[string(stage)].Count
}

func TestUnknownStageIsIgnored(t *testing.T) {
	before := Collect()
	ObserveStage(Stage("raw_address_label_stage"), 3*time.Millisecond)
	after := Collect()

	if _, ok := after.ValidationStages["raw_address_label_stage"]; ok {
		t.Fatal("unknown stage must not create a metric series")
	}
	if stageCount(after, StageCandidateBuild) != stageCount(before, StageCandidateBuild) {
		t.Fatal("unknown stage must not touch known series")
	}
}

func TestNoLabelLeakage(t *testing.T) {
	secret := "Jl. Gatot Subroto No.86 Bandung"
	ObserveStage(Stage(secret), time.Millisecond)
	ObserveHTTP("GET", secret, 200, time.Millisecond)
	ObserveStage(StagePostalCodeFallback, time.Millisecond)

	out, err := json.Marshal(Collect())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), secret) {
		t.Fatalf("snapshot leaked raw input: %s", out)
	}
}

func TestBoundedMethodDimension(t *testing.T) {
	before := Collect()
	ObserveHTTP("weird-method", "/v1/validate", 200, time.Millisecond)
	ObserveHTTP("GET", "/raw/12345", 200, time.Millisecond)
	after := Collect()

	for _, r := range after.HTTP.ByRoute {
		if r.Method == "weird-method" {
			t.Fatal("unbounded method must not become a metric dimension")
		}
		if r.Route == "/raw/12345" {
			t.Fatal("unbounded route must not become a metric dimension")
		}
	}
	if after.HTTP.RequestsTotal != before.HTTP.RequestsTotal+2 {
		t.Fatalf("requests_total = %d, want %d", after.HTTP.RequestsTotal, before.HTTP.RequestsTotal+2)
	}

	ObserveHTTP("POST", "/v1/validate", 200, time.Millisecond)
	for _, r := range Collect().HTTP.ByRoute {
		if r.Route == "/v1/validate" {
			return
		}
	}
	t.Fatal("registered route template must survive bounding")
}

func TestStatusClassAndHistogram(t *testing.T) {
	ObserveHTTP("POST", "/v1/validate", 400, 3*time.Millisecond)
	ObserveHTTP("POST", "/v1/validate", 503, 3*time.Millisecond)
	ObserveStage(StageCandidateBuild, 3*time.Millisecond)

	s := Collect()
	if s.HTTP.Errors4xxTotal == 0 {
		t.Fatal("4xx count not separable")
	}
	if s.HTTP.Errors5xxTotal == 0 {
		t.Fatal("5xx count not separable")
	}
	agg := s.ValidationStages[string(StageCandidateBuild)]
	if agg.Count == 0 {
		t.Fatal("stage not recorded")
	}
	idx := -1
	for i, b := range DurationBucketsMs {
		if b == 5 {
			idx = i
		}
	}
	if idx < 0 || len(agg.BucketCounts) != len(DurationBucketsMs)+1 {
		t.Fatal("unexpected bucket layout")
	}
	if agg.BucketCounts[idx] == 0 {
		t.Fatal("3ms must fall into the 5ms inclusive upper bound")
	}
	if agg.SumMs <= 0 {
		t.Fatal("sum must reflect the observation")
	}
}

func TestSnapshotIsIndependent(t *testing.T) {
	ObserveStage(StageEvidenceResolution, 2*time.Millisecond)
	first := Collect()
	firstCount := stageCount(first, StageEvidenceResolution)
	ObserveStage(StageEvidenceResolution, 2*time.Millisecond)

	if stageCount(first, StageEvidenceResolution) != firstCount {
		t.Fatal("snapshot must be a copy, not a live view")
	}
	if stageCount(Collect(), StageEvidenceResolution) != firstCount+1 {
		t.Fatal("later observation must be visible in a fresh snapshot")
	}
}

func TestConcurrentObserve(t *testing.T) {
	before := Collect()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				ObserveHTTP("POST", "/v1/validate", 200, time.Millisecond)
				ObserveStage(StageCandidateEvaluation, time.Millisecond)
				Collect()
			}
		}()
	}
	wg.Wait()
	after := Collect()

	if after.HTTP.RequestsTotal != before.HTTP.RequestsTotal+1000 {
		t.Fatalf("requests_total = %d, want %d", after.HTTP.RequestsTotal, before.HTTP.RequestsTotal+1000)
	}
	if stageCount(after, StageCandidateEvaluation) != stageCount(before, StageCandidateEvaluation)+1000 {
		t.Fatal("concurrent stage observations were lost")
	}
}
