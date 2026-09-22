package service

import (
	"context"
	"testing"

	"address-quality/internal/database"
	"address-quality/internal/model"
)

// newFixtureService builds an in-memory Service with a small Indonesia-like
// hierarchy:
//
//	Jawa Barat (1)
//	├── Kota Bandung (10)
//	│   ├── Cicendo (100): Pasirkaliki (1000), Arjuna (1001)
//	│   └── Cibeunying Kaler (102): Cihaur Geulis (1003), Cihapit (1004)
//	└── Kabupaten Majalengka (11)
//	    └── Palasah (101): Palasah (1002)
func newFixtureService() *Service {
	svc := &Service{
		hierarchyCache: &database.HierarchyMap{
			CityToProvince:    map[int64]int64{10: 1, 11: 1},
			DistrictToCity:    map[int64]int64{100: 10, 101: 11, 102: 10},
			SubDistrictToDist: map[int64]int64{1000: 100, 1001: 100, 1002: 101, 1003: 102, 1004: 102},
		},
		cityByID: map[int64]*cityEntry{
			10: {ID: 10, Name: "Kota Bandung", PostalCode: "40171"},
			11: {ID: 11, Name: "Kabupaten Majalengka"},
		},
		provinceByID: map[int64]*provinceEntry{
			1: {ID: 1, Name: "Jawa Barat"},
		},
		districtByID: map[int64]*districtEntry{
			100: {ID: 100, Name: "Cicendo"},
			101: {ID: 101, Name: "Palasah"},
			102: {ID: 102, Name: "Cibeunying Kaler"},
		},
		subDistrictByID: map[int64]*subDistrictEntry{
			1000: {ID: 1000, Name: "Pasirkaliki", PostalCode: "40171"},
			1001: {ID: 1001, Name: "Arjuna", PostalCode: "40172"},
			1002: {ID: 1002, Name: "Palasah", PostalCode: "45465"},
			1003: {ID: 1003, Name: "Cihaur Geulis", PostalCode: "40122"},
			1004: {ID: 1004, Name: "Cihapit", PostalCode: "40114"},
		},
	}
	return svc
}

func ev(t model.EvidenceType, v string, candidates ...model.Entity) model.ResolvedEvidence {
	return model.ResolvedEvidence{Evidence: model.Evidence{Type: t, Value: v}, Candidates: candidates}
}

func TestDamerauLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"cihuar", "cihaur", 1}, // adjacent transposition
		{"bndung", "bandung", 1},
		{"cihaurgelis", "cihaurgeulis", 1},
		{"aceh", "aceh", 0},
		{"abc", "", 3},
	}
	for _, c := range cases {
		if got := damerauLevenshtein(c.a, c.b); got != c.want {
			t.Errorf("damerauLevenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
	// space-insensitivity happens in the caller: "pasir kaliki" -> "pasirkaliki"
	if s := damerauSimilarity(compactSpaces("pasir kaliki"), "pasirkaliki"); s != 1 {
		t.Errorf("similarity(compactSpaces(pasir kaliki), pasirkaliki) = %f, want 1", s)
	}
	if s := damerauSimilarity("bndung", "bandung"); s < fuzzyMatchThreshold {
		t.Errorf("similarity(bndung, bandung) = %f, want >= %f", s, fuzzyMatchThreshold)
	}
	if s := damerauSimilarity(compactSpaces("cihuar gelis"), "cihaurgeulis"); s < fuzzyMatchThreshold {
		t.Errorf("similarity(cihuargelis, cihaurgeulis) = %f, want >= %f", s, fuzzyMatchThreshold)
	}
}

// Spec test: multi-word location with misleading exact matches. "Pasir" and
// "Kaliki" resolve globally (Majalengka / Merauke) but are contextually wrong;
// recovery must still find Pasirkaliki via the Cicendo neighborhood.
func TestRecoverMultiWordWithMisleadingExactMatches(t *testing.T) {
	svc := newFixtureService()
	normalized := "jl cihampelas no pasir kaliki cicendo bandung jawa barat 40171"
	roadTokens := map[string]bool{"cihampelas": true}
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "pasir", model.Entity{ID: 1002, Name: "Palasah", Level: "SUBDISTRICT"}),      // Majalengka stand-in
		ev(model.EvidencePlaceName, "kaliki", model.Entity{ID: 900, Name: "Kaliki", Level: "SUBDISTRICT"}),       // Merauke stand-in
		ev(model.EvidencePlaceName, "cicendo", model.Entity{ID: 100, Name: "Cicendo", Level: "DISTRICT"}),        //
		ev(model.EvidencePlaceName, "bandung", model.Entity{ID: 10, Name: "Kota Bandung", Level: "CITY"}),        //
		ev(model.EvidencePlaceName, "jawa", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),          //
		ev(model.EvidencePlaceName, "barat", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),         //
		ev(model.EvidencePostalCode, "40171", model.Entity{ID: 1000, Name: "Pasirkaliki", Level: "SUBDISTRICT"}), //
	}

	candidates := buildFixtureCandidates(svc, resolved)
	recovered, corrections, explained := svc.RecoverContextualEvidence(candidates, resolved, normalized, roadTokens)

	if len(recovered) != 1 {
		t.Fatalf("expected 1 recovered evidence, got %+v", recovered)
	}
	if recovered[0].Value != "pasir kaliki" {
		t.Fatalf("expected span \"pasir kaliki\", got %q", recovered[0].Value)
	}
	if len(recovered[0].Candidates) != 1 || recovered[0].Candidates[0].ID != 1000 || recovered[0].Candidates[0].Level != "SUBDISTRICT" {
		t.Fatalf("expected Pasirkaliki subdistrict entity, got %+v", recovered[0].Candidates)
	}
	if corrections["pasir kaliki"] != "Pasirkaliki" {
		t.Fatalf("unexpected corrections: %v", corrections)
	}
	if !explained["pasir"] || !explained["kaliki"] {
		t.Fatalf("expected consumed tokens marked explained, got %v", explained)
	}
}

