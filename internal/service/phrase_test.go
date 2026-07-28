// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"reflect"
	"testing"

	"address-quality/internal/model"
)

func phraseTestService() *Service {
	return &Service{
		phraseDict: map[string]map[string][]model.Entity{
			"1:cibeunying kaler": {
				"DISTRICT": {
					{ID: 100, Name: "Cibeunying Kaler", Level: "DISTRICT"},
				},
			},
			"1:cibeunying kidul": {
				"DISTRICT": {
					{ID: 101, Name: "Cibeunying Kidul", Level: "DISTRICT"},
				},
			},
			"1:cibeunying": {
				"SUBDISTRICT": {
					{ID: 200, Name: "Cibeunying", Level: "SUBDISTRICT"},
				},
			},
			"1:bandung": {
				"CITY": {
					{ID: 10, Name: "Kota Bandung", Level: "CITY"},
				},
			},
			"1:bandung wetan": {
				"DISTRICT": {
					{ID: 102, Name: "Bandung Wetan", Level: "DISTRICT"},
				},
			},
			"1:aceh barat": {
				"CITY": {
					{ID: 20, Name: "Aceh Barat", Level: "CITY"},
				},
			},
			"1:aceh barat daya": {
				"CITY": {
					{ID: 21, Name: "Aceh Barat Daya", Level: "CITY"},
				},
			},
			"1:daerah istimewa yogyakarta": {
				"PROVINCE": {
					{ID: 30, Name: "Daerah Istimewa Yogyakarta", Level: "PROVINCE"},
				},
			},
			"1:yogyakarta": {
				"CITY": {
					{ID: 31, Name: "Kota Yogyakarta", Level: "CITY"},
				},
			},
			"1:multi": {
				"DISTRICT": {
					{ID: 50, Name: "Multi", Level: "DISTRICT"},
				},
				"SUBDISTRICT": {
					{ID: 51, Name: "Multi", Level: "SUBDISTRICT"},
				},
			},
		},
	}
}

func entityIDsEqual(got, want []int64) bool {
	if len(got) != len(want) {
		return false
	}
	gotSet := make(map[int64]bool)
	for _, id := range got {
		gotSet[id] = true
	}
	for _, id := range want {
		if !gotSet[id] {
			return false
		}
	}
	return true
}

func collectEntityIDs(entities []model.Entity) []int64 {
	ids := make([]int64, len(entities))
	for i, e := range entities {
		ids[i] = e.ID
	}
	return ids
}

func TestMatchPhrases_SimpleSingleWord(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "bandung")

	entities, ok := result["bandung"]
	if !ok {
		t.Fatal("expected 'bandung' to be in match result")
	}
	if !entityIDsEqual(collectEntityIDs(entities), []int64{10}) {
		t.Fatalf("expected entity ID 10, got %v", collectEntityIDs(entities))
	}
}

func TestMatchPhrases_MultiWordPhrase(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "cibeunying kaler")

	for _, word := range []string{"cibeunying", "kaler"} {
		entities, ok := result[word]
		if !ok {
			t.Fatalf("expected %q to be in match result", word)
		}
		if !entityIDsEqual(collectEntityIDs(entities), []int64{100}) {
			t.Fatalf("word %q: expected entity ID 100, got %v", word, collectEntityIDs(entities))
		}
	}
}

func TestMatchPhrases_LongestMatchWins(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "cibeunying kaler")

	entities := result["cibeunying"]
	if entities == nil {
		t.Fatal("expected 'cibeunying' to be in match result")
	}
	gotIDs := collectEntityIDs(entities)
	if entityIDsEqual(gotIDs, []int64{200}) {
		t.Fatal("'cibeunying' should NOT get subdistrict ID 200 when longer phrase 'cibeunying kaler' exists")
	}
	if !entityIDsEqual(gotIDs, []int64{100}) {
		t.Fatalf("'cibeunying' should get district ID 100 from 'cibeunying kaler', got %v", gotIDs)
	}

	for _, word := range []string{"cibeunying", "kaler"} {
		if !entityIDsEqual(collectEntityIDs(result[word]), []int64{100}) {
			t.Fatalf("both words should map to the 'cibeunying kaler' entity (ID 100)")
		}
	}
}

func TestMatchPhrases_NoMatch(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "unknownword")

	if _, ok := result["unknownword"]; ok {
		t.Fatal("expected 'unknownword' not to be in match result")
	}
}

func TestMatchPhrases_MixedMatchAndNoMatch(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "cibeunying kaler unknown bandung")

	if _, ok := result["unknown"]; ok {
		t.Fatal("expected 'unknown' not to be in match result")
	}

	entities := result["bandung"]
	if entities == nil {
		t.Fatal("expected 'bandung' to be in match result")
	}
	if !entityIDsEqual(collectEntityIDs(entities), []int64{10}) {
		t.Fatalf("expected 'bandung' entity ID 10, got %v", collectEntityIDs(entities))
	}

	for _, word := range []string{"cibeunying", "kaler"} {
		if !entityIDsEqual(collectEntityIDs(result[word]), []int64{100}) {
			t.Fatalf("expected 'cibeunying'/'kaler' entity ID 100, got %v", collectEntityIDs(result[word]))
		}
	}
}

