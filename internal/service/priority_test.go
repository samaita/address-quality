// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"testing"

	"address-quality/internal/model"
)

func priorityTestService() *Service {
	return &Service{
		cityPrioritySet: map[string]string{
			"bandung": "KOTA",
			"depok":   "KOTA",
			"karanganyar": "KABUPATEN",
		},
	}
}

// candidate helpers
func ent(id int64, name, level string) model.Entity {
	return model.Entity{ID: id, Name: name, Level: level}
}

func TestApplyCityPriority_SuppressesSubdistrictWhenSoleEvidence(t *testing.T) {
	svc := priorityTestService()
	resolved := []model.ResolvedEvidence{
		{
			Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "bandung"},
			Candidates: []model.Entity{
				ent(1, "Kota Bandung", "CITY"),
				ent(2, "Kabupaten Bandung", "CITY"),
				ent(3, "Bandung", "SUBDISTRICT"),   // Tulungagung subdist
				ent(4, "Bandung", "DISTRICT"),
			},
		},
	}

	svc.applyCityPriority(resolved, nil)

	got := resolved[0].Candidates
	if len(got) != 2 {
		t.Fatalf("expected 2 CITY candidates after suppression, got %d: %+v", len(got), got)
	}
	for _, c := range got {
		if c.Level != "CITY" {
			t.Fatalf("expected only CITY candidates, got %s (%s)", c.Level, c.Name)
		}
	}
	// KOTA preferred first
	if got[0].Name != "Kota Bandung" {
		t.Fatalf("expected Kota Bandung first (KOTA preference), got %s", got[0].Name)
	}
}

func TestApplyCityPriority_KeepsSubdistrictWithOtherPlaceName(t *testing.T) {
	svc := priorityTestService()
	// "Jl. Kenangan, Depok, Yogyakarta" — yogyakarta resolves as place_name
	resolved := []model.ResolvedEvidence{
		{
			Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "depok"},
			Candidates: []model.Entity{
				ent(10, "Kota Depok", "CITY"),
				ent(11, "Depok", "SUBDISTRICT"),
				ent(12, "Depok", "DISTRICT"),
			},
		},
		{
			Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "yogyakarta"},
			Candidates: []model.Entity{
				ent(20, "Kota Yogyakarta", "CITY"),
			},
		},
	}

	svc.applyCityPriority(resolved, nil)

	// depok must keep ALL candidates (no suppression)
	got := resolved[0].Candidates
	if len(got) != 3 {
		t.Fatalf("expected 3 candidates (no suppression), got %d: %+v", len(got), got)
	}
}

func TestApplyCityPriority_KeepsSubdistrictWithPostalCode(t *testing.T) {
	svc := priorityTestService()
	// "Depok 16431" — postal code resolves -> no suppression
	resolved := []model.ResolvedEvidence{
		{
			Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "depok"},
			Candidates: []model.Entity{
				ent(10, "Kota Depok", "CITY"),
				ent(11, "Depok", "SUBDISTRICT"),
			},
		},
		{
			Evidence: model.Evidence{Type: model.EvidencePostalCode, Value: "16431"},
			Candidates: []model.Entity{
				ent(30, "Pancoran Mas", "SUBDISTRICT"),
			},
		},
	}

	svc.applyCityPriority(resolved, nil)

	if len(resolved[0].Candidates) != 2 {
		t.Fatalf("expected no suppression with postal code, got %d", len(resolved[0].Candidates))
	}
}

func TestApplyCityPriority_RoadNameNotEvidence(t *testing.T) {
	svc := priorityTestService()
	// "Jl. Aceh, Bandung" — aceh is road_name, unresolved -> priority applies
	resolved := []model.ResolvedEvidence{
		{
			Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "bandung"},
			Candidates: []model.Entity{
				ent(1, "Kota Bandung", "CITY"),
				ent(3, "Bandung", "SUBDISTRICT"),
			},
		},
		{
			Evidence: model.Evidence{Type: model.EvidenceRoadName, Value: "aceh"},
			Candidates: nil, // road names never resolve
		},
	}

	svc.applyCityPriority(resolved, map[string]bool{"aceh": true})

	// suppression must happen (road name is not evidence)
	got := resolved[0].Candidates
	if len(got) != 1 || got[0].Name != "Kota Bandung" {
		t.Fatalf("expected only Kota Bandung after suppression, got %+v", got)
	}
}

func TestApplyCityPriority_KabupatenOnly(t *testing.T) {
	svc := priorityTestService()
	resolved := []model.ResolvedEvidence{
		{
			Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "karanganyar"},
			Candidates: []model.Entity{
				ent(40, "Kabupaten Karanganyar", "CITY"),
				ent(41, "Karanganyar", "SUBDISTRICT"),
			},
		},
	}

	svc.applyCityPriority(resolved, nil)

	got := resolved[0].Candidates
	if len(got) != 1 || got[0].Name != "Kabupaten Karanganyar" {
		t.Fatalf("expected Kabupaten Karanganyar (KABUPATEN only), got %+v", got)
	}
}

func TestApplyCityPriority_EmptySetNoop(t *testing.T) {
	svc := &Service{} // no priority set
	resolved := []model.ResolvedEvidence{
		{
			Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "bandung"},
			Candidates: []model.Entity{
				ent(1, "Kota Bandung", "CITY"),
				ent(3, "Bandung", "SUBDISTRICT"),
			},
		},
	}

	svc.applyCityPriority(resolved, nil)

	if len(resolved[0].Candidates) != 2 {
		t.Fatalf("expected no change with empty priority set, got %d", len(resolved[0].Candidates))
	}
}

func TestDetectRoadContextTokens(t *testing.T) {
	cases := []struct {
		raw  string
		want map[string]bool
	}{
		{"Jl. Aceh, Bandung", map[string]bool{"aceh": true}},
		{"Jalan Merdeka No.1", map[string]bool{"merdeka": true}},
		{"Gg. Haji", map[string]bool{"haji": true}},
		{"Jl. Gatot Subroto No.86 Bandung", map[string]bool{"gatot": true, "subroto": true}},
		{"Bandung", map[string]bool{}}, // no road prefix
		{"Kecamatan Bandung", map[string]bool{}}, // admin prefix, not road
	}
	for _, tc := range cases {
		got := detectRoadContextTokens(tc.raw)
		if len(got) != len(tc.want) {
			t.Errorf("%q: got %v want %v", tc.raw, got, tc.want)
			continue
		}
		for k := range tc.want {
			if !got[k] {
				t.Errorf("%q: missing %q in %v", tc.raw, k, got)
			}
		}
	}
}
