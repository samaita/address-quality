// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"math"

	"address-quality/internal/database"
	"address-quality/internal/model"
)

func EvaluateCandidate(
	provinceCands, cityCands, districtCands, subDistrictCands []model.Candidate,
	winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID int64,
	postalCodeMatched bool,
	inputPostalCode string,
	hierarchy *database.HierarchyMap,
) model.CandidateEvaluation {
	hierarchyValid := validateHierarchy(winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID, hierarchy)
	matched, missing := evaluateCompleteness(winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID, postalCodeMatched, inputPostalCode)
	conflicts := detectConflicts(hierarchyValid, winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID, postalCodeMatched, inputPostalCode)
	confidence := scoreConfidence(provinceCands, cityCands, districtCands, subDistrictCands, winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID, hierarchyValid, postalCodeMatched)
	status := assessQuality(conflicts, missing, winnerProvinceID)

	reasons := buildReasons(hierarchyValid, provinceCands, cityCands, districtCands, subDistrictCands)

	return model.CandidateEvaluation{
		Confidence: confidence,
		Status:     status,
		Matched:    matched,
		Missing:    missing,
		Conflicts:  conflicts,
		Reasons:    reasons,
	}
}

func validateHierarchy(provinceID, cityID, districtID, subDistrictID int64, hierarchy *database.HierarchyMap) bool {
	if hierarchy == nil {
		return cityID == 0 && districtID == 0 && subDistrictID == 0
	}
	if cityID > 0 {
		if hierarchy.CityToProvince[cityID] != provinceID {
			return false
		}
	}
	if districtID > 0 {
		if hierarchy.DistrictToCity[districtID] != cityID {
			return false
		}
	}
	if subDistrictID > 0 {
		if hierarchy.SubDistrictToDist[subDistrictID] != districtID {
			return false
		}
	}
	return true
}

func evaluateCompleteness(provinceID, cityID, districtID, subDistrictID int64, postalCodeMatched bool, inputPostalCode string) ([]model.Component, []model.Component) {
	var matched, missing []model.Component

	check := func(id int64, comp model.Component) {
		if id > 0 {
			matched = append(matched, comp)
		} else {
			missing = append(missing, comp)
		}
	}

	check(provinceID, model.ComponentProvince)
	check(cityID, model.ComponentCity)
	check(districtID, model.ComponentDistrict)
	check(subDistrictID, model.ComponentSubDistrict)

	if inputPostalCode != "" {
		if postalCodeMatched {
			matched = append(matched, model.ComponentPostalCode)
		} else {
			missing = append(missing, model.ComponentPostalCode)
		}
	}

	return matched, missing
}

func detectConflicts(hierarchyValid bool, provinceID, cityID, districtID, subDistrictID int64, postalCodeMatched bool, inputPostalCode string) []model.Conflict {
	var conflicts []model.Conflict

	if cityID > 0 && !hierarchyValid {
		conflicts = append(conflicts, model.Conflict{
			Type:    "hierarchy_conflict",
			Message: "administrative hierarchy is invalid",
		})
	}

	if inputPostalCode != "" && subDistrictID > 0 && !postalCodeMatched {
		conflicts = append(conflicts, model.Conflict{
			Type:    "postal_code_mismatch",
			Message: "postal code does not match the resolved sub-district",
		})
	}

	return conflicts
}

func scoreConfidence(
	provinceCands, cityCands, districtCands, subDistrictCands []model.Candidate,
	winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID int64,
	hierarchyValid bool,
	postalCodeMatched bool,
) float64 {
	var score float64

	hasExact := false
	for _, c := range provinceCands {
		if c.MatchType == "EXACT" {
			hasExact = true
			break
		}
	}
	if !hasExact {
		for _, c := range cityCands {
			if c.MatchType == "EXACT" {
				hasExact = true
				break
			}
		}
	}
	if !hasExact {
		for _, c := range districtCands {
			if c.MatchType == "EXACT" {
				hasExact = true
				break
			}
		}
	}
	if !hasExact {
		for _, c := range subDistrictCands {
			if c.MatchType == "EXACT" {
				hasExact = true
				break
			}
		}
	}
	if hasExact {
		score += 0.40
	}

	if hierarchyValid && winnerCityID > 0 {
		score += 0.30
	}

	if postalCodeMatched {
		score += 0.20
	}

	if winnerProvinceID > 0 {
		score += 0.10
	}

	if score > 1.0 {
		score = 1.0
	}
	return math.Round(score*10000) / 10000
}

func assessQuality(conflicts []model.Conflict, missing []model.Component, winnerProvinceID int64) model.QualityStatus {
	if len(conflicts) > 0 {
		return model.StatusConflict
	}

	if winnerProvinceID == 0 {
		return model.StatusUnknown
	}

	for _, m := range missing {
		if m != model.ComponentPostalCode {
			return model.StatusIncomplete
		}
	}

	return model.StatusValid
}

func buildReasons(hierarchyValid bool, provinceCands, cityCands, districtCands, subDistrictCands []model.Candidate) []string {
	var reasons []string

	hasExact := false
	for _, list := range [][]model.Candidate{provinceCands, cityCands, districtCands, subDistrictCands} {
		for _, c := range list {
			if c.MatchType == "EXACT" {
				hasExact = true
				break
			}
		}
		if hasExact {
			break
		}
	}
	if hasExact {
		reasons = append(reasons, "exact_match")
	}

	if hierarchyValid {
		reasons = append(reasons, "hierarchy_validation")
	}

	return reasons
}