// Spec test: higher-level typo recovered through hierarchy context.
func TestRecoverCityTypo(t *testing.T) {
	svc := newFixtureService()
	normalized := "cicendo bndung jawa barat"
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "cicendo", model.Entity{ID: 100, Name: "Cicendo", Level: "DISTRICT"}),
		ev(model.EvidencePlaceName, "jawa", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePlaceName, "barat", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePlaceName, "bndung"),
	}
	candidates := buildFixtureCandidates(svc, resolved)
	recovered, corrections, _ := svc.RecoverContextualEvidence(candidates, resolved, normalized, nil)

	if len(recovered) != 1 || recovered[0].Value != "bndung" {
		t.Fatalf("expected recovered \"bndung\", got %+v", recovered)
	}
	if len(recovered[0].Candidates) != 1 || recovered[0].Candidates[0].ID != 10 {
		t.Fatalf("expected Kota Bandung entity, got %+v", recovered[0].Candidates)
	}
	if corrections["bndung"] != "Kota Bandung" {
		t.Fatalf("unexpected corrections: %v", corrections)
	}
}

// Spec test: lower-level multi-word typo recovered when context exists.
func TestRecoverSubdistrictTypo(t *testing.T) {
	svc := newFixtureService()
	normalized := "cihuar gelis cibeunying kaler bandung jawa barat"
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "cibeunying", model.Entity{ID: 102, Name: "Cibeunying Kaler", Level: "DISTRICT"}),
		ev(model.EvidencePlaceName, "kaler", model.Entity{ID: 102, Name: "Cibeunying Kaler", Level: "DISTRICT"}),
		ev(model.EvidencePlaceName, "bandung", model.Entity{ID: 10, Name: "Kota Bandung", Level: "CITY"}),
		ev(model.EvidencePlaceName, "jawa", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePlaceName, "barat", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePlaceName, "cihuar"),
		ev(model.EvidencePlaceName, "gelis"),
	}
	candidates := buildFixtureCandidates(svc, resolved)
	recovered, corrections, _ := svc.RecoverContextualEvidence(candidates, resolved, normalized, nil)

	if len(recovered) != 1 || recovered[0].Value != "cihuar gelis" {
		t.Fatalf("expected recovered \"cihuar gelis\", got %+v", recovered)
	}
	if len(recovered[0].Candidates) != 1 || recovered[0].Candidates[0].ID != 1003 {
		t.Fatalf("expected Cihaur Geulis entity, got %+v", recovered[0].Candidates)
	}
	if corrections["cihuar gelis"] != "Cihaur Geulis" {
		t.Fatalf("unexpected corrections: %v", corrections)
	}
}

