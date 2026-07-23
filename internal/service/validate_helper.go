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
	for _, r := range rows {
		entry := &provinceEntry{ID: r.ID, Name: r.Name, Kode: r.Kode}
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], entry)
		kodeKey := fmt.Sprintf("%d:%s", r.SourceID, r.Kode)
		kodeToEntry[kodeKey] = entry
	}
	svc.provinceCache = cache
	svc.provinceKodeToEntry = kodeToEntry
}

func (svc *Service) loadCities(ctx context.Context) {
	rows, err := svc.locationRepo.FindAllCities(ctx)
	if err != nil {
		svc.cityErr = err
		return
	}
	cache := make(map[string][]*cityEntry)
	for _, r := range rows {
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], &cityEntry{
			ID:         r.ID,
			Name:       r.Name,
			Kode:       r.Kode,
			PostalCode: r.PostalCode,
		})
	}
	svc.cityCache = cache
}

func (svc *Service) loadCityProvinceMapping(ctx context.Context, sourceID int64) {
	mapping, err := svc.locationRepo.LoadCityProvinceMapping(ctx, sourceID)
	if err != nil {
		svc.cityErr = err
		return
	}
	svc.cityProvinceMap = mapping
}

func (svc *Service) loadDistricts(ctx context.Context, sourceID int64) {
	rows, err := svc.locationRepo.FindAllDistricts(ctx, sourceID)
	if err != nil {
		svc.districtErr = err
		return
	}
	cache := make(map[string][]*districtEntry)
	for _, r := range rows {
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], &districtEntry{
			ID:   r.ID,
			Name: r.Name,
			Kode: r.Kode,
		})
	}
	svc.districtCache = cache
}

