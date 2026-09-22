// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"strings"

	"address-quality/internal/model"
)

// Tunable knobs for contextual fuzzy recovery. All thresholds are relative to
// the best-supported candidate, so genuine ambiguity survives; tune them via
// the typo/context benchmark (benchmark_contextual_test.go).
const (
	// fuzzyMatchThreshold is the minimum Damerau-Levenshtein similarity
	// (space-insensitive) between an unexplained input span and a
	// neighborhood location name. 0.8 because the canonical case "cihuar
	// gelis" -> "Cihaur Geulis" needs two edits (transposition + insertion)
	// on short words: 1 - 2/12 = 0.833.
	fuzzyMatchThreshold = 0.8
	// fuzzyMaxSpanTokens caps how many adjacent input tokens one recovered
	// name may span (e.g. "pasir kaliki" spans 2).
	fuzzyMaxSpanTokens = 4
	// fuzzyMinFrontierSupport is the minimum number of used evidence items a
	// candidate needs before it may drive fuzzy retrieval.
	fuzzyMinFrontierSupport = 2
	// fuzzyFrontierCoverageGap is how far below the best candidate coverage a
	// candidate may fall and still join the expansion frontier.
	fuzzyFrontierCoverageGap = 0.05
	// fuzzyMinFrontierCoverage is the absolute coverage floor that guards
	// against fully-garbage contexts. Low on purpose: noisy real addresses
	// legitimately have low coverage with a well-supported candidate (the
	// used-evidence count is the context signal, not the ratio).
	fuzzyMinFrontierCoverage = 0.1
	// fuzzyMinTargetKeyLen skips degenerate match targets ("no", "rt").
	fuzzyMinTargetKeyLen = 4
)

// fuzzyStopwords are structural address words that must never be fuzzy
// recovered or counted against coverage. Road prefixes (jl/jalan/gg/gang) are
// stripped earlier; road-name tokens are excluded via roadTokens.
var fuzzyStopwords = map[string]bool{"no": true, "rt": true, "rw": true, "km": true}

// CandidateCoverage measures how much of the available evidence a candidate
// hierarchy explains. Coverage — not raw score — decides expansion
// eligibility: a candidate that ignores most of the input must not be allowed
// to expand merely because one token matched exactly.
type CandidateCoverage struct {
	Candidate      *model.AdminCandidate
	UsedEvidence   []model.Evidence
	UnusedEvidence []model.Evidence
	Coverage       float64
}

// evaluateCandidateCoverage classifies location-relevant evidence as used or
// unused by the candidate. Road-context tokens, function words and digits are
// never counted against coverage.
func evaluateCandidateCoverage(candidate *model.AdminCandidate, allEvidence []model.Evidence, roadTokens map[string]bool) CandidateCoverage {
	usedValues := make(map[string]bool, len(candidate.Evidence))
	for _, me := range candidate.Evidence {
		usedValues[me.Value] = true
	}

	var cov CandidateCoverage
	cov.Candidate = candidate
	for _, ev := range allEvidence {
		if roadTokens[ev.Value] || fuzzyStopwords[ev.Value] || isDigitOnly(ev.Value) {
			continue
		}
		if usedValues[ev.Value] {
			cov.UsedEvidence = append(cov.UsedEvidence, ev)
		} else {
			cov.UnusedEvidence = append(cov.UnusedEvidence, ev)
		}
	}
	if total := len(cov.UsedEvidence) + len(cov.UnusedEvidence); total > 0 {
		cov.Coverage = float64(len(cov.UsedEvidence)) / float64(total)
	}
	return cov
}

// selectExpansionFrontier picks candidates trusted enough to provide fuzzy
// context: minimum positive evidence support, high coverage, low unused
// evidence — relative to the best candidate so ambiguity survives. Only
// pre-fuzzy evidence feeds this decision (no self-reinforcing expansion).
func selectExpansionFrontier(coverages []CandidateCoverage) []CandidateCoverage {
	best := 0.0
	for i := range coverages {
		if coverages[i].Coverage > best {
			best = coverages[i].Coverage
		}
	}
	if best < fuzzyMinFrontierCoverage {
		return nil
	}
	var frontier []CandidateCoverage
	for i := range coverages {
		if len(coverages[i].UsedEvidence) >= fuzzyMinFrontierSupport &&
			coverages[i].Coverage >= best-fuzzyFrontierCoverageGap {
			frontier = append(frontier, coverages[i])
		}
	}
	return frontier
}

