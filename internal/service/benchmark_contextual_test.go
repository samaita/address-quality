// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/sanitizer"
)

// TestBenchmarkContextualRecovery is the dedicated typo/context benchmark:
// it corrupts real location names (deletion, insertion, substitution,
// transposition, missing space, extra space) at every admin level, runs the
// full V1 pipeline, and reports:
//
//   - recovery recall:      typo cases where contextual recovery fired
//   - false recovery rate:  clean cases where recovery fired anyway
//   - hierarchy accuracy:   all-four-levels correct rate, clean vs typo, per level
//
// The goal is not "did fuzzy find something" but "did contextual recovery
// improve the final hierarchy without introducing false evidence". Thresholds
// in fuzzy_context.go are tuned against these numbers. Skips without the
// seeded SQLite location DB.
func TestBenchmarkContextualRecovery(t *testing.T) {
	if _, err := os.Stat("../../db/location.db"); err != nil {
		t.Skip("db/location.db not found")
	}
	repo, err := database.NewLocationDB("../../db/location.db", 2)
	if err != nil {
		t.Skipf("cannot open location db: %v", err)
	}
	svc := New(nil, repo, nil, sanitizer.New(sanitizer.DefaultPolicy()), 1000, "kemendagri", false, true, "", "")

	ctx := context.Background()
	sourceID, _, err := repo.FindSourceByCode(ctx, "kemendagri")
	if err != nil {
		t.Skipf("kemendagri source not in db: %v", err)
	}
	if err := svc.ensureEntitiesCachesLoaded(ctx, sourceID); err != nil {
		t.Fatalf("load caches: %v", err)
	}

	// sample subdistricts with a complete hierarchy chain
	rnd := rand.New(rand.NewSource(42)) // deterministic
	type place struct {
		sub, dist, city, prov, postal string
	}
	var places []place
	for _, subs := range svc.subDistrictCache {
		for _, s := range subs {
			distID, ok1 := svc.hierarchyCache.SubDistrictToDist[s.ID]
			dist := svc.districtByID[distID]
			if !ok1 || dist == nil {
				continue
			}
			cityID, ok2 := svc.hierarchyCache.DistrictToCity[dist.ID]
			city := svc.cityByID[cityID]
			if !ok2 || city == nil {
				continue
			}
			provID, ok3 := svc.hierarchyCache.CityToProvince[city.ID]
			prov := svc.provinceByID[provID]
			if !ok3 || prov == nil {
				continue
			}
			places = append(places, place{s.Name, dist.Name, city.Name, prov.Name, s.PostalCode})
		}
	}
	if len(places) < 20 {
		t.Fatalf("not enough hierarchy-complete places: %d", len(places))
	}
	rnd.Shuffle(len(places), func(i, j int) { places[i], places[j] = places[j], places[i] })
	const sampleSize = 20
	places = places[:sampleSize]

	// typo operations over a name's words
	type typoOp struct {
		name string
		fn   func(words []string, r *rand.Rand) string
	}
	pickWord := func(words []string, r *rand.Rand) int { return r.Intn(len(words)) }
	ops := []typoOp{
		{"deletion", func(words []string, r *rand.Rand) string {
			i := pickWord(words, r)
			w := []rune(words[i])
			if len(w) < 4 {
				return strings.Join(words, " ")
			}
			p := r.Intn(len(w))
			words[i] = string(append(w[:p:p], w[p+1:]...))
			return strings.Join(words, " ")
		}},
		{"insertion", func(words []string, r *rand.Rand) string {
			i := pickWord(words, r)
			w := []rune(words[i])
			p := r.Intn(len(w) + 1)
			c := rune('a' + r.Intn(26))
			words[i] = string(append(append(w[:p:p], c), w[p:]...))
			return strings.Join(words, " ")
		}},
		{"substitution", func(words []string, r *rand.Rand) string {
			i := pickWord(words, r)
			w := []rune(words[i])
			p := r.Intn(len(w))
			c := rune('a' + r.Intn(26))
			if w[p] == c {
				c = 'a' + (c-'a'+1)%26
			}
			w[p] = c
			words[i] = string(w)
			return strings.Join(words, " ")
		}},
		{"transposition", func(words []string, r *rand.Rand) string {
			i := pickWord(words, r)
			w := []rune(words[i])
			if len(w) < 2 {
				return strings.Join(words, " ")
			}
			p := r.Intn(len(w) - 1)
			if w[p] == w[p+1] {
				return strings.Join(words, " ")
			}
			w[p], w[p+1] = w[p+1], w[p]
			words[i] = string(w)
			return strings.Join(words, " ")
		}},
		{"missing_space", func(words []string, r *rand.Rand) string {
			if len(words) < 2 {
				return strings.Join(words, " ")
			}
			w := rnd.Intn(len(words) - 1)
			out := append([]string{}, words...)
			out[w] = out[w] + out[w+1]
			return strings.Join(append(out[:w+1:w+1], out[w+2:]...), " ")
		}},
		{"extra_space", func(words []string, r *rand.Rand) string {
			i := pickWord(words, r)
			w := []rune(words[i])
			if len(w) < 6 {
				return strings.Join(words, " ")
			}
			p := 1 + r.Intn(len(w)-1)
			out := append([]string{}, words...)
			out[i] = string(w[:p]) + " " + string(w[p:])
			return strings.Join(out, " ")
		}},
	}

	address := func(p place, level int, nameOverride string) string {
		sub, dist, city, prov := p.sub, p.dist, p.city, p.prov
		switch level {
		case 2:
			prov = nameOverride
		case 3:
			city = nameOverride
		case 4:
			dist = nameOverride
		case 5:
			sub = nameOverride
		}
		return fmt.Sprintf("Jl. Merdeka No.1, %s, %s, %s, %s", sub, dist, city, prov)
	}

	type result struct {
		corrections bool
		correct     bool
	}
	levels := map[int]string{2: "province", 3: "city", 4: "district", 5: "subdistrict"}
	cleanResults := make([]result, 0, len(places))
	typoResults := make(map[int][]result) // level -> results

	run := func(addr string, p place) result {
		resp, err := svc.ValidateAddressV1(ctx, &model.AddressRequest{Address: addr, SourceCode: "kemendagri"}, "benchmark")
		if err != nil {
			t.Fatalf("validate %q: %v", addr, err)
		}
		loc := resp.Data.Location
		return result{
			corrections: len(resp.Data.Metadata.FuzzyCorrections) > 0,
			correct: strings.EqualFold(loc.Province, p.prov) &&
				strings.EqualFold(loc.City, p.city) &&
				strings.EqualFold(loc.District, p.dist) &&
				strings.EqualFold(loc.SubDistrict, p.sub),
		}
	}

	for _, p := range places {
		// clean control: expected = its own chain
		resp, err := svc.ValidateAddressV1(ctx, &model.AddressRequest{
			Address:    fmt.Sprintf("Jl. Merdeka No.1, %s, %s, %s, %s", p.sub, p.dist, p.city, p.prov),
			SourceCode: "kemendagri",
		}, "benchmark")
		if err != nil {
			t.Fatalf("validate clean %q: %v", p.sub, err)
		}
		loc := resp.Data.Location
		cleanResults = append(cleanResults, result{
			corrections: len(resp.Data.Metadata.FuzzyCorrections) > 0,
			correct: strings.EqualFold(loc.Province, p.prov) &&
				strings.EqualFold(loc.City, p.city) &&
				strings.EqualFold(loc.District, p.dist) &&
				strings.EqualFold(loc.SubDistrict, p.sub),
		})

		// typo variants: corruption at each level, expected = its own chain
		for level := range levels {
			name := map[int]string{2: p.prov, 3: p.city, 4: p.dist, 5: p.sub}[level]
			words := strings.Fields(name)
			corrupted := ops[rnd.Intn(len(ops))].fn(append([]string{}, words...), rnd)
			if corrupted == name {
				continue // op was a no-op (e.g. single-word name)
			}
			typoResults[level] = append(typoResults[level], run(address(p, level, corrupted), p))
		}
	}

	pct := func(n, total int) float64 {
		if total == 0 {
			return 0
		}
		return float64(n) / float64(total) * 100
	}

	var cleanOK, cleanRecovered int
	for _, r := range cleanResults {
		if r.correct {
			cleanOK++
		}
		if r.corrections {
			cleanRecovered++
		}
	}
	t.Logf("CLEAN  : accuracy=%.1f%% false_recovery=%.1f%% (n=%d)",
		pct(cleanOK, len(cleanResults)), pct(cleanRecovered, len(cleanResults)), len(cleanResults))

	totalTypo, totalTypoOK, totalRecovered := 0, 0, 0
	for _, level := range []int{2, 3, 4, 5} {
		rs := typoResults[level]
		var ok, recovered int
		for _, r := range rs {
			if r.correct {
				ok++
			}
			if r.corrections {
				recovered++
			}
		}
		t.Logf("TYPO %-12s: accuracy=%.1f%% recovery_recall=%.1f%% (n=%d)",
			levels[level], pct(ok, len(rs)), pct(recovered, len(rs)), len(rs))
		totalTypo += len(rs)
		totalTypoOK += ok
		totalRecovered += recovered
	}
	t.Logf("TYPO  : accuracy=%.1f%% recovery_recall=%.1f%% (n=%d)",
		pct(totalTypoOK, totalTypo), pct(totalRecovered, totalTypo), totalTypo)

	// sanity gate: the pipeline must resolve clean addresses at all
	if cleanOK == 0 {
		t.Fatalf("clean accuracy is 0%% — pipeline broken")
	}
}