func (svc *Service) loadSubDistricts(ctx context.Context, sourceID int64) {
	rows, err := svc.locationRepo.FindAllSubDistricts(ctx, sourceID)
	if err != nil {
		svc.subDistrictErr = err
		return
	}
	cache := make(map[string][]*subDistrictEntry)
	for _, r := range rows {
		normalizedKey := normalizer.Normalize(r.Name)
		key := fmt.Sprintf("%d:%s", r.SourceID, normalizedKey)
		cache[key] = append(cache[key], &subDistrictEntry{
			ID:         r.ID,
			Name:       r.Name,
			Kode:       r.Kode,
			PostalCode: r.PostalCode,
		})
	}
	svc.subDistrictCache = cache
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

func extractNgramFromKey(key string) string {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return key
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

func (svc *Service) findProvinceCandidates(ctx context.Context, sourceID int64, normalized string) ([]model.Candidate, error) {
	if err := ensureProvincesLoaded(svc, ctx); err != nil {
		return nil, err
	}
	matchedKeys := matchCandidates(svc.provinceCache, sourceID, normalized)
	candidates := make([]model.Candidate, 0, len(matchedKeys))
	for _, key := range matchedKeys {
		for _, entry := range svc.provinceCache[key] {
			matchedNgram := extractNgramFromKey(key)
			matchType := "PARTIAL"
			if strings.TrimSpace(normalized) == entry.Kode || matchedNgram == normalizer.Normalize(entry.Name) {
				matchType = "EXACT"
			}
			candidates = append(candidates, model.Candidate{
				LocationID: entry.ID,
				Name:       entry.Name,
				Level:      "PROVINCE",
				Score:      1.0,
				Source:     "cache",
				MatchType:  matchType,
			})
		}
	}
	return candidates, nil
}

func (svc *Service) findCityCandidates(ctx context.Context, sourceID int64, normalized string, provinceCandidates []model.Candidate) ([]model.Candidate, error) {
	if err := ensureCitiesLoaded(svc, ctx); err != nil {
		return nil, err
	}
	if len(provinceCandidates) > 0 {
		if err := ensureHierarchyLoaded(svc, ctx, sourceID); err != nil {
			return nil, err
		}
	}
	matchedKeys := matchCandidates(svc.cityCache, sourceID, normalized)
	candidates := make([]model.Candidate, 0, len(matchedKeys))
	for _, key := range matchedKeys {
		for _, entry := range svc.cityCache[key] {
			provinceOK := false
			if len(provinceCandidates) == 0 {
				provinceOK = true
			} else {
				for _, pc := range provinceCandidates {
					if svc.hierarchyCache.CityToProvince[entry.ID] == pc.LocationID {
						provinceOK = true
						break
					}
				}
			}
			if !provinceOK {
				continue
			}
			matchedNgram := extractNgramFromKey(key)
			matchType := "PARTIAL"
			if matchedNgram == normalizer.Normalize(entry.Name) {
				matchType = "EXACT"
			}
			candidates = append(candidates, model.Candidate{
				LocationID: entry.ID,
				Name:       entry.Name,
				Level:      "CITY",
				Score:      1.0,
				Source:     "cache",
				MatchType:  matchType,
			})
		}
	}
	return candidates, nil
}

func (svc *Service) findDistrictCandidates(ctx context.Context, sourceID int64, normalized string, cityCandidates []model.Candidate) ([]model.Candidate, error) {
	if err := ensureDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return nil, err
	}
	if len(cityCandidates) > 0 {
		if err := ensureHierarchyLoaded(svc, ctx, sourceID); err != nil {
			return nil, err
		}
	}
	matchedKeys := matchCandidates(svc.districtCache, sourceID, normalized)
	candidates := make([]model.Candidate, 0, len(matchedKeys))
	for _, key := range matchedKeys {
		for _, entry := range svc.districtCache[key] {
			cityOK := false
			if len(cityCandidates) == 0 {
				cityOK = true
			} else {
				for _, cc := range cityCandidates {
					if svc.hierarchyCache.DistrictToCity[entry.ID] == cc.LocationID {
						cityOK = true
						break
					}
				}
			}
			if !cityOK {
				continue
			}
			matchedNgram := extractNgramFromKey(key)
			matchType := "PARTIAL"
			if matchedNgram == normalizer.Normalize(entry.Name) {
				matchType = "EXACT"
			}
			candidates = append(candidates, model.Candidate{
				LocationID: entry.ID,
				Name:       entry.Name,
				Level:      "DISTRICT",
				Score:      1.0,
				Source:     "cache",
				MatchType:  matchType,
			})
		}
	}
	return candidates, nil
}

func (svc *Service) findSubDistrictCandidates(ctx context.Context, sourceID int64, normalized string, districtCandidates []model.Candidate) ([]model.Candidate, error) {
	if err := ensureSubDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return nil, err
	}
	if len(districtCandidates) > 0 {
		if err := ensureHierarchyLoaded(svc, ctx, sourceID); err != nil {
			return nil, err
		}
	}
	matchedKeys := matchCandidates(svc.subDistrictCache, sourceID, normalized)
	candidates := make([]model.Candidate, 0, len(matchedKeys))
	for _, key := range matchedKeys {
		for _, entry := range svc.subDistrictCache[key] {
			distOK := false
			if len(districtCandidates) == 0 {
				distOK = true
			} else {
				for _, dc := range districtCandidates {
					if svc.hierarchyCache.SubDistrictToDist[entry.ID] == dc.LocationID {
						distOK = true
						break
					}
				}
			}
			if !distOK {
				continue
			}
			matchedNgram := extractNgramFromKey(key)
			matchType := "PARTIAL"
			if matchedNgram == normalizer.Normalize(entry.Name) {
				matchType = "EXACT"
			}
			candidates = append(candidates, model.Candidate{
				LocationID: entry.ID,
				Name:       entry.Name,
				Level:      "SUBDISTRICT",
				Score:      1.0,
				Source:     "cache",
				MatchType:  matchType,
				PostalCode: entry.PostalCode,
			})
		}
	}
	return candidates, nil
}

