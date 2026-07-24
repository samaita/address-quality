// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"testing"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

func testBuildPathsHierarchy() *database.HierarchyMap {
	return &database.HierarchyMap{
		CityToProvince: map[int64]int64{
			29310: 1,
			34135: 1,
		},
		DistrictToCity: map[int64]int64{
			8312:  29310,
			45246: 34135,
			54904: 34135,
		},
		SubDistrictToDist: map[int64]int64{
			8320: 8312, 12973: 8312, 13020: 8312, 34258: 8312,
			57335: 8312, 58287: 8312, 59839: 8312, 61082: 8312, 80777: 8312,
			34188: 45246,
			36106: 45246, 36589: 45246, 37758: 45246, 39022: 45246,
			40714: 45246, 42241: 45246, 43778: 45246, 44021: 45246,
			45254: 54904, 48858: 54904, 49062: 54904, 49354: 54904,
			52235: 54904, 53737: 54904, 54905: 54904,
		},
	}
}

func provinces() []model.Entity {
	return []model.Entity{
		{ID: 1, Name: "Jawa Barat", Level: "PROVINCE"},
	}
}

func cities() []model.Entity {
	return []model.Entity{
		{ID: 29310, Name: "Kabupaten Bandung", Level: "CITY"},
		{ID: 34135, Name: "Kota Bandung", Level: "CITY"},
	}
}

func districts() []model.Entity {
	return []model.Entity{
		{ID: 8312, Name: "Merdeka", Level: "DISTRICT"},
		{ID: 45246, Name: "Bandung", Level: "DISTRICT"},
		{ID: 54904, Name: "Bandung", Level: "DISTRICT"},
	}
}

func subdistricts() []model.Entity {
	return []model.Entity{
		{ID: 8320, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "22153"},
		{ID: 12973, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "20154"},
		{ID: 13020, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "21135"},
		{ID: 34258, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "40113"},
		{ID: 57335, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "85362"},
		{ID: 58287, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "85884"},
		{ID: 59839, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "86681"},
		{ID: 61082, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "85225"},
		{ID: 80777, Name: "Merdeka", Level: "SUBDISTRICT", PostalCode: "97586"},
		{ID: 34188, Name: "Citarum", Level: "SUBDISTRICT", PostalCode: "40115"},
		{ID: 36106, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "54317"},
		{ID: 36589, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "54211"},
		{ID: 37758, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "57382"},
		{ID: 39022, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "57252"},
		{ID: 40714, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "59465"},
		{ID: 42241, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "51262"},
		{ID: 43778, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "52137"},
		{ID: 44021, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "55861"},
		{ID: 45254, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "66274"},
		{ID: 48858, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "61351"},
		{ID: 49062, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "61471"},
		{ID: 49354, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "64484"},
		{ID: 52235, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "69175"},
		{ID: 53737, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "42252"},
		{ID: 54905, Name: "Bandung", Level: "SUBDISTRICT", PostalCode: "42176"},
	}
}

func TestBuildFromProvince(t *testing.T) {
	b := &pathBuilder{}
	ents := provinces()
	candidates := b.buildFromProvince(ents)

	if len(candidates) != len(ents) {
		t.Fatalf("got %d candidates, want %d", len(candidates), len(ents))
	}

	for i, c := range candidates {
		if c.Location.Province == nil {
			t.Errorf("candidate[%d]: Province is nil", i)
		}
		if c.Location.Province.ID != ents[i].ID {
			t.Errorf("candidate[%d] province ID = %d, want %d", i, c.Location.Province.ID, ents[i].ID)
		}
		if c.Location.City != nil {
			t.Errorf("candidate[%d]: City should be nil", i)
		}
		if c.Location.District != nil {
			t.Errorf("candidate[%d]: District should be nil", i)
		}
		if c.Location.SubDistrict != nil {
			t.Errorf("candidate[%d]: SubDistrict should be nil", i)
		}
	}
}

func TestBuildFromCity(t *testing.T) {
	b := &pathBuilder{}
	ents := cities()
	candidates := b.buildFromCity(ents)

	if len(candidates) != len(ents) {
		t.Fatalf("got %d candidates, want %d", len(candidates), len(ents))
	}

	for i, c := range candidates {
		if c.Location.City == nil {
			t.Errorf("candidate[%d]: City is nil", i)
		}
		if c.Location.City.ID != ents[i].ID {
			t.Errorf("candidate[%d] city ID = %d, want %d", i, c.Location.City.ID, ents[i].ID)
		}
		if c.Location.Province != nil {
			t.Errorf("candidate[%d]: Province should be nil", i)
		}
		if c.Location.District != nil {
			t.Errorf("candidate[%d]: District should be nil", i)
		}
		if c.Location.SubDistrict != nil {
			t.Errorf("candidate[%d]: SubDistrict should be nil", i)
		}
	}
}