func TestMatchPhrases_ThreeWordPhrase(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "aceh barat daya")

	for _, word := range []string{"aceh", "barat", "daya"} {
		entities, ok := result[word]
		if !ok {
			t.Fatalf("expected %q to be in match result", word)
		}
		if !entityIDsEqual(collectEntityIDs(entities), []int64{21}) {
			t.Fatalf("word %q: expected entity ID 21 (Aceh Barat Daya), got %v", word, collectEntityIDs(entities))
		}
	}
}

func TestMatchPhrases_PartialMatchOnly(t *testing.T) {
	svc := phraseTestService()
	phraseDict := map[string]map[string][]model.Entity{
		"1:aceh barat": {
			"CITY": {
				{ID: 1001, Name: "Aceh Barat", Level: "CITY"},
			},
		},
	}
	svc.phraseDict = phraseDict

	result := svc.matchPhrases(1, "aceh barat daya")

	for _, word := range []string{"aceh", "barat"} {
		entities, ok := result[word]
		if !ok {
			t.Fatalf("expected %q to be in match result", word)
		}
		if !entityIDsEqual(collectEntityIDs(entities), []int64{1001}) {
			t.Fatalf("word %q: expected entity ID 1001, got %v", word, collectEntityIDs(entities))
		}
	}

	if _, ok := result["daya"]; ok {
		t.Fatal("'daya' should NOT be in match result (not in any phrase)")
	}
}

func TestMatchPhrases_DifferentSourceID(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(2, "bandung")

	if _, ok := result["bandung"]; ok {
		t.Fatal("expected no match for sourceID 2 when phrase dict has sourceID 1 entries")
	}
}

func TestMatchPhrases_MultipleAdminLevels(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "multi")

	entities, ok := result["multi"]
	if !ok {
		t.Fatal("expected 'multi' to be in match result")
	}
	if !entityIDsEqual(collectEntityIDs(entities), []int64{50, 51}) {
		t.Fatalf("expected entity IDs 50 and 51, got %v", collectEntityIDs(entities))
	}
}

func TestMatchPhrases_FourWordPhrase(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "daerah istimewa yogyakarta")

	for _, word := range []string{"daerah", "istimewa", "yogyakarta"} {
		entities, ok := result[word]
		if !ok {
			t.Fatalf("expected %q to be in match result", word)
		}
		if !entityIDsEqual(collectEntityIDs(entities), []int64{30}) {
			t.Fatalf("word %q: expected province ID 30, got %v", word, collectEntityIDs(entities))
		}
	}

	result2 := svc.matchPhrases(1, "yogyakarta")
	entities2, ok := result2["yogyakarta"]
	if !ok {
		t.Fatal("expected 'yogyakarta' (alone) to be in match result")
	}
	if !entityIDsEqual(collectEntityIDs(entities2), []int64{31}) {
		t.Fatalf("expected yogyakarta city ID 31, got %v", collectEntityIDs(entities2))
	}
}

func TestMatchPhrases_EmptyInput(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "")

	if len(result) != 0 {
		t.Fatalf("expected empty result for empty input, got %v", result)
	}
}

func TestMatchPhrases_ShorterAdjacentPhrases(t *testing.T) {
	svc := phraseTestService()
	result := svc.matchPhrases(1, "cibeunying kaler cibeunying kidul")

	cibeunyingEntities := result["cibeunying"]
	if cibeunyingEntities == nil {
		t.Fatal("expected 'cibeunying' to be in match result")
	}
	gotIDs := collectEntityIDs(cibeunyingEntities)
	for _, id := range gotIDs {
		if id == 200 {
			t.Fatal("'cibeunying' should NOT be matched to subdistrict ID 200; should be part of a longer phrase")
		}
	}

	entities := result["kaler"]
	if !entityIDsEqual(collectEntityIDs(entities), []int64{100}) {
		t.Fatalf("'kaler' should get 'Cibeunying Kaler' (ID 100), got %v", collectEntityIDs(entities))
	}

	entities2 := result["kidul"]
	if !entityIDsEqual(collectEntityIDs(entities2), []int64{101}) {
		t.Fatalf("'kidul' should get 'Cibeunying Kidul' (ID 101), got %v", collectEntityIDs(entities2))
	}
}

func TestMatchPhrases_UsedIsStable(t *testing.T) {
	svc := phraseTestService()
	r1 := svc.matchPhrases(1, "cibeunying kaler bandung")
	r2 := svc.matchPhrases(1, "cibeunying kaler bandung")

	if !reflect.DeepEqual(r1, r2) {
		t.Fatal("matchPhrases should be deterministic")
	}
}