// Spec test: weak candidate (single generic evidence) must not expand.
func TestWeakCandidateDoesNotExpand(t *testing.T) {
	svc := newFixtureService()
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "bandung", model.Entity{ID: 10, Name: "Kota Bandung", Level: "CITY"}),
	}
	candidates := buildFixtureCandidates(svc, resolved)
	recovered, corrections, explained := svc.RecoverContextualEvidence(candidates, resolved, "bandung", nil)
	if recovered != nil || corrections != nil || explained != nil {
		t.Fatalf("expected no recovery for weak candidate, got %+v", recovered)
	}
}

// Spec test: ambiguous strong candidates both join the frontier.
func TestAmbiguousStrongCandidatesBothExpand(t *testing.T) {
	coverages := []CandidateCoverage{
		{Coverage: 0.80, UsedEvidence: make([]model.Evidence, 3)},
		{Coverage: 0.78, UsedEvidence: make([]model.Evidence, 3)},
		{Coverage: 0.35, UsedEvidence: make([]model.Evidence, 3)},
		{Coverage: 0.20, UsedEvidence: make([]model.Evidence, 1)},
	}
	frontier := selectExpansionFrontier(coverages)
	if len(frontier) != 2 {
		t.Fatalf("expected 2 frontier candidates (0.80, 0.78), got %d", len(frontier))
	}
}

// Spec test: no useful initial context -> recovery skipped entirely.
func TestNoUsefulContextSkipsRecovery(t *testing.T) {
	svc := newFixtureService()
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "zzzz"),
		ev(model.EvidencePlaceName, "qqqq"),
	}
	candidates := buildFixtureCandidates(svc, resolved)
	recovered, corrections, explained := svc.RecoverContextualEvidence(candidates, resolved, "zzzz qqqq", nil)
	if recovered != nil || corrections != nil || explained != nil {
		t.Fatalf("expected recovery to be skipped, got %+v", recovered)
	}
}

// Lembang regression: a repeated input token can be swallowed by a longer
// phrase ("Lembang, Kec. Lembang" -> dict phrase "lembang lembang" from a
// foreign subdistrict), so the real district never resolves exactly. Postal
// evidence (which never enters discovery) resolved the right kelurahan; its
// entities must join the recovery neighborhood when they fit the frontier
// candidate's hierarchy, so both district and kelurahan get recovered.
func TestRecoverPostalVocabularyInNeighborhood(t *testing.T) {
	svc := newFixtureService()
	svc.districtByID[103] = &districtEntry{ID: 103, Name: "Lembang"}
	svc.hierarchyCache.DistrictToCity[103] = 10
	svc.subDistrictByID[1005] = &subDistrictEntry{ID: 1005, Name: "Lembang", PostalCode: "40391"}
	svc.hierarchyCache.SubDistrictToDist[1005] = 103

	normalized := "no lembang lembang bandung jawa barat 40391"
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "lembang", model.Entity{ID: 900, Name: "Lembang Lembang", Level: "SUBDISTRICT"}), // phrase-swallowed foreign match, out of province
		ev(model.EvidencePlaceName, "bandung", model.Entity{ID: 10, Name: "Kota Bandung", Level: "CITY"}),
		ev(model.EvidencePlaceName, "jawa", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePlaceName, "barat", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePostalCode, "40391", model.Entity{ID: 1005, Name: "Lembang", Level: "SUBDISTRICT"}),
	}

	candidates := buildFixtureCandidates(svc, resolved)
	recovered, _, _ := svc.RecoverContextualEvidence(candidates, resolved, normalized, nil)
	if len(recovered) != 1 || recovered[0].Value != "lembang" {
		t.Fatalf("expected recovered \"lembang\", got %+v", recovered)
	}
	got := map[int64]bool{}
	for _, c := range recovered[0].Candidates {
		got[c.ID] = true
	}
	if !got[103] || !got[1005] {
		t.Fatalf("expected district 103 + kelurahan 1005 recovered, got %v", recovered[0].Candidates)
	}
}

