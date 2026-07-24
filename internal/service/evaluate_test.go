// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"testing"

	"address-quality/internal/database"
	"address-quality/internal/model"
)

func validHierarchy() *database.HierarchyMap {
	return &database.HierarchyMap{
		CityToProvince:    map[int64]int64{2: 1},
		DistrictToCity:    map[int64]int64{3: 2},
		SubDistrictToDist: map[int64]int64{4: 3},
	}
}

func TestEvaluateHierarchy(t *testing.T) {
	tests := []struct {
		name   string
		provID int64
		cityID int64
		distID int64
		subID  int64
		hier   *database.HierarchyMap
		want   int
	}{
		{
			name:   "valid full chain",
			provID: 1, cityID: 2, distID: 3, subID: 4,
			hier: validHierarchy(),
			want: 0,
		},
		{
			name:   "invalid city-province",
			provID: 1, cityID: 99, distID: 0, subID: 0,
			hier: validHierarchy(),
			want: 1,
		},
		{
			name:   "invalid district-city",
			provID: 1, cityID: 2, distID: 99, subID: 0,
			hier: validHierarchy(),
			want: 1,
		},
		{
			name:   "invalid subdistrict-district",
			provID: 1, cityID: 2, distID: 3, subID: 99,
			hier: validHierarchy(),
			want: 1,
		},
		{
			name:   "empty IDs valid",
			provID: 0, cityID: 0, distID: 0, subID: 0,
			hier: validHierarchy(),
			want: 0,
		},
		{
			name:   "province only valid",
			provID: 1, cityID: 0, distID: 0, subID: 0,
			hier: validHierarchy(),
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{
				Location: model.AdminLocation{},
			}
			if tt.provID > 0 {
				candidate.Location.Province = &model.Province{ID: tt.provID}
			}
			if tt.cityID > 0 {
				candidate.Location.City = &model.City{ID: tt.cityID}
			}
			if tt.distID > 0 {
				candidate.Location.District = &model.District{ID: tt.distID}
			}
			if tt.subID > 0 {
				candidate.Location.SubDistrict = &model.SubDistrict{ID: tt.subID}
			}

			evaluateHierarchy(candidate, tt.hier)
			if len(candidate.Location.Conflicts) != tt.want {
				t.Errorf("got %d conflicts, want %d: %v", len(candidate.Location.Conflicts), tt.want, candidate.Location.Conflicts)
			}
		})
	}
}

func TestEvaluateCompleteness(t *testing.T) {
	tests := []struct {
		name        string
		provID      int64
		cityID      int64
		distID      int64
		subID       int64
		wantMatched []model.Component
		wantMissing []model.Component
	}{
		{
			name:        "all matched",
			provID: 1, cityID: 2, distID: 3, subID: 4,
			wantMatched: []model.Component{
				model.ComponentProvince, model.ComponentCity,
				model.ComponentDistrict, model.ComponentSubDistrict,
			},
			wantMissing: nil,
		},
		{
			name:        "province only",
			provID: 1, cityID: 0, distID: 0, subID: 0,
			wantMatched: []model.Component{model.ComponentProvince},
			wantMissing: []model.Component{
				model.ComponentCity, model.ComponentDistrict, model.ComponentSubDistrict,
			},
		},
		{
			name:        "no province and no postal input",
			provID: 0, cityID: 0, distID: 0, subID: 0,
			wantMatched: nil,
			wantMissing: []model.Component{
				model.ComponentProvince, model.ComponentCity,
				model.ComponentDistrict, model.ComponentSubDistrict,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{
				Location: model.AdminLocation{},
			}
			if tt.provID > 0 {
				candidate.Location.Province = &model.Province{ID: tt.provID, Name: "Province"}
			}
			if tt.cityID > 0 {
				candidate.Location.City = &model.City{ID: tt.cityID, Name: "City"}
			}
			if tt.distID > 0 {
				candidate.Location.District = &model.District{ID: tt.distID, Name: "District"}
			}
			if tt.subID > 0 {
				candidate.Location.SubDistrict = &model.SubDistrict{ID: tt.subID, Name: "SubDistrict"}
			}

			matched := getMatchedComponents(candidate)
			missing := getMissingComponents(candidate)

			if !componentsEqual(matched, tt.wantMatched) {
				t.Errorf("Matched = %v, want %v", matched, tt.wantMatched)
			}
			if !componentsEqual(missing, tt.wantMissing) {
				t.Errorf("Missing = %v, want %v", missing, tt.wantMissing)
			}
		})
	}
}

