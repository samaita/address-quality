// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"regexp"
	"strings"

	"address-quality/internal/model"
)

var roadPrefixPattern = regexp.MustCompile(`\b(jl|jalan|gg|gang)\b`)

func ExtractEvidence(normalized string) []model.Evidence {
	var evidence []model.Evidence

	if m := postalCodePattern.FindStringSubmatch(normalized); len(m) > 1 {
		evidence = append(evidence, model.Evidence{
			Type:  model.EvidencePostalCode,
			Value: m[1],
		})
	}

	rest := roadPrefixPattern.ReplaceAllString(normalized, "")
	rest = strings.TrimSpace(rest)

	words := strings.Fields(rest)
	seen := make(map[string]bool)
	for _, w := range words {
		if postalCodePattern.MatchString(w) {
			continue
		}
		if roadPrefixPattern.MatchString(w) {
			continue
		}
		lower := strings.ToLower(w)
		if seen[lower] {
			continue
		}
		seen[lower] = true
		if isRoadNameWord(w) {
			evidence = append(evidence, model.Evidence{
				Type:  model.EvidenceRoadName,
				Value: w,
			})
		} else {
			evidence = append(evidence, model.Evidence{
				Type:  model.EvidencePlaceName,
				Value: w,
			})
		}
	}

	return evidence
}

func isRoadNameWord(w string) bool {
	roadIndicators := []string{"jl", "jalan", "gg", "gang"}
	for _, ri := range roadIndicators {
		if strings.EqualFold(w, ri) {
			return false
		}
	}
	return false
}
