// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"math"

	"address-quality/internal/database"
	"address-quality/internal/model"
)

func EvaluateCandidate(candidate *model.AdminCandidate, hierarchy *database.HierarchyMap, allEvidence []model.Evidence) model.CandidateEvaluation {
	evaluateHierarchy(candidate, hierarchy)
	evaluateCompleteness(candidate)
	unused := evaluateEvidenceCoverage(candidate, allEvidence)
	detectConflicts(candidate, hierarchy)
	confidence := scoreConfidence(candidate)
	status := assessQuality(candidate)

	return model.CandidateEvaluation{
		Candidate:      *candidate,
		Confidence:     confidence,
		Status:         status,
		Matched:        getMatchedComponents(candidate),
		Missing:        getMissingComponents(candidate),
		UnusedEvidence: unused,
		Conflicts:      extractConflicts(candidate),
	}
}

func getMatchedComponents(candidate *model.AdminCandidate) []model.Component {
	var matched []model.Component
	if candidate.Location.Province != nil {
		matched = append(matched, model.ComponentProvince)
	}
	if candidate.Location.City != nil {
		matched = append(matched, model.ComponentCity)
	}
	if candidate.Location.District != nil {
		matched = append(matched, model.ComponentDistrict)
	}
	if candidate.Location.SubDistrict != nil {
		matched = append(matched, model.ComponentSubDistrict)
	}
	if candidate.Location.PostalCode != nil {
		matched = append(matched, model.ComponentPostalCode)
	}
	return matched
}

func getMissingComponents(candidate *model.AdminCandidate) []model.Component {
	var missing []model.Component
	if candidate.Location.Province == nil {
		missing = append(missing, model.ComponentProvince)
	}
	if candidate.Location.City == nil {
		missing = append(missing, model.ComponentCity)
	}
	if candidate.Location.District == nil {
		missing = append(missing, model.ComponentDistrict)
	}
	if candidate.Location.SubDistrict == nil {
		missing = append(missing, model.ComponentSubDistrict)
	}
	return missing
}

func extractConflicts(candidate *model.AdminCandidate) []model.Conflict {
	return candidate.Location.Conflicts
}

func evaluateHierarchy(candidate *model.AdminCandidate, hierarchy *database.HierarchyMap) {
	if hierarchy == nil {
		return
	}

	loc := &candidate.Location
	if loc.City != nil && loc.Province != nil {
		if hierarchy.CityToProvince[loc.City.ID] != loc.Province.ID {
			addConflict(candidate, "hierarchy_conflict", "city does not belong to province")
			return
		}
	}
	if loc.District != nil && loc.City != nil {
		if hierarchy.DistrictToCity[loc.District.ID] != loc.City.ID {
			addConflict(candidate, "hierarchy_conflict", "district does not belong to city")
			return
		}
	}
	if loc.SubDistrict != nil && loc.District != nil {
		if hierarchy.SubDistrictToDist[loc.SubDistrict.ID] != loc.District.ID {
			addConflict(candidate, "hierarchy_conflict", "sub-district does not belong to district")
			return
		}
	}
}

func addConflict(candidate *model.AdminCandidate, conflictType, message string) {
	candidate.Location.Conflicts = append(candidate.Location.Conflicts, model.Conflict{
		Type:    conflictType,
		Message: message,
	})
}

func evaluateCompleteness(candidate *model.AdminCandidate) {
}

func evaluateEvidenceCoverage(candidate *model.AdminCandidate, allEvidence []model.Evidence) []model.Evidence {
	var unused []model.Evidence
	if len(allEvidence) == 0 {
		return unused
	}

	matchedValues := make(map[string]bool)
	for _, me := range candidate.Evidence {
		matchedValues[me.Value] = true
	}

	for _, ev := range allEvidence {
		if !matchedValues[ev.Value] {
			unused = append(unused, ev)
		}
	}
	return unused
}

