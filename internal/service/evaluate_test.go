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
		name    string
		provID  int64
		cityID  int64
		distID  int64
		subID   int64
		hier    *database.HierarchyMap
		want    bool
	}{
		{
			name:   "valid full chain",
			provID: 1, cityID: 2, distID: 3, subID: 4,
			hier: validHierarchy(),
			want: true,
		},
		{
			name:   "invalid city-province",
			provID: 1, cityID: 99, distID: 0, subID: 0,
			hier: validHierarchy(),
			want: false,
		},
		{
			name:   "invalid district-city",
			provID: 1, cityID: 2, distID: 99, subID: 0,
			hier: validHierarchy(),
			want: false,
		},
		{
			name:   "invalid subdistrict-district",
			provID: 1, cityID: 2, distID: 3, subID: 99,
			hier: validHierarchy(),
			want: false,
		},
		{
			name:   "nil hierarchy no children",
			provID: 1, cityID: 0, distID: 0, subID: 0,
			hier: nil,
			want: true,
		},
		{
			name:   "nil hierarchy with children",
			provID: 1, cityID: 2, distID: 0, subID: 0,
			hier: nil,
			want: false,
		},
		{
			name:   "empty IDs valid",
			provID: 0, cityID: 0, distID: 0, subID: 0,
			hier: validHierarchy(),
			want: true,
		},
		{
			name:   "province only valid",
			provID: 1, cityID: 0, distID: 0, subID: 0,
			hier: validHierarchy(),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &EvaluationContext{
				WinnerProvinceID:    tt.provID,
				WinnerCityID:        tt.cityID,
				WinnerDistrictID:    tt.distID,
				WinnerSubDistrictID: tt.subID,
			}
			evaluateHierarchy(ctx, tt.hier)
			if ctx.HierarchyValid != tt.want {
				t.Errorf("got %v, want %v", ctx.HierarchyValid, tt.want)
			}
		})
	}
}