// postal entities must NOT join the neighborhood when their hierarchy
// contradicts the frontier candidate.
func TestPostalVocabularyFilteredByHierarchy(t *testing.T) {
	svc := newFixtureService()
	// Cihaur Geulis (1003) sits under Cibeunying Kaler (102) -> Kota Bandung
	// (10); the frontier candidate is Majalengka city (11).
	normalized := "palasah cihaur gelis"
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "palasah", model.Entity{ID: 101, Name: "Palasah", Level: "DISTRICT"}),
		ev(model.EvidencePlaceName, "cihaur"),
		ev(model.EvidencePlaceName, "gelis"),
		ev(model.EvidencePostalCode, "40122", model.Entity{ID: 1003, Name: "Cihaur Geulis", Level: "SUBDISTRICT"}),
	}
	candidates := buildFixtureCandidates(svc, resolved)
	recovered, corrections, _ := svc.RecoverContextualEvidence(candidates, resolved, normalized, nil)
	for _, r := range recovered {
		for _, c := range r.Candidates {
			if c.ID == 1003 {
				t.Fatalf("postal entity 1003 should be filtered out, got %+v", recovered)
			}
		}
	}
	if corrections != nil && corrections["cihaur gelis"] == "Cihaur Geulis" {
		t.Fatalf("postal entity should not be recoverable here: %v", corrections)
	}
}

// Noisy real-world input: most tokens are junk, so coverage is far below the
// old 0.5 floor even though the candidate has solid support (used >= 2).
// Recovery must still fire on the well-supported candidate.
func TestRecoverOnNoisyLowCoverageInput(t *testing.T) {
	svc := newFixtureService()
	normalized := "bursa efek kantor gedung cihaur gelis cibeunying kaler bandung jawa barat"
	resolved := []model.ResolvedEvidence{
		ev(model.EvidencePlaceName, "cibeunying", model.Entity{ID: 102, Name: "Cibeunying Kaler", Level: "DISTRICT"}),
		ev(model.EvidencePlaceName, "kaler", model.Entity{ID: 102, Name: "Cibeunying Kaler", Level: "DISTRICT"}),
		ev(model.EvidencePlaceName, "bandung", model.Entity{ID: 10, Name: "Kota Bandung", Level: "CITY"}),
		ev(model.EvidencePlaceName, "jawa", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePlaceName, "barat", model.Entity{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"}),
		ev(model.EvidencePlaceName, "bursa"),
		ev(model.EvidencePlaceName, "efek"),
		ev(model.EvidencePlaceName, "kantor"),
		ev(model.EvidencePlaceName, "gedung"),
		ev(model.EvidencePlaceName, "cihaur"),
		ev(model.EvidencePlaceName, "gelis"),
	}
	candidates := buildFixtureCandidates(svc, resolved)
	recovered, corrections, _ := svc.RecoverContextualEvidence(candidates, resolved, normalized, nil)
	if len(recovered) != 1 || recovered[0].Value != "cihaur gelis" {
		t.Fatalf("expected recovered \"cihaur gelis\" on noisy input, got %+v", recovered)
	}
	if len(recovered[0].Candidates) != 1 || recovered[0].Candidates[0].ID != 1003 {
		t.Fatalf("expected Cihaur Geulis entity, got %+v", recovered[0].Candidates)
	}
	if corrections["cihaur gelis"] != "Cihaur Geulis" {
		t.Fatalf("unexpected corrections: %v", corrections)
	}
}

func TestEvaluateCandidateCoverageIgnoresNoise(t *testing.T) {
	candidate := &model.AdminCandidate{}
	candidate.Evidence = []model.MatchedEvidence{
		{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "bandung"}, Resolved: &model.Entity{ID: 10}},
	}
	all := []model.Evidence{
		{Type: model.EvidencePlaceName, Value: "bandung"},
		{Type: model.EvidencePlaceName, Value: "no"},
		{Type: model.EvidencePlaceName, Value: "7"},
		{Type: model.EvidencePlaceName, Value: "pasir"},
	}
	cov := evaluateCandidateCoverage(candidate, all, map[string]bool{})
	if cov.Coverage != 0.5 || len(cov.UsedEvidence) != 1 || len(cov.UnusedEvidence) != 1 {
		t.Fatalf("unexpected coverage: %+v", cov)
	}
}

// buildFixtureCandidates runs the real candidate pipeline over hand-made
// resolved evidence, so coverage/frontier see exactly what production sees.
func buildFixtureCandidates(svc *Service, resolved []model.ResolvedEvidence) []model.AdminCandidate {
	candidates := svc.DiscoverCandidates(resolved, []model.DiscoveryStrategy{model.DiscoveryTopDown, model.DiscoveryAnyLevel})
	candidates = DeduplicateCandidates(candidates)
	candidates = svc.EnrichCandidates(candidates)
	return BuildConclusions(candidates, svc.hierarchyCache, resolved)
}

var _ = context.Background // keep import if unused later
