// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"os"
	"testing"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/sanitizer"
)

// TestMatchPhrasesRepeatedWord: "lembang kec lembang" normalizes to the
// repeated span "lembang lembang". The reduplicated-name phrase match must
// not hide the single-name entities (kelurahan + kecamatan "Lembang").
func TestMatchPhrasesRepeatedWord(t *testing.T) {
	svc := &Service{phraseDict: map[string]map[string][]model.Entity{
		"1:lembang": {
			"SUBDISTRICT": {{ID: 2, Name: "Lembang", Level: "SUBDISTRICT"}},
			"DISTRICT":    {{ID: 3, Name: "Lembang", Level: "DISTRICT"}},
		},
		"1:lembang lembang": {
			"SUBDISTRICT": {{ID: 1, Name: "Lembang Lembang", Level: "SUBDISTRICT"}},
		},
		"1:bandung barat": {
			"CITY": {{ID: 4, Name: "Kabupaten Bandung Barat", Level: "CITY"}},
		},
	}}
	got := svc.matchPhrases(1, "lembang lembang bandung barat")

	ids := map[int64]bool{}
	for _, e := range got["lembang"] {
		ids[e.ID] = true
	}
	if !ids[1] {
		t.Error("reduplicated-name entity lost")
	}
	if !ids[2] || !ids[3] {
		t.Error("single-name kelurahan/district entities hidden by repeated-word phrase match")
	}
	if len(got["bandung"]) == 0 || len(got["barat"]) == 0 {
		t.Error("multi-word city phrase match broken")
	}
}

func TestRepeatedPhrase(t *testing.T) {
	if !repeatedPhrase([]string{"lembang", "lembang"}) {
		t.Error("lembang lembang is repeated")
	}
	if repeatedPhrase([]string{"lembang"}) {
		t.Error("single word is not a repeated phrase")
	}
	if repeatedPhrase([]string{"bandung", "barat"}) {
		t.Error("bandung barat is not a repeated phrase")
	}
}


// name ("lembang kec lembang"), the shared token must resolve the kelurahan
// too — without any postal code. Postal code is support data; its absence
// must not change the outcome.
func TestSameNameChildRecovery(t *testing.T) {
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

	// Lembang, Kecamatan Lembang, Kabupaten Bandung Barat — kelurahan and
	// kecamatan share the name; no postal code in the input.
	resp, err := svc.ValidateAddressV1(ctx, &model.AddressRequest{
		Address:    "Jl. Grand Hotel No.23, Lembang, Kec. Lembang, Kabupaten Bandung Barat",
		SourceCode: "kemendagri",
	}, "test")
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if resp.Data.Location.SubDistrict != "Lembang" {
		t.Fatalf("kelurahan Lembang not resolved: got %q", resp.Data.Location.SubDistrict)
	}
	if resp.Data.Location.District != "Lembang" {
		t.Fatalf("district Lembang not resolved: got %q", resp.Data.Location.District)
	}
	if resp.Data.Location.City != "Kabupaten Bandung Barat" {
		t.Fatalf("city not resolved: got %q", resp.Data.Location.City)
	}
}