// ensureChildrenIndexesLoaded builds parent->children indexes once from the
// already-loaded byID caches and hierarchy map. In-memory only: no DB round
// trips (spec: hierarchy determines WHERE to search, cheaply).
func (svc *Service) ensureChildrenIndexesLoaded() {
	svc.childrenOnce.Do(func() {
		svc.citiesByProvince = make(map[int64][]*cityEntry)
		svc.districtsByCity = make(map[int64][]*districtEntry)
		svc.subDistrictsByDistrict = make(map[int64][]*subDistrictEntry)
		if svc.hierarchyCache == nil {
			return
		}
		for id, e := range svc.cityByID {
			if pid, ok := svc.hierarchyCache.CityToProvince[id]; ok {
				svc.citiesByProvince[pid] = append(svc.citiesByProvince[pid], e)
			}
		}
		for id, e := range svc.districtByID {
			if cid, ok := svc.hierarchyCache.DistrictToCity[id]; ok {
				svc.districtsByCity[cid] = append(svc.districtsByCity[cid], e)
			}
		}
		for id, e := range svc.subDistrictByID {
			if did, ok := svc.hierarchyCache.SubDistrictToDist[id]; ok {
				svc.subDistrictsByDistrict[did] = append(svc.subDistrictsByDistrict[did], e)
			}
		}
	})
}

// candidateNeighborhood returns the candidate's ancestor chain plus children
// one level below its deepest resolved node, as deduplicated entities.
func (svc *Service) candidateNeighborhood(c *model.AdminCandidate) []model.Entity {
	if svc.hierarchyCache == nil {
		return nil
	}
	svc.ensureChildrenIndexesLoaded()

	seen := make(map[int64]bool)
	out := make([]model.Entity, 0, 8)
	add := func(id int64, name, level, postal string) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, model.Entity{ID: id, Name: name, Level: level, PostalCode: postal})
	}

	loc := c.Location
	if loc.Province != nil {
		add(loc.Province.ID, loc.Province.Name, "PROVINCE", "")
	}
	if loc.City != nil {
		add(loc.City.ID, loc.City.Name, "CITY", loc.City.PostalCode)
	}
	if loc.District != nil {
		add(loc.District.ID, loc.District.Name, "DISTRICT", "")
	}
	if loc.SubDistrict != nil {
		add(loc.SubDistrict.ID, loc.SubDistrict.Name, "SUBDISTRICT", loc.SubDistrict.PostalCode)
	}

	// children for unresolved lower levels, one level below the deepest node
	switch {
	case loc.SubDistrict != nil: // leaf
	case loc.District != nil:
		for _, e := range svc.subDistrictsByDistrict[loc.District.ID] {
			add(e.ID, e.Name, "SUBDISTRICT", e.PostalCode)
		}
	case loc.City != nil:
		for _, e := range svc.districtsByCity[loc.City.ID] {
			add(e.ID, e.Name, "DISTRICT", "")
		}
	case loc.Province != nil:
		for _, e := range svc.citiesByProvince[loc.Province.ID] {
			add(e.ID, e.Name, "CITY", e.PostalCode)
		}
	}
	return out
}

