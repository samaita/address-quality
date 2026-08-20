// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"regexp"
	"strings"

	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

var roadPrefixPattern = regexp.MustCompile(`\b(jl|jalan|gg|gang)\b`)

// detectRoadContextTokens returns the set of normalized tokens that follow a
// road prefix (jl/jalan/gg/gang) in the raw (pre-normalization) text. Such
// tokens are road names, not location evidence, and must not block the city
// priority. e.g. "Jl. Aceh, Bandung" -> {"aceh"}; "Jl. Gatot Subroto No.86" -> {"gatot","subroto"}.
func detectRoadContextTokens(raw string) map[string]bool {
	tokens := make(map[string]bool)
	lower := strings.ToLower(raw)
	re := regexp.MustCompile(`\b(jl\.?|jalan|gg\.?|gang)\s+([a-z][a-z\s]*)`)
	for _, m := range re.FindAllStringSubmatch(lower, -1) {
		if len(m) > 2 && m[2] != "" {
			phrase := strings.TrimSpace(m[2])
			// cut at common road-name terminators: numbers, "no", "rt", "rw", comma
			cut := regexp.MustCompile(`\b(no\.?|rt\.?|rw\.?|\d+|,)\b`)
			if idx := cut.FindStringIndex(phrase); idx != nil {
				phrase = phrase[:idx[0]]
			}
			phrase = strings.TrimSpace(phrase)
			if phrase == "" {
				continue
			}
			for _, w := range strings.Fields(phrase) {
				tokens[normalizer.Normalize(w)] = true
			}
		}
	}
	return tokens
}

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