func TestEvaluateCompleteness(t *testing.T) {
	tests := []struct {
		name          string
		provID        int64
		cityID        int64
		distID        int64
		subID         int64
		postalMatched bool
		inputPostal   string
		wantMatched   []model.Component
		wantMissing   []model.Component
	}{
		{
			name:          "all matched",
			provID: 1, cityID: 2, distID: 3, subID: 4,
			postalMatched: true, inputPostal: "12345",
			wantMatched: []model.Component{
				model.ComponentProvince, model.ComponentCity,
				model.ComponentDistrict, model.ComponentSubDistrict,
				model.ComponentPostalCode,
			},
			wantMissing: nil,
		},
		{
			name:          "province only no postal",
			provID: 1, cityID: 0, distID: 0, subID: 0,
			postalMatched: false, inputPostal: "",
			wantMatched:   []model.Component{model.ComponentProvince},
			wantMissing: []model.Component{
				model.ComponentCity, model.ComponentDistrict, model.ComponentSubDistrict,
			},
		},
		{
			name:          "postal code missing when input provided",
			provID: 1, cityID: 2, distID: 3, subID: 4,
			postalMatched: false, inputPostal: "12345",
			wantMatched: []model.Component{
				model.ComponentProvince, model.ComponentCity,
				model.ComponentDistrict, model.ComponentSubDistrict,
			},
			wantMissing: []model.Component{model.ComponentPostalCode},
		},
		{
			name:          "no province and no postal input",
			provID: 0, cityID: 0, distID: 0, subID: 0,
			postalMatched: false, inputPostal: "",
			wantMatched:   nil,
			wantMissing: []model.Component{
				model.ComponentProvince, model.ComponentCity,
				model.ComponentDistrict, model.ComponentSubDistrict,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &EvaluationContext{
				WinnerProvinceID:    tt.provID,
				WinnerCityID:        tt.cityID,
				WinnerDistrictID:    tt.distID,
				WinnerSubDistrictID: tt.subID,
				PostalCodeMatched:   tt.postalMatched,
				InputPostalCode:     tt.inputPostal,
			}
			evaluateCompleteness(ctx)
			if !componentsEqual(ctx.Matched, tt.wantMatched) {
				t.Errorf("Matched = %v, want %v", ctx.Matched, tt.wantMatched)
			}
			if !componentsEqual(ctx.Missing, tt.wantMissing) {
				t.Errorf("Missing = %v, want %v", ctx.Missing, tt.wantMissing)
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
		name          string
		hierValid     bool
		provID        int64
		cityID        int64
		distID        int64
		subID         int64
		postalMatched bool
		inputPostal   string
		hier          *database.HierarchyMap
		wantCount     int
	}{
		{
			name:      "no conflicts",
			hierValid: true, provID: 1, cityID: 2, subID: 4,
			postalMatched: true, inputPostal: "12345",
			wantCount: 0,
		},
		{
			name:      "hierarchy conflict",
			hierValid: false, provID: 1, cityID: 2, subID: 0,
			postalMatched: false, inputPostal: "",
			wantCount: 1,
		},
		{
			name:      "postal code mismatch",
			hierValid: true, provID: 1, cityID: 2, subID: 4,
			postalMatched: false, inputPostal: "12345",
			wantCount: 1,
		},
		{
			name:      "both conflicts",
			hierValid: false, provID: 1, cityID: 2, subID: 4,
			postalMatched: false, inputPostal: "12345",
			wantCount: 2,
		},
		{
			name:      "no postal conflict when no input postal",
			hierValid: true, provID: 1, cityID: 2, subID: 4,
			postalMatched: false, inputPostal: "",
			wantCount: 0,
		},
		{
			name:      "no postal conflict when no subdistrict",
			hierValid: true, provID: 1, cityID: 2, subID: 0,
			postalMatched: false, inputPostal: "12345",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &EvaluationContext{
				WinnerProvinceID:    tt.provID,
				WinnerCityID:        tt.cityID,
				WinnerDistrictID:    tt.distID,
				WinnerSubDistrictID: tt.subID,
				PostalCodeMatched:   tt.postalMatched,
				InputPostalCode:     tt.inputPostal,
				HierarchyValid:      tt.hierValid,
			}
			detectConflicts(ctx, tt.hier, defaultConflictRules)
			if len(ctx.Conflicts) != tt.wantCount {
				t.Errorf("got %d conflicts, want %d: %v", len(ctx.Conflicts), tt.wantCount, ctx.Conflicts)
			}
		})
	}
}

func TestScoreConfidence(t *testing.T) {
	tests := []struct {
		name          string
		hierValid     bool
		cityID        int64
		provID        int64
		postalMatched bool
		exactMatch    bool
		want          float64
	}{
		{
			name:          "all signals",
			hierValid: true, cityID: 2, provID: 1,
			postalMatched: true, exactMatch: true,
			want: 1.0,
		},
		{
			name:          "exact match only",
			hierValid: false, cityID: 0, provID: 0,
			postalMatched: false, exactMatch: true,
			want: 0.40,
		},
		{
			name:          "hierarchy only",
			hierValid: true, cityID: 2, provID: 0,
			postalMatched: false, exactMatch: false,
			want: 0.30,
		},
		{
			name:          "postal code only",
			hierValid: false, cityID: 0, provID: 0,
			postalMatched: true, exactMatch: false,
			want: 0.20,
		},
		{
			name:          "province only",
			hierValid: false, cityID: 0, provID: 1,
			postalMatched: false, exactMatch: false,
			want: 0.10,
		},
		{
			name:          "no signals",
			hierValid: false, cityID: 0, provID: 0,
			postalMatched: false, exactMatch: false,
			want: 0.0,
		},
		{
			name:          "exact + hierarchy",
			hierValid: true, cityID: 2, provID: 0,
			postalMatched: false, exactMatch: true,
			want: 0.70,
		},
		{
			name:          "exact + province",
			hierValid: false, cityID: 0, provID: 1,
			postalMatched: false, exactMatch: true,
			want: 0.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &EvaluationContext{
				WinnerProvinceID:  tt.provID,
				WinnerCityID:      tt.cityID,
				HierarchyValid:    tt.hierValid,
				PostalCodeMatched: tt.postalMatched,
				ExactMatchFound:   tt.exactMatch,
			}
			got := scoreConfidence(ctx)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssessQuality(t *testing.T) {
	tests := []struct {
		name     string
		conflict bool
		provID   int64
		missing  []model.Component
		want     model.QualityStatus
	}{
		{
			name:     "valid",
			conflict: false, provID: 1,
			missing: nil,
			want:   model.StatusValid,
		},
		{
			name:     "valid with postal code missing",
			conflict: false, provID: 1,
			missing: []model.Component{model.ComponentPostalCode},
			want:   model.StatusValid,
		},
		{
			name:     "conflict",
			conflict: true, provID: 1,
			missing: nil,
			want:   model.StatusConflict,
		},
		{
			name:     "unknown no province",
			conflict: false, provID: 0,
			missing: nil,
			want:   model.StatusUnknown,
		},
		{
			name:     "conflict overrides unknown",
			conflict: true, provID: 0,
			missing: nil,
			want:   model.StatusConflict,
		},
		{
			name:     "incomplete missing city",
			conflict: false, provID: 1,
			missing: []model.Component{model.ComponentCity},
			want:   model.StatusIncomplete,
		},
		{
			name:     "incomplete multiple missing",
			conflict: false, provID: 1,
			missing: []model.Component{model.ComponentCity, model.ComponentDistrict},
			want:   model.StatusIncomplete,
		},
		{
			name:     "incomplete with conflict takes precedence",
			conflict: true, provID: 1,
			missing: []model.Component{model.ComponentCity},
			want:   model.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &EvaluationContext{
				WinnerProvinceID: tt.provID,
				Missing:          tt.missing,
			}
			if tt.conflict {
				ctx.Conflicts = []model.Conflict{{Type: "test"}}
			}
			got := assessQuality(ctx)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateCandidate_FullPipeline(t *testing.T) {
	hier := validHierarchy()

	t.Run("valid address", func(t *testing.T) {
		ctx := &EvaluationContext{
			WinnerProvinceID:    1,
			WinnerCityID:        2,
			WinnerDistrictID:    3,
			WinnerSubDistrictID: 4,
			PostalCodeMatched:   true,
			InputPostalCode:     "12345",
			ExactMatchFound:     true,
		}
		eval := EvaluateCandidate(ctx, hier)
		if eval.Status != model.StatusValid {
			t.Errorf("Status = %v, want VALID", eval.Status)
		}
		if eval.Confidence != 1.0 {
			t.Errorf("Confidence = %v, want 1.0", eval.Confidence)
		}
		if len(eval.Conflicts) != 0 {
			t.Errorf("Conflicts = %v, want none", eval.Conflicts)
		}
	})

	t.Run("conflict address", func(t *testing.T) {
		ctx := &EvaluationContext{
			WinnerProvinceID:    1,
			WinnerCityID:        99,
			WinnerDistrictID:    3,
			WinnerSubDistrictID: 4,
			PostalCodeMatched:   true,
			InputPostalCode:     "12345",
			ExactMatchFound:     false,
		}
		eval := EvaluateCandidate(ctx, hier)
		if eval.Status != model.StatusConflict {
			t.Errorf("Status = %v, want CONFLICT", eval.Status)
		}
		if len(eval.Conflicts) == 0 {
			t.Errorf("expected conflicts, got none")
		}
	})

	t.Run("unknown address", func(t *testing.T) {
		ctx := &EvaluationContext{
			WinnerProvinceID:    0,
			WinnerCityID:        0,
			WinnerDistrictID:    0,
			WinnerSubDistrictID: 0,
			PostalCodeMatched:   false,
			InputPostalCode:     "",
			ExactMatchFound:     false,
		}
		eval := EvaluateCandidate(ctx, hier)
		if eval.Status != model.StatusUnknown {
			t.Errorf("Status = %v, want UNKNOWN", eval.Status)
		}
		if eval.Confidence != 0.0 {
			t.Errorf("Confidence = %v, want 0.0", eval.Confidence)
		}
	})

	t.Run("incomplete address", func(t *testing.T) {
		ctx := &EvaluationContext{
			WinnerProvinceID:    1,
			WinnerCityID:        0,
			WinnerDistrictID:    0,
			WinnerSubDistrictID: 0,
			PostalCodeMatched:   false,
			InputPostalCode:     "",
			ExactMatchFound:     false,
		}
		eval := EvaluateCandidate(ctx, hier)
		if eval.Status != model.StatusIncomplete {
			t.Errorf("Status = %v, want INCOMPLETE", eval.Status)
		}
	})
}

func TestBuildExplainability(t *testing.T) {
	tests := []struct {
		name       string
		exactMatch bool
		hierValid  bool
		want       []string
	}{
		{
			name:       "exact match and hierarchy",
			exactMatch: true, hierValid: true,
			want: []string{"exact_match", "hierarchy_validation"},
		},
		{
			name:       "exact match only",
			exactMatch: true, hierValid: false,
			want: []string{"exact_match"},
		},
		{
			name:       "hierarchy only",
			exactMatch: false, hierValid: true,
			want: []string{"hierarchy_validation"},
		},
		{
			name:       "no reasons",
			exactMatch: false, hierValid: false,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &EvaluationContext{
				ExactMatchFound: tt.exactMatch,
				HierarchyValid:  tt.hierValid,
			}
			got := BuildExplainability(ctx)
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func stringSliceEqual(a, b []string) bool {
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

func TestCustomConflictRule(t *testing.T) {
	ctx := &EvaluationContext{
		WinnerProvinceID:    1,
		WinnerCityID:        2,
		HierarchyValid:      true,
		PostalCodeMatched:   true,
		InputPostalCode:     "12345",
		WinnerSubDistrictID: 4,
	}

	rule := &customRule{}
	detectConflicts(ctx, validHierarchy(), []ConflictRule{rule})
	if len(ctx.Conflicts) != 1 {
		t.Fatalf("expected 1 custom conflict, got %d", len(ctx.Conflicts))
	}
	if ctx.Conflicts[0].Type != "custom" {
		t.Errorf("expected type 'custom', got %q", ctx.Conflicts[0].Type)
	}
}

type customRule struct{}

func (customRule) Evaluate(_ *EvaluationContext, _ *database.HierarchyMap) *model.Conflict {
	return &model.Conflict{
		Type:    "custom",
		Message: "custom rule triggered",
	}
}
