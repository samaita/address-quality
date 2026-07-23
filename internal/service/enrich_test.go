// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"testing"

	"address-quality/internal/database"
	"address-quality/internal/model"
)

func enrichTestService() *Service {
	return &Service{
		hierarchyCache: &database.HierarchyMap{
			SubDistrictToDist: map[int64]int64{
				456: 789,
				999: 888,
			},
			DistrictToCity: map[int64]int64{
				789: 123,
				888: 321,
			},
			CityToProvince: map[int64]int64{
				123: 1,
				321: 2,
			},
		},
		provinceByID: map[int64]*provinceEntry{
			1: {ID: 1, Name: "Jawa Barat"},
			2: {ID: 2, Name: "Jawa Timur"},
		},
		cityByID: map[int64]*cityEntry{
			123: {ID: 123, Name: "Kota Bandung", PostalCode: "40111"},
			321: {ID: 321, Name: "Kota Surabaya", PostalCode: "60111"},
		},
		districtByID: map[int64]*districtEntry{
			789: {ID: 789, Name: "Cicadas"},
			888: {ID: 888, Name: "Tegalsari"},
		},
	}
}

func TestEnrichCandidates_SubdistrictFillsParents(t *testing.T) {
	svc := enrichTestService()
	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			SubDistrict: &model.SubDistrict{ID: 456, Name: "Cicadas", NormalizedName: "cicadas"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if result[0].Location.District == nil || result[0].Location.District.ID != 789 {
		t.Fatal("expected district 789 to be filled")
	}
	if result[0].Location.City == nil || result[0].Location.City.ID != 123 {
		t.Fatal("expected city 123 to be filled")
	}
	if result[0].Location.Province == nil || result[0].Location.Province.ID != 1 {
		t.Fatal("expected province 1 to be filled")
	}
	if result[0].Location.District.Name != "Cicadas" {
		t.Fatalf("expected district name 'Cicadas', got %q", result[0].Location.District.Name)
	}
	if result[0].Location.City.Name != "Kota Bandung" {
		t.Fatalf("expected city name 'Kota Bandung', got %q", result[0].Location.City.Name)
	}
	if result[0].Location.Province.Name != "Jawa Barat" {
		t.Fatalf("expected province name 'Jawa Barat', got %q", result[0].Location.Province.Name)
	}
	if result[0].Location.City.PostalCode != "40111" {
		t.Fatalf("expected city postal code '40111', got %q", result[0].Location.City.PostalCode)
	}
}

func TestEnrichCandidates_DistrictFillsParents(t *testing.T) {
	svc := enrichTestService()
	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			District: &model.District{ID: 789, Name: "Cicadas", NormalizedName: "cicadas"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if result[0].Location.City == nil || result[0].Location.City.ID != 123 {
		t.Fatal("expected city 123 to be filled")
	}
	if result[0].Location.Province == nil || result[0].Location.Province.ID != 1 {
		t.Fatal("expected province 1 to be filled")
	}
	if result[0].Location.SubDistrict != nil {
		t.Fatal("subdistrict should remain nil")
	}
}

func TestEnrichCandidates_CityFillsProvince(t *testing.T) {
	svc := enrichTestService()
	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			City: &model.City{ID: 123, Name: "Kota Bandung", NormalizedName: "kota bandung"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if result[0].Location.Province == nil || result[0].Location.Province.ID != 1 {
		t.Fatal("expected province 1 to be filled")
	}
	if result[0].Location.District != nil {
		t.Fatal("district should remain nil")
	}
	if result[0].Location.SubDistrict != nil {
		t.Fatal("subdistrict should remain nil")
	}
	if result[0].Location.Province.Name != "Jawa Barat" {
		t.Fatalf("expected province name 'Jawa Barat', got %q", result[0].Location.Province.Name)
	}
}

func TestEnrichCandidates_FullyPopulatedNoChange(t *testing.T) {
	svc := enrichTestService()
	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			Province:    &model.Province{ID: 1, Name: "Jawa Barat"},
			City:        &model.City{ID: 123, Name: "Kota Bandung"},
			District:    &model.District{ID: 789, Name: "Cicadas"},
			SubDistrict: &model.SubDistrict{ID: 456, Name: "Cicadas"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if result[0].Location.Province.ID != 1 {
		t.Fatal("province should remain unchanged")
	}
	if result[0].Location.City.ID != 123 {
		t.Fatal("city should remain unchanged")
	}
	if result[0].Location.District.ID != 789 {
		t.Fatal("district should remain unchanged")
	}
	if result[0].Location.SubDistrict.ID != 456 {
		t.Fatal("subdistrict should remain unchanged")
	}
}

func TestEnrichCandidates_UnknownSubdistrictNoFill(t *testing.T) {
	svc := enrichTestService()
	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			SubDistrict: &model.SubDistrict{ID: 999999, Name: "Unknown"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if result[0].Location.District != nil {
		t.Fatal("district should remain nil for unknown subdistrict")
	}
	if result[0].Location.City != nil {
		t.Fatal("city should remain nil")
	}
	if result[0].Location.Province != nil {
		t.Fatal("province should remain nil")
	}
}

func TestEnrichCandidates_NilHierarchy(t *testing.T) {
	svc := &Service{hierarchyCache: nil}
	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			SubDistrict: &model.SubDistrict{ID: 456, Name: "Cicadas"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if len(result) != 1 {
		t.Fatal("should return same number of candidates")
	}
	if result[0].Location.District != nil {
		t.Fatal("should not enrich when hierarchy is nil")
	}
}

func TestEnrichCandidates_MultipleCandidates(t *testing.T) {
	svc := enrichTestService()
	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			SubDistrict: &model.SubDistrict{ID: 456, Name: "Cicadas", NormalizedName: "cicadas"},
		}},
		{Location: model.AdminLocation{
			SubDistrict: &model.SubDistrict{ID: 999, Name: "Tegalsari", NormalizedName: "tegalsari"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if len(result) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(result))
	}

	if result[0].Location.District == nil || result[0].Location.District.ID != 789 {
		t.Fatal("first candidate: expected district 789")
	}
	if result[0].Location.City == nil || result[0].Location.City.ID != 123 {
		t.Fatal("first candidate: expected city 123")
	}
	if result[0].Location.Province == nil || result[0].Location.Province.ID != 1 {
		t.Fatal("first candidate: expected province 1")
	}

	if result[1].Location.District == nil || result[1].Location.District.ID != 888 {
		t.Fatal("second candidate: expected district 888")
	}
	if result[1].Location.City == nil || result[1].Location.City.ID != 321 {
		t.Fatal("second candidate: expected city 321")
	}
	if result[1].Location.Province == nil || result[1].Location.Province.ID != 2 {
		t.Fatal("second candidate: expected province 2")
	}
}

func TestEnrichCandidates_CityOnlyNoProvinceInMap(t *testing.T) {
	svc := enrichTestService()
	svc.provinceByID = map[int64]*provinceEntry{} // empty, so province lookup fails

	candidates := []model.AdminCandidate{
		{Location: model.AdminLocation{
			City: &model.City{ID: 123, Name: "Kota Bandung"},
		}},
	}

	result := svc.EnrichCandidates(candidates)

	if result[0].Location.Province != nil {
		t.Fatal("province should remain nil when not found in map")
	}
}