// postalEntityCompatible reports whether a postal-resolved subdistrict fits
// the candidate's resolved hierarchy: every level the candidate already
// asserts must agree with the entity's ancestry chain.
func (svc *Service) postalEntityCompatible(e model.Entity, c *model.AdminCandidate) bool {
	if svc.hierarchyCache == nil || e.Level != "SUBDISTRICT" {
		return false
	}
	distID, ok := svc.hierarchyCache.SubDistrictToDist[e.ID]
	if !ok {
		return false
	}
	loc := c.Location
	if loc.District != nil && loc.District.ID != distID {
		return false
	}
	cityID, hasCity := svc.hierarchyCache.DistrictToCity[distID]
	if loc.City != nil && (!hasCity || cityID != loc.City.ID) {
		return false
	}
	if loc.Province != nil {
		if !hasCity {
			return false
		}
		provID, ok := svc.hierarchyCache.CityToProvince[cityID]
		if !ok || provID != loc.Province.ID {
			return false
		}
	}
	return true
}

// buildMatchTargets indexes neighborhood names for fuzzy comparison: full
// names AND their individual words, lowercased with spaces stripped, so both
// "pasir kaliki"↔"Pasirkaliki" and "bndung"↔"Bandung" (word of "Kota
// Bandung") are reachable.
func buildMatchTargets(entities []model.Entity) map[string][]model.Entity {
	targets := make(map[string][]model.Entity)
	add := func(key string, e model.Entity) {
		if len([]rune(key)) < fuzzyMinTargetKeyLen {
			return
		}
		targets[key] = append(targets[key], e)
	}
	for _, e := range entities {
		norm := strings.ToLower(e.Name)
		add(compactSpaces(norm), e)
		for _, w := range strings.Fields(norm) {
			add(compactSpaces(w), e)
		}
	}
	return targets
}

// unexplainedRuns returns runs of consecutive word indices that are not
// explained by the frontier candidates and are not structural noise.
func unexplainedRuns(words []string, explained, roadTokens map[string]bool) [][]int {
	var runs [][]int
	var cur []int
	for i, w := range words {
		if explained[w] || roadTokens[w] || fuzzyStopwords[w] || isDigitOnly(w) {
			if len(cur) > 0 {
				runs = append(runs, cur)
				cur = nil
			}
			continue
		}
		cur = append(cur, i)
	}
	if len(cur) > 0 {
		runs = append(runs, cur)
	}
	return runs
}

// RecoverContextualEvidence is the second-pass evidence recovery: use the
// exact-resolution candidates as context, compare their neighborhood
// vocabulary against unexplained spans of the original normalized input, and
// append recovered evidence with pre-resolved entities. One pass only.
//
// Returns the recovered evidence, span-level corrections (typo/normalization
// use the same metadata mechanism), and the set of input token values the
// recovery consumed (so typo tokens no longer count as unused evidence).
func (svc *Service) RecoverContextualEvidence(candidates []model.AdminCandidate, resolved []model.ResolvedEvidence, normalized string, roadTokens map[string]bool) (recovered []model.ResolvedEvidence, corrections map[string]string, explained map[string]bool) {
	coverages := make([]CandidateCoverage, len(candidates))
	for i := range candidates {
		coverages[i] = evaluateCandidateCoverage(&candidates[i], resolvedEvidenceToEvidence(resolved), roadTokens)
	}
	frontier := selectExpansionFrontier(coverages)
	if len(frontier) == 0 {
		return nil, nil, nil
	}

	// batch: one deduplicated neighborhood across all frontier candidates.
	// Postal evidence never enters discovery (least reliable), but its
	// resolved entities are cheap in-memory vocabulary for recovery when
	// they fit a frontier candidate's hierarchy (e.g. postal 40391 ->
	// kelurahan Lembang when the city Kab. Bandung Barat is the context).
	postalEntities := make([]model.Entity, 0, 4)
	for _, re := range resolved {
		if re.Evidence.Type != model.EvidencePostalCode {
			continue
		}
		for _, c := range re.Candidates {
			postalEntities = append(postalEntities, c)
		}
	}

	seenEntity := make(map[int64]bool)
	var entities []model.Entity
	for _, fc := range frontier {
		for _, e := range svc.candidateNeighborhood(fc.Candidate) {
			if !seenEntity[e.ID] {
				seenEntity[e.ID] = true
				entities = append(entities, e)
			}
		}
		for _, e := range postalEntities {
			if !seenEntity[e.ID] && svc.postalEntityCompatible(e, fc.Candidate) {
				seenEntity[e.ID] = true
				entities = append(entities, e)
			}
		}
	}
	targets := buildMatchTargets(entities)

	explainedByFrontier := make(map[string]bool)
	for _, fc := range frontier {
		for _, ev := range fc.UsedEvidence {
			explainedByFrontier[ev.Value] = true
		}
	}

	knownEntity := make(map[int64]bool)
	for _, re := range resolved {
		if re.Evidence.Type == model.EvidencePostalCode {
			continue // postal entities never enter candidate discovery
		}
		for _, c := range re.Candidates {
			knownEntity[c.ID] = true
		}
	}

	words := strings.Fields(normalized)
	fuzzyExplained := make(map[string]bool)
	for _, run := range unexplainedRuns(words, explainedByFrontier, roadTokens) {
		i := 0
		for i < len(run) {
			matched := false
			maxLen := min(fuzzyMaxSpanTokens, len(run)-i)
			for l := maxLen; l >= 1; l-- {
				span := words[run[i] : run[i]+l] // run is contiguous word indices
				spanText := strings.Join(span, " ")
				ents, _ := bestSpanMatch(compactSpaces(strings.ToLower(spanText)), targets)
				if ents == nil {
					continue
				}
				matched = true
				var fresh []model.Entity
				for _, e := range ents {
					if !knownEntity[e.ID] {
						knownEntity[e.ID] = true
						fresh = append(fresh, e)
					}
				}
				if len(fresh) > 0 {
					recovered = append(recovered, model.ResolvedEvidence{
						Evidence:   model.Evidence{Type: model.EvidencePlaceName, Value: spanText},
						Candidates: fresh,
					})
					if corrections == nil {
						corrections = make(map[string]string)
					}
					corrections[spanText] = fresh[0].Name
				}
				for _, w := range span {
					fuzzyExplained[w] = true
				}
				i += l
				break
			}
			if !matched {
				i++
			}
		}
	}
	if len(recovered) == 0 {
		return nil, nil, nil
	}
	return recovered, corrections, fuzzyExplained
}