func componentsEqual(a, b []model.Component) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestDetectConflicts(t *testing.T) {
	tests := []struct {
		name      string
		provID    int64
		cityID    int64
		distID    int64
		subID     int64
		postalSub bool
		evidence  []model.MatchedEvidence
		wantCount int
	}{
		{
			name:      "no conflicts",
			provID: 1, cityID: 2, distID: 3, subID: 4,
			postalSub: true,
			wantCount: 0,
		},
		{
			name:      "orphan city",
			provID: 0, cityID: 2, distID: 1, subID: 0,
			postalSub: false,
			wantCount: 1,
		},
		{
			name:      "no postal conflict when no input postal",
			provID: 1, cityID: 2, subID: 0,
			postalSub: false,
			wantCount: 0,
		},
		{
			name:      "multiple cities",
			provID: 1, cityID: 2, distID: 0, subID: 0,
			evidence: []model.MatchedEvidence{
				{Evidence: model.Evidence{Value: "a"}, Resolved: &model.Entity{ID: 2, Level: "CITY"}},
				{Evidence: model.Evidence{Value: "b"}, Resolved: &model.Entity{ID: 99, Level: "CITY"}},
			},
			wantCount: 1,
		},
		{
			name:      "duplicate level",
			provID: 1, cityID: 2, distID: 0, subID: 0,
			evidence: []model.MatchedEvidence{
				{Evidence: model.Evidence{Value: "a"}, Resolved: &model.Entity{ID: 99, Level: "PROVINCE"}},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{
				Location: model.AdminLocation{},
				Evidence: tt.evidence,
			}
			if tt.provID > 0 {
				candidate.Location.Province = &model.Province{ID: tt.provID}
			}
			if tt.cityID > 0 {
				candidate.Location.City = &model.City{ID: tt.cityID}
			}
			if tt.distID > 0 {
				candidate.Location.District = &model.District{ID: tt.distID}
			}
			if tt.subID > 0 {
				candidate.Location.SubDistrict = &model.SubDistrict{ID: tt.subID}
			}
			if tt.postalSub {
				candidate.Location.PostalCode = &model.PostalCode{ID: tt.subID}
			}

			detectConflicts(candidate, nil)

			if len(candidate.Location.Conflicts) != tt.wantCount {
				t.Errorf("got %d conflicts, want %d: %v", len(candidate.Location.Conflicts), tt.wantCount, candidate.Location.Conflicts)
			}
		})
	}
}

