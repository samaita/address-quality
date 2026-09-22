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
var roadContextPattern = regexp.MustCompile(`(?i)\b(jl\.?|jalan|gg\.?|gang)\s+([a-z][a-z\s]*)`)
var roadContextCutPattern = regexp.MustCompile(`(?i)\b(no\.?|rt\.?|rw\.?|\d+|,)\b`)

// detectRoadContextTokens returns the set of normalized tokens that follow a
// road prefix (jl/jalan/gg/gang) in the raw (pre-normalization) text. Such
// tokens are road names, not location evidence, and must not block the city
// priority. e.g. "Jl. Aceh, Bandung" -> {"aceh"}; "Jl. Gatot Subroto No.86" -> {"gatot","subroto"}.
func detectRoadContextTokens(raw string) map[string]bool {
	tokens := make(map[string]bool)
	lower := strings.ToLower(raw)
	for _, m := range roadContextPattern.FindAllStringSubmatch(lower, -1) {
		if len(m) > 2 && m[2] != "" {
			phrase := strings.TrimSpace(m[2])
			// cut at common road-name terminators: numbers, "no", "rt", "rw", comma
			if idx := roadContextCutPattern.FindStringIndex(phrase); idx != nil {
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

// stripRoadContext removes road prefixes and their following road-name span
// while retaining the rest of the address. It is used only by compact exact
// matching, so a road name cannot become a whitespace-normalized location
// alias while a later administrative occurrence remains available.
func stripRoadContext(raw string) string {
	return roadContextPattern.ReplaceAllStringFunc(raw, func(match string) string {
		parts := roadContextPattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		phrase := parts[2]
		if idx := roadContextCutPattern.FindStringIndex(phrase); idx != nil {
			return phrase[idx[0]:]
		}
		return ""
	})
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
		if postalCodePattern.MatchString(w) || roadPrefixPattern.MatchString(w) {
			continue
		}
		lower := strings.ToLower(w)
		if seen[lower] {
			continue
		}
		seen[lower] = true
		evidence = append(evidence, model.Evidence{Type: model.EvidencePlaceName, Value: w})
	}

	return evidence
}