func detectConflicts(candidate *model.AdminCandidate, hierarchy *database.HierarchyMap) {
	loc := &candidate.Location

	if loc.PostalCode != nil && loc.SubDistrict != nil {
		if loc.PostalCode.ID != loc.SubDistrict.ID {
			addConflict(candidate, "postal_code_mismatch", "postal code does not match sub-district")
		}
	}

	levelCount := 0
	if loc.Province != nil {
		levelCount++
	}
	if loc.City != nil {
		levelCount++
	}
	if loc.District != nil {
		levelCount++
	}
	if loc.SubDistrict != nil {
		levelCount++
	}
	if levelCount > 1 && loc.City != nil {
		if loc.Province == nil {
			addConflict(candidate, "orphan_city", "city resolved without province")
		}
	}

	detectMultipleCities(candidate)
	detectDuplicateLevel(candidate)
}

func detectMultipleCities(candidate *model.AdminCandidate) {
	seen := make(map[int64]bool)
	for _, me := range candidate.Evidence {
		if me.Resolved == nil || me.Resolved.Level != "CITY" {
			continue
		}
		if seen[me.Resolved.ID] {
			continue
		}
		if candidate.Location.City != nil && me.Resolved.ID != candidate.Location.City.ID {
			addConflict(candidate, "multiple_city", "multiple city candidates: "+me.Resolved.Name)
		}
		seen[me.Resolved.ID] = true
	}
}

func detectDuplicateLevel(candidate *model.AdminCandidate) {
	levels := make(map[string]int64)
	if candidate.Location.Province != nil {
		levels["PROVINCE"] = candidate.Location.Province.ID
	}
	if candidate.Location.District != nil {
		levels["DISTRICT"] = candidate.Location.District.ID
	}
	if candidate.Location.SubDistrict != nil {
		levels["SUBDISTRICT"] = candidate.Location.SubDistrict.ID
	}

	for _, me := range candidate.Evidence {
		if me.Resolved == nil {
			continue
		}
		if me.Resolved.Level == "CITY" {
			continue
		}
		if existingID, ok := levels[me.Resolved.Level]; ok {
			if existingID != me.Resolved.ID {
				addConflict(candidate, "duplicate_level", "duplicate "+me.Resolved.Level+": "+me.Resolved.Name)
			}
		}
	}
}

func scoreConfidence(candidate *model.AdminCandidate) float64 {
	var score float64

	hasExactMatch := false
	for _, me := range candidate.Evidence {
		if me.Resolved != nil {
			hasExactMatch = true
			break
		}
	}
	if hasExactMatch {
		score += WeightExactMatch
	}

	hasHierarchyConflict := false
	for _, c := range candidate.Location.Conflicts {
		if c.Type == "hierarchy_conflict" {
			hasHierarchyConflict = true
			break
		}
	}

	if !hasHierarchyConflict && candidate.Location.City != nil {
		score += WeightHierarchy
	}

	if candidate.Location.PostalCode != nil {
		score += WeightPostalCode
	}

	if candidate.Location.Province != nil {
		score += WeightProvince
	}

	if score > 1.0 {
		score = 1.0
	}
	return math.Round(score*10000) / 10000
}

func assessQuality(candidate *model.AdminCandidate) model.QualityStatus {
	if len(candidate.Location.Conflicts) > 0 {
		return model.StatusConflict
	}

	hasAny := candidate.Location.Province != nil ||
		candidate.Location.City != nil ||
		candidate.Location.District != nil ||
		candidate.Location.SubDistrict != nil

	if !hasAny {
		return model.StatusUnknown
	}

	if candidate.Location.Province == nil ||
		candidate.Location.City == nil ||
		candidate.Location.District == nil {
		return model.StatusIncomplete
	}

	return model.StatusValid
}

func BuildReasons(candidate *model.AdminCandidate) []model.Reason {
	var reasons []model.Reason

	hasExactMatch := false
	for _, me := range candidate.Evidence {
		if me.Resolved != nil {
			hasExactMatch = true
			break
		}
	}
	if hasExactMatch {
		reasons = append(reasons, model.Reason("exact_match"))
	}

	hasHierarchyConflict := false
	for _, c := range candidate.Location.Conflicts {
		if c.Type == "hierarchy_conflict" {
			hasHierarchyConflict = true
			break
		}
	}
	if !hasHierarchyConflict {
		reasons = append(reasons, model.Reason("hierarchy_validation"))
	}

	for _, strategy := range candidate.DiscoveryStrategies {
		reasons = append(reasons, model.Reason("strategy_"+string(strategy)))
	}

	return reasons
}