func TestScoreConfidence(t *testing.T) {
	tests := []struct {
		name     string
		provID   int64
		cityID   int64
		postalID int64
		evidence []model.MatchedEvidence
		conflicts []model.Conflict
		want     float64
	}{
		{
			name:     "all signals",
			provID: 1, cityID: 2, postalID: 3,
			evidence: []model.MatchedEvidence{{Resolved: &model.Entity{ID: 1}}},
			conflicts: nil,
			want:     0.30,
		},
		{
			name:     "exact match only",
			provID: 0, cityID: 0, postalID: 0,
			evidence: []model.MatchedEvidence{{Resolved: &model.Entity{ID: 1}}},
			want:     0.10,
		},
		{
			name:     "hierarchy only (city without conflict)",
			provID: 0, cityID: 2, postalID: 0,
			evidence: nil,
			want:     0.15,
		},
		{
			name:     "postal code only",
			provID: 0, cityID: 0, postalID: 3,
			evidence: nil,
			want:     0.05,
		},
		{
			name:     "province only",
			provID: 1, cityID: 0, postalID: 0,
			evidence: nil,
			want:     0.0,
		},
		{
			name:     "no signals",
			provID: 0, cityID: 0, postalID: 0,
			evidence: nil,
			want:     0.0,
		},
		{
			name:     "exact + hierarchy",
			provID: 0, cityID: 2, postalID: 0,
			evidence: []model.MatchedEvidence{{Resolved: &model.Entity{ID: 1}}},
			want:     0.25,
		},
		{
			name:     "exact + province",
			provID: 1, cityID: 0, postalID: 0,
			evidence: []model.MatchedEvidence{{Resolved: &model.Entity{ID: 1}}},
			want:     0.10,
		},
		{
			name:     "hierarchy conflict blocks hierarchy weight",
			provID: 1, cityID: 2, postalID: 0,
			evidence: nil,
			conflicts: []model.Conflict{{Type: "hierarchy_conflict"}},
			want:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{
				Location: model.AdminLocation{},
				Evidence: tt.evidence,
			}
			if tt.provID > 0 {
				candidate.Location.Province = &model.Province{ID: tt.provID}
			}
			if tt.cityID > 0 {
				candidate.Location.City = &model.City{ID: tt.cityID}
			}
			if tt.postalID > 0 {
				candidate.Location.PostalCode = &model.PostalCode{ID: tt.postalID}
			}
			candidate.Location.Conflicts = tt.conflicts

			got := scoreConfidence(candidate)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssessQuality(t *testing.T) {
	tests := []struct {
		name       string
		provID     int64
		cityID     int64
		distID     int64
		subID      int64
		conflicts  []model.Conflict
		want       model.QualityStatus
	}{
		{
			name:   "valid",
			provID: 1, cityID: 2, distID: 3,
			want:   model.StatusValid,
		},
		{
			name:   "conflict",
			provID: 1, cityID: 2, distID: 3,
			conflicts: []model.Conflict{{Type: "hierarchy_conflict"}},
			want:     model.StatusConflict,
		},
		{
			name:   "unknown truly empty",
			provID: 0, cityID: 0, distID: 0, subID: 0,
			want: model.StatusUnknown,
		},
		{
			name:   "conflict overrides unknown",
			provID: 0, cityID: 0, distID: 0, subID: 0,
			conflicts: []model.Conflict{{Type: "hierarchy_conflict"}},
			want:     model.StatusConflict,
		},
		{
			name:   "incomplete missing city",
			provID: 1, cityID: 0, distID: 0,
			want:   model.StatusIncomplete,
		},
		{
			name:   "incomplete missing city and district",
			provID: 1, cityID: 0, distID: 0, subID: 4,
			want:   model.StatusIncomplete,
		},
		{
			name:   "incomplete has subdistrict but no province",
			provID: 0, cityID: 0, distID: 0, subID: 4,
			want:   model.StatusIncomplete,
		},
		{
			name:   "incomplete has city but no province",
			provID: 0, cityID: 2, distID: 3,
			want:   model.StatusIncomplete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{
				Location: model.AdminLocation{},
			}
			if tt.provID > 0 {
				candidate.Location.Province = &model.Province{ID: tt.provID}
			}
			if tt.cityID > 0 {
				candidate.Location.City = &model.City{ID: tt.cityID}
			}
			if tt.distID > 0 {
				candidate.Location.District = &model.District{ID: tt.distID}
			}
			if tt.subID > 0 {
				candidate.Location.SubDistrict = &model.SubDistrict{ID: tt.subID}
			}
			candidate.Location.Conflicts = tt.conflicts

			got := assessQuality(candidate)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateCandidate_FullPipeline(t *testing.T) {
	hier := validHierarchy()

	t.Run("valid address", func(t *testing.T) {
		candidate := &model.AdminCandidate{
			Location: model.AdminLocation{
				Province:    &model.Province{ID: 1, Name: "Province"},
				City:        &model.City{ID: 2, Name: "City"},
				District:    &model.District{ID: 3, Name: "District"},
				SubDistrict: &model.SubDistrict{ID: 4, Name: "SubDistrict"},
			},
			Evidence: []model.MatchedEvidence{
				{Evidence: model.Evidence{Value: "bandung"}, Resolved: &model.Entity{ID: 2, Level: "CITY"}},
			},
		}
		eval := EvaluateCandidate(candidate, hier, []model.Evidence{{Value: "bandung"}})
		if eval.Status != model.StatusValid {
			t.Errorf("Status = %v, want VALID", eval.Status)
		}
		if eval.Confidence != 0.37 {
			t.Errorf("Confidence = %v, want 0.37", eval.Confidence)
		}
	})

	t.Run("conflict address", func(t *testing.T) {
		candidate := &model.AdminCandidate{
			Location: model.AdminLocation{
				Province:    &model.Province{ID: 1, Name: "Province"},
				City:        &model.City{ID: 99, Name: "WrongCity"},
				District:    &model.District{ID: 3, Name: "District"},
				SubDistrict: &model.SubDistrict{ID: 4, Name: "SubDistrict"},
			},
		}
		eval := EvaluateCandidate(candidate, hier, nil)
		if eval.Status != model.StatusConflict {
			t.Errorf("Status = %v, want CONFLICT", eval.Status)
		}
		if len(eval.Conflicts) == 0 {
			t.Errorf("expected conflicts, got none")
		}
	})

	t.Run("unknown address", func(t *testing.T) {
		candidate := &model.AdminCandidate{Location: model.AdminLocation{}}
		eval := EvaluateCandidate(candidate, hier, nil)
		if eval.Status != model.StatusUnknown {
			t.Errorf("Status = %v, want UNKNOWN", eval.Status)
		}
		if eval.Confidence != 0.0 {
			t.Errorf("Confidence = %v, want 0.0", eval.Confidence)
		}
	})

	t.Run("incomplete address", func(t *testing.T) {
		candidate := &model.AdminCandidate{
			Location: model.AdminLocation{
				Province: &model.Province{ID: 1, Name: "Province"},
			},
		}
		eval := EvaluateCandidate(candidate, hier, nil)
		if eval.Status != model.StatusIncomplete {
			t.Errorf("Status = %v, want INCOMPLETE", eval.Status)
		}
	})
}

func TestBuildReasons(t *testing.T) {
	tests := []struct {
		name       string
		evidence   []model.MatchedEvidence
		conflicts  []model.Conflict
		strategies []model.DiscoveryStrategy
		wantLen    int
	}{
		{
			name:       "exact match and hierarchy",
			evidence:   []model.MatchedEvidence{{Resolved: &model.Entity{ID: 1}}},
			conflicts:  nil,
			wantLen:    2,
		},
		{
			name:       "exact match only",
			evidence:   []model.MatchedEvidence{{Resolved: &model.Entity{ID: 1}}},
			conflicts:  []model.Conflict{{Type: "hierarchy_conflict"}},
			wantLen:    1,
		},
		{
			name:       "with strategies",
			evidence:   []model.MatchedEvidence{{Resolved: &model.Entity{ID: 1}}},
			conflicts:  nil,
			strategies: []model.DiscoveryStrategy{model.DiscoveryTopDown, model.DiscoveryAnyLevel},
			wantLen:    4,
		},
		{
			name:       "no reasons",
			evidence:   nil,
			conflicts:  []model.Conflict{{Type: "hierarchy_conflict"}},
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{
				Location:            model.AdminLocation{Conflicts: tt.conflicts},
				Evidence:            tt.evidence,
				DiscoveryStrategies: tt.strategies,
			}
			got := BuildReasons(candidate)
			if len(got) != tt.wantLen {
				t.Errorf("got %d reasons, want %d: %v", len(got), tt.wantLen, got)
			}
		})
	}
}

func TestEvidenceCoverageReturnsUnused(t *testing.T) {
	allEvidence := []model.Evidence{
		{Value: "bandung"},
		{Value: "citarum"},
	}

	candidate := &model.AdminCandidate{
		Location: model.AdminLocation{
			Province: &model.Province{ID: 1},
			City:     &model.City{ID: 2},
		},
		Evidence: []model.MatchedEvidence{
			{Evidence: model.Evidence{Value: "bandung"}, Resolved: &model.Entity{ID: 2}},
		},
	}

	eval := EvaluateCandidate(candidate, validHierarchy(), allEvidence)

	if len(eval.UnusedEvidence) != 1 {
				t.Errorf("expected 1 unused evidence, got %d: %v", len(eval.UnusedEvidence), eval.UnusedEvidence)
	}
	if len(eval.UnusedEvidence) > 0 && eval.UnusedEvidence[0].Value != "citarum" {
		t.Errorf("expected unused 'citarum', got %v", eval.UnusedEvidence[0].Value)
	}
}

func TestMultipleCitiesConflict(t *testing.T) {
	candidate := &model.AdminCandidate{
		Location: model.AdminLocation{
			Province: &model.Province{ID: 1},
			City:     &model.City{ID: 2},
		},
		Evidence: []model.MatchedEvidence{
			{Evidence: model.Evidence{Value: "a"}, Resolved: &model.Entity{ID: 2, Level: "CITY", Name: "CityA"}},
			{Evidence: model.Evidence{Value: "b"}, Resolved: &model.Entity{ID: 99, Level: "CITY", Name: "CityB"}},
		},
	}

	detectMultipleCities(candidate)

	if len(candidate.Location.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d: %v", len(candidate.Location.Conflicts), candidate.Location.Conflicts)
	}
	if len(candidate.Location.Conflicts) > 0 && candidate.Location.Conflicts[0].Type != "multiple_city" {
		t.Errorf("expected type 'multiple_city', got %q", candidate.Location.Conflicts[0].Type)
	}
}

func TestDuplicateLevelConflict(t *testing.T) {
	candidate := &model.AdminCandidate{
		Location: model.AdminLocation{
			Province: &model.Province{ID: 1},
		},
		Evidence: []model.MatchedEvidence{
			{Evidence: model.Evidence{Value: "a"}, Resolved: &model.Entity{ID: 99, Level: "PROVINCE", Name: "OtherProvince"}},
		},
	}

	detectDuplicateLevel(candidate)

	if len(candidate.Location.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d: %v", len(candidate.Location.Conflicts), candidate.Location.Conflicts)
	}
	if len(candidate.Location.Conflicts) > 0 && candidate.Location.Conflicts[0].Type != "duplicate_level" {
		t.Errorf("expected type 'duplicate_level', got %q", candidate.Location.Conflicts[0].Type)
	}
}

func TestMultiEvidenceBonus(t *testing.T) {
	tests := []struct {
		name     string
		evidence []model.MatchedEvidence
		want     float64
	}{
		{
			name: "single city evidence no bonus",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
			},
			want: 0,
		},
		{
			name: "two city evidence same entity",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
			},
			want: 0.15,
		},
		{
			name: "three city evidence same entity capped",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
			},
			want: 0.15,
		},
		{
			name: "province + city both redundant",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 1, Level: "PROVINCE"}},
				{Resolved: &model.Entity{ID: 1, Level: "PROVINCE"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
			},
			want: 0.35,
		},
		{
			name: "district redundant",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 789, Level: "DISTRICT"}},
				{Resolved: &model.Entity{ID: 789, Level: "DISTRICT"}},
			},
			want: 0.10,
		},
		{
			name: "subdistrict redundant",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 456, Level: "SUBDISTRICT"}},
				{Resolved: &model.Entity{ID: 456, Level: "SUBDISTRICT"}},
			},
			want: 0.05,
		},
		{
			name: "all levels redundant capped at max",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 1, Level: "PROVINCE"}},
				{Resolved: &model.Entity{ID: 1, Level: "PROVINCE"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 789, Level: "DISTRICT"}},
				{Resolved: &model.Entity{ID: 789, Level: "DISTRICT"}},
			},
			want: 0.40,
		},
		{
			name: "different cities no bonus",
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 456, Level: "CITY"}},
			},
			want: 0,
		},
		{
			name: "no resolved evidence",
			evidence: []model.MatchedEvidence{
				{Resolved: nil},
				{Resolved: nil},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{Evidence: tt.evidence}
			got := multiEvidenceBonus(candidate)
			if got != tt.want {
				t.Errorf("multiEvidenceBonus = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScoreConfidence_MultiEvidenceIntegration(t *testing.T) {
	tests := []struct {
		name      string
		provID    int64
		cityID    int64
		postalID  int64
		evidence  []model.MatchedEvidence
		conflicts []model.Conflict
		want      float64
	}{
		{
			name:   "exact + hierarchy + province + city redundancy",
			provID: 1, cityID: 123, postalID: 0,
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
				{Resolved: &model.Entity{ID: 123, Level: "CITY"}},
			},
			want: 0.52,
		},
		{
			name:   "province redundancy with other signals",
			provID: 1, cityID: 123, postalID: 0,
			evidence: []model.MatchedEvidence{
				{Resolved: &model.Entity{ID: 1, Level: "PROVINCE"}},
				{Resolved: &model.Entity{ID: 1, Level: "PROVINCE"}},
			},
			want: 0.60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &model.AdminCandidate{
				Location: model.AdminLocation{},
				Evidence: tt.evidence,
			}
			if tt.provID > 0 {
				candidate.Location.Province = &model.Province{ID: tt.provID}
			}
			if tt.cityID > 0 {
				candidate.Location.City = &model.City{ID: tt.cityID}
			}
			if tt.postalID > 0 {
				candidate.Location.PostalCode = &model.PostalCode{ID: tt.postalID}
			}
			candidate.Location.Conflicts = tt.conflicts

			got := scoreConfidence(candidate)
			if got != tt.want {
				t.Errorf("scoreConfidence = %v, want %v", got, tt.want)
			}
		})
	}
}
