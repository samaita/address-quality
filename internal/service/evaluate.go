// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"math"

	"address-quality/internal/database"
	"address-quality/internal/model"
)

type EvaluationContext struct {
	WinnerProvinceID    int64
	WinnerCityID        int64
	WinnerDistrictID    int64
	WinnerSubDistrictID int64
	PostalCodeMatched   bool
	InputPostalCode     string
	ExactMatchFound     bool

	HierarchyValid bool
	Matched        []model.Component
	Missing        []model.Component
	Conflicts      []model.Conflict
}

type ConflictRule interface {
	Evaluate(ctx *EvaluationContext, hierarchy *database.HierarchyMap) *model.Conflict
}

type hierarchyRule struct{}

func (hierarchyRule) Evaluate(ctx *EvaluationContext, hierarchy *database.HierarchyMap) *model.Conflict {
	if ctx.WinnerCityID > 0 && !ctx.HierarchyValid {
		return &model.Conflict{
			Type:    "hierarchy_conflict",
			Message: "administrative hierarchy is invalid",
		}
	}
	return nil
}

type postalCodeRule struct{}

func (postalCodeRule) Evaluate(ctx *EvaluationContext, _ *database.HierarchyMap) *model.Conflict {
	if ctx.InputPostalCode != "" && ctx.WinnerSubDistrictID > 0 && !ctx.PostalCodeMatched {
		return &model.Conflict{
			Type:    "postal_code_mismatch",
			Message: "postal code does not match the resolved sub-district",
		}
	}
	return nil
}

var defaultConflictRules = []ConflictRule{
	hierarchyRule{},
	postalCodeRule{},
}

func EvaluateCandidate(ctx *EvaluationContext, hierarchy *database.HierarchyMap) model.CandidateEvaluation {
	evaluateHierarchy(ctx, hierarchy)
	evaluateCompleteness(ctx)
	detectConflicts(ctx, hierarchy, defaultConflictRules)
	confidence := scoreConfidence(ctx)
	status := assessQuality(ctx)

	return model.CandidateEvaluation{
		Confidence: confidence,
		Status:     status,
		Matched:    ctx.Matched,
		Missing:    ctx.Missing,
		Conflicts:  ctx.Conflicts,
	}
}

func evaluateHierarchy(ctx *EvaluationContext, hierarchy *database.HierarchyMap) {
	if hierarchy == nil {
		ctx.HierarchyValid = ctx.WinnerCityID == 0 && ctx.WinnerDistrictID == 0 && ctx.WinnerSubDistrictID == 0
		return
	}
	if ctx.WinnerCityID > 0 {
		if hierarchy.CityToProvince[ctx.WinnerCityID] != ctx.WinnerProvinceID {
			ctx.HierarchyValid = false
			return
		}
	}
	if ctx.WinnerDistrictID > 0 {
		if hierarchy.DistrictToCity[ctx.WinnerDistrictID] != ctx.WinnerCityID {
			ctx.HierarchyValid = false
			return
		}
	}
	if ctx.WinnerSubDistrictID > 0 {
		if hierarchy.SubDistrictToDist[ctx.WinnerSubDistrictID] != ctx.WinnerDistrictID {
			ctx.HierarchyValid = false
			return
		}
	}
	ctx.HierarchyValid = true
}

func evaluateCompleteness(ctx *EvaluationContext) {
	ctx.Matched = nil
	ctx.Missing = nil

	check := func(id int64, comp model.Component) {
		if id > 0 {
			ctx.Matched = append(ctx.Matched, comp)
		} else {
			ctx.Missing = append(ctx.Missing, comp)
		}
	}

	check(ctx.WinnerProvinceID, model.ComponentProvince)
	check(ctx.WinnerCityID, model.ComponentCity)
	check(ctx.WinnerDistrictID, model.ComponentDistrict)
	check(ctx.WinnerSubDistrictID, model.ComponentSubDistrict)

	if ctx.InputPostalCode != "" {
		if ctx.PostalCodeMatched {
			ctx.Matched = append(ctx.Matched, model.ComponentPostalCode)
		} else {
			ctx.Missing = append(ctx.Missing, model.ComponentPostalCode)
		}
	}
}

func detectConflicts(ctx *EvaluationContext, hierarchy *database.HierarchyMap, rules []ConflictRule) {
	ctx.Conflicts = nil
	for _, rule := range rules {
		if c := rule.Evaluate(ctx, hierarchy); c != nil {
			ctx.Conflicts = append(ctx.Conflicts, *c)
		}
	}
}

func scoreConfidence(ctx *EvaluationContext) float64 {
	var score float64

	if ctx.ExactMatchFound {
		score += WeightExactMatch
	}

	if ctx.HierarchyValid && ctx.WinnerCityID > 0 {
		score += WeightHierarchy
	}

	if ctx.PostalCodeMatched {
		score += WeightPostalCode
	}

	if ctx.WinnerProvinceID > 0 {
		score += WeightProvince
	}

	if score > 1.0 {
		score = 1.0
	}
	return math.Round(score*10000) / 10000
}

func assessQuality(ctx *EvaluationContext) model.QualityStatus {
	if len(ctx.Conflicts) > 0 {
		return model.StatusConflict
	}

	if ctx.WinnerProvinceID == 0 {
		return model.StatusUnknown
	}

	for _, m := range ctx.Missing {
		if m != model.ComponentPostalCode {
			return model.StatusIncomplete
		}
	}

	return model.StatusValid
}

func BuildExplainability(ctx *EvaluationContext) []string {
	var reasons []string

	if ctx.ExactMatchFound {
		reasons = append(reasons, "exact_match")
	}

	if ctx.HierarchyValid {
		reasons = append(reasons, "hierarchy_validation")
	}

	return reasons
}