func TestBuildFromDistrict(t *testing.T) {
	b := &pathBuilder{}
	ents := districts()
	candidates := b.buildFromDistrict(ents)

	if len(candidates) != len(ents) {
		t.Fatalf("got %d candidates, want %d", len(candidates), len(ents))
	}

	for i, c := range candidates {
		if c.Location.District == nil {
			t.Errorf("candidate[%d]: District is nil", i)
		}
		if c.Location.District.ID != ents[i].ID {
			t.Errorf("candidate[%d] district ID = %d, want %d", i, c.Location.District.ID, ents[i].ID)
		}
		if c.Location.Province != nil {
			t.Errorf("candidate[%d]: Province should be nil", i)
		}
		if c.Location.City != nil {
			t.Errorf("candidate[%d]: City should be nil", i)
		}
		if c.Location.SubDistrict != nil {
			t.Errorf("candidate[%d]: SubDistrict should be nil", i)
		}
	}
}

func TestBuildFromSubdistrict(t *testing.T) {
	b := &pathBuilder{}
	ents := subdistricts()
	candidates := b.buildFromSubdistrict(ents)

	if len(candidates) != len(ents) {
		t.Fatalf("got %d candidates, want %d", len(candidates), len(ents))
	}

	for i, c := range candidates {
		if c.Location.SubDistrict == nil {
			t.Errorf("candidate[%d]: SubDistrict is nil", i)
		}
		if c.Location.SubDistrict.ID != ents[i].ID {
			t.Errorf("candidate[%d] subdistrict ID = %d, want %d", i, c.Location.SubDistrict.ID, ents[i].ID)
		}
		if c.Location.Province != nil {
			t.Errorf("candidate[%d]: Province should be nil", i)
		}
		if c.Location.City != nil {
			t.Errorf("candidate[%d]: City should be nil", i)
		}
		if c.Location.District != nil {
			t.Errorf("candidate[%d]: District should be nil", i)
		}
	}
}

func TestBuildConclusions(t *testing.T) {
	hierarchy := testBuildPathsHierarchy()

	provByID := make(map[int64]model.Province)
	for _, e := range provinces() {
		provByID[e.ID] = model.Province{ID: e.ID, Name: e.Name, NormalizedName: normalizer.Normalize(e.Name)}
	}
	cityByID := make(map[int64]model.City)
	for _, e := range cities() {
		cityByID[e.ID] = model.City{ID: e.ID, Name: e.Name, PostalCode: e.PostalCode, NormalizedName: normalizer.Normalize(e.Name)}
	}
	distByID := make(map[int64]model.District)
	for _, e := range districts() {
		distByID[e.ID] = model.District{ID: e.ID, Name: e.Name, NormalizedName: normalizer.Normalize(e.Name)}
	}

	flat := []model.AdminCandidate{}
	flat = append(flat, (&pathBuilder{}).buildFromProvince(provinces())...)
	flat = append(flat, (&pathBuilder{}).buildFromCity(cities())...)
	flat = append(flat, (&pathBuilder{}).buildFromDistrict(districts())...)
	flat = append(flat, (&pathBuilder{}).buildFromSubdistrict(subdistricts())...)

	// Enrich flat candidates bottom-up (simulates EnrichCandidates)
	for i := range flat {
		c := &flat[i]
		if c.Location.SubDistrict != nil && c.Location.District == nil {
			if distID, ok := hierarchy.SubDistrictToDist[c.Location.SubDistrict.ID]; ok {
				if d, ok2 := distByID[distID]; ok2 {
					c.Location.District = &d
				}
			}
		}
		if c.Location.District != nil && c.Location.City == nil {
			if cityID, ok := hierarchy.DistrictToCity[c.Location.District.ID]; ok {
				if cty, ok2 := cityByID[cityID]; ok2 {
					c.Location.City = &cty
				}
			}
		}
		if c.Location.City != nil && c.Location.Province == nil {
			if provID, ok := hierarchy.CityToProvince[c.Location.City.ID]; ok {
				if p, ok2 := provByID[provID]; ok2 {
					c.Location.Province = &p
				}
			}
		}
	}

	combined := BuildConclusions(flat, hierarchy, nil)

	if len(combined) == 0 {
		t.Fatal("BuildConclusions returned no candidates")
	}

	expectedLevelCounts := map[int]int{
		4: 25, 3: 3, 2: 2, 1: 1,
	}
	gotCounts := make(map[int]int)
	for _, c := range combined {
		n := countNonNil(c.Location)
		gotCounts[n]++
	}

	for levels, want := range expectedLevelCounts {
		got := gotCounts[levels]
		if got != want {
			t.Errorf("candidates with %d levels: got %d, want %d", levels, got, want)
		}
	}

	expectedTotal := 25 + 3 + 2 + 1
	if len(combined) != expectedTotal {
		t.Fatalf("total candidates: got %d, want %d", len(combined), expectedTotal)
	}

	if countNonNil(combined[0].Location) != 4 {
		t.Errorf("top candidate should have 4 levels, got %d", countNonNil(combined[0].Location))
	}

	lastIdx := len(combined) - 1
	if countNonNil(combined[lastIdx].Location) != 1 {
		t.Errorf("last candidate should have 1 level, got %d", countNonNil(combined[lastIdx].Location))
	}

	for i := 1; i < len(combined); i++ {
		if countNonNil(combined[i-1].Location) < countNonNil(combined[i].Location) {
			t.Errorf("candidates not sorted by level count desc at index %d", i)
		}
	}
}
