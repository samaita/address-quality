// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

var postalCodePattern = regexp.MustCompile(`\b(\d{5})\b`)

func extractPostalCode(s string) string {
	matches := postalCodePattern.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (svc *Service) sanitize(input string) string {
	return svc.s.Sanitize(input)
}

func buildAddressRecord(requestID string, data model.ResponseData, now time.Time) *database.AddressRecord {
	outputJSON, _ := json.Marshal(data)

	return &database.AddressRecord{
		ID:              requestID,
		AddressID:       data.AddressID,
		RawInput:        data.RawInput,
		NormalizedAddr:  data.FormattedAddr,
		Confidence:      data.Confidence,
		PostalCode:      data.Location.PostalCode,
		SubDistrict:     data.Location.SubDistrict,
		District:        data.Location.District,
		City:            data.Location.City,
		Province:        data.Location.Province,
		LocationVersion: data.Metadata.LocationVersion,
		OutputJSON:      string(outputJSON),
		CreatedAt:       now,
	}
}

func (svc *Service) loadProvinces(ctx context.Context) {
	rows, err := svc.locationRepo.FindAllProvinces(ctx)
	if err != nil {
		svc.provinceErr = err
		return
	}
	cache := make(map[string][]*provinceEntry)
	kodeToEntry := make(map[string]*provinceEntry)
	idToEntry := make(map[int64]*provinceEntry)
	for _, r := range rows {
		entry := &provinceEntry{ID: r.ID, Name: r.Name, Kode: r.Kode}
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], entry)
		kodeKey := fmt.Sprintf("%d:%s", r.SourceID, r.Kode)
		kodeToEntry[kodeKey] = entry
		idToEntry[r.ID] = entry
	}
	svc.provinceCache = cache
	svc.provinceKodeToEntry = kodeToEntry
	svc.provinceByID = idToEntry
}

func (svc *Service) loadCities(ctx context.Context) {
	rows, err := svc.locationRepo.FindAllCities(ctx)
	if err != nil {
		svc.cityErr = err
		return
	}
	cache := make(map[string][]*cityEntry)
	idToEntry := make(map[int64]*cityEntry)
	for _, r := range rows {
		entry := &cityEntry{
			ID:         r.ID,
			Name:       r.Name,
			Kode:       r.Kode,
			PostalCode: r.PostalCode,
		}
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], entry)
		idToEntry[r.ID] = entry
	}
	svc.cityCache = cache
	svc.cityByID = idToEntry
}

func (svc *Service) loadDistricts(ctx context.Context, sourceID int64) {
	rows, err := svc.locationRepo.FindAllDistricts(ctx, sourceID)
	if err != nil {
		svc.districtErr = err
		return
	}
	cache := make(map[string][]*districtEntry)
	idToEntry := make(map[int64]*districtEntry)
	for _, r := range rows {
		entry := &districtEntry{
			ID:   r.ID,
			Name: r.Name,
			Kode: r.Kode,
		}
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], entry)
		idToEntry[r.ID] = entry
	}
	svc.districtCache = cache
	svc.districtByID = idToEntry
}

func (svc *Service) loadSubDistricts(ctx context.Context, sourceID int64) {
	rows, err := svc.locationRepo.FindAllSubDistricts(ctx, sourceID)
	if err != nil {
		svc.subDistrictErr = err
		return
	}
	cache := make(map[string][]*subDistrictEntry)
	idToEntry := make(map[int64]*subDistrictEntry)
	for _, r := range rows {
		entry := &subDistrictEntry{
			ID:         r.ID,
			Name:       r.Name,
			Kode:       r.Kode,
			PostalCode: r.PostalCode,
		}
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], entry)
		idToEntry[r.ID] = entry
	}
	svc.subDistrictCache = cache
	svc.subDistrictByID = idToEntry
}

func (svc *Service) loadHierarchy(ctx context.Context, sourceID int64) {
	h, err := svc.locationRepo.LoadFullHierarchy(ctx, sourceID)
	if err != nil {
		svc.hierarchyErr = err
		return
	}
	svc.hierarchyCache = h
}

func ensureProvincesLoaded(svc *Service, ctx context.Context) error {
	svc.provinceOnce.Do(func() { svc.loadProvinces(ctx) })
	return svc.provinceErr
}

func ensureCitiesLoaded(svc *Service, ctx context.Context) error {
	svc.cityOnce.Do(func() { svc.loadCities(ctx) })
	return svc.cityErr
}

func ensureDistrictsLoaded(svc *Service, ctx context.Context, sourceID int64) error {
	svc.districtOnce.Do(func() { svc.loadDistricts(ctx, sourceID) })
	return svc.districtErr
}

func ensureSubDistrictsLoaded(svc *Service, ctx context.Context, sourceID int64) error {
	svc.subDistrictOnce.Do(func() { svc.loadSubDistricts(ctx, sourceID) })
	return svc.subDistrictErr
}

func ensureHierarchyLoaded(svc *Service, ctx context.Context, sourceID int64) error {
	svc.hierarchyOnce.Do(func() { svc.loadHierarchy(ctx, sourceID) })
	return svc.hierarchyErr
}

func matchCandidates[T any](cache map[string]T, sourceID int64, normalized string) []string {
	var keys []string
	words := strings.Fields(normalized)
	seen := make(map[string]bool)
	for n := len(words); n >= 1; n-- {
		for i := 0; i <= len(words)-n; i++ {
			ngram := strings.Join(words[i:i+n], " ")
			key := fmt.Sprintf("%d:%s", sourceID, ngram)
			if _, ok := cache[key]; ok && !seen[key] {
				keys = append(keys, key)
				seen[key] = true
			}
		}
	}
	return keys
}