// bestSpanMatch returns all entities whose key similarity to the span equals
// the best score, provided that score clears the threshold (ties kept so
// genuine ambiguity is preserved).
func bestSpanMatch(spanKey string, targets map[string][]model.Entity) ([]model.Entity, float64) {
	if len(spanKey) < fuzzyMinTargetKeyLen {
		return nil, 0
	}
	best := 0.0
	for k := range targets {
		if s := damerauSimilarity(spanKey, k); s > best {
			best = s
		}
	}
	if best < fuzzyMatchThreshold {
		return nil, 0
	}
	seen := make(map[int64]bool)
	var out []model.Entity
	for k, ents := range targets {
		if damerauSimilarity(spanKey, k) < best {
			continue
		}
		for _, e := range ents {
			if !seen[e.ID] {
				seen[e.ID] = true
				out = append(out, e)
			}
		}
	}
	return out, best
}

func resolvedEvidenceToEvidence(resolved []model.ResolvedEvidence) []model.Evidence {
	out := make([]model.Evidence, len(resolved))
	for i, re := range resolved {
		out[i] = re.Evidence
	}
	return out
}

func compactSpaces(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "")
}

func isDigitOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// damerauLevenshtein computes the restricted Damerau-Levenshtein (optimal
// string alignment) distance: insertion, deletion, substitution and adjacent
// transposition each cost 1.
func damerauLevenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	la, lb := len(ar), len(br)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev2 := make([]int, lb+1)
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			m := min(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && ar[i-1] == br[j-2] && ar[i-2] == br[j-1] {
				if t := prev2[j-2] + 1; t < m {
					m = t
				}
			}
			cur[j] = m
		}
		prev2, prev, cur = prev, cur, prev2
	}
	return prev[lb]
}

func damerauSimilarity(a, b string) float64 {
	longest := max(len([]rune(a)), len([]rune(b)))
	if longest == 0 {
		return 1
	}
	return 1 - float64(damerauLevenshtein(a, b))/float64(longest)
}
