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

func resolveWinner(allProvinces, allCities, allDistricts, allSubDistricts []model.Candidate, hierarchy *database.HierarchyMap) (int64, int64, int64, int64, bool) {
	if len(allProvinces) == 0 {
		return 0, 0, 0, 0, false
	}

	type path struct {
		provinceID    int64
		cityID        int64
		districtID    int64
		subdistrictID int64
		score         int
	}

	provinceID := allProvinces[0].LocationID

	var bestPath *path

	for _, city := range allCities {
		if hierarchy != nil {
			if hierarchy.CityToProvince[city.LocationID] != provinceID {
				continue
			}
		}
		if len(allDistricts) == 0 {
			p := &path{provinceID: provinceID, cityID: city.LocationID, score: 2}
			if bestPath == nil || p.score > bestPath.score {
				bestPath = p
			}
			continue
		}
		for _, district := range allDistricts {
			if hierarchy != nil {
				if hierarchy.DistrictToCity[district.LocationID] != city.LocationID {
					continue
				}
			}
			if len(allSubDistricts) == 0 {
				p := &path{provinceID: provinceID, cityID: city.LocationID, districtID: district.LocationID, score: 3}
				if bestPath == nil || p.score > bestPath.score {
					bestPath = p
				}
				continue
			}
			for _, subdistrict := range allSubDistricts {
				if hierarchy != nil {
					if hierarchy.SubDistrictToDist[subdistrict.LocationID] != district.LocationID {
						continue
					}
				}
				p := &path{provinceID: provinceID, cityID: city.LocationID, districtID: district.LocationID, subdistrictID: subdistrict.LocationID, score: 4}
				if bestPath == nil || p.score > bestPath.score {
					bestPath = p
				}
			}
		}
	}

	if bestPath == nil {
		return provinceID, 0, 0, 0, true
	}
	return bestPath.provinceID, bestPath.cityID, bestPath.districtID, bestPath.subdistrictID, true
}

func (svc *Service) findCandidatesByLevels(ctx context.Context, sourceID int64, normalized string) ([]model.Candidate, []model.Candidate, []model.Candidate, []model.Candidate, error) {
	provinceCandidates, err := svc.findProvinceCandidates(ctx, sourceID, normalized)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	cityCandidates, err := svc.findCityCandidates(ctx, sourceID, normalized, provinceCandidates)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	districtCandidates, err := svc.findDistrictCandidates(ctx, sourceID, normalized, cityCandidates)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	subDistrictCandidates, err := svc.findSubDistrictCandidates(ctx, sourceID, normalized, districtCandidates)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return provinceCandidates, cityCandidates, districtCandidates, subDistrictCandidates, nil
}

func (svc *Service) findBCandidates(ctx context.Context, sourceID int64, normalized string) (
	[]model.Candidate, []model.Candidate, []model.Candidate, []model.Candidate,
	int64, int64, int64, int64, error,
) {
	bCityCandidates, bDistrictCandidates, bSubDistrictCandidates, err := svc.findCandidatesByAnyLevel(ctx, sourceID, normalized)
	if err != nil {
		return nil, nil, nil, nil, 0, 0, 0, 0, err
	}
	bProvinceCandidates, bCityCandidates, bDistrictCandidates := svc.inferProvinceCandidates(bCityCandidates, bDistrictCandidates, bSubDistrictCandidates)
	bWinnerProvinceID, bWinnerCityID, bWinnerDistrictID, bWinnerSubDistrictID, _ := resolveWinner(
		bProvinceCandidates, bCityCandidates, bDistrictCandidates, bSubDistrictCandidates, svc.hierarchyCache,
	)
	return bProvinceCandidates, bCityCandidates, bDistrictCandidates, bSubDistrictCandidates,
		bWinnerProvinceID, bWinnerCityID, bWinnerDistrictID, bWinnerSubDistrictID, nil
}

func (svc *Service) findCandidatesByAnyLevel(ctx context.Context, sourceID int64, normalized string) ([]model.Candidate, []model.Candidate, []model.Candidate, error) {
	if err := ensureCitiesLoaded(svc, ctx); err != nil {
		return nil, nil, nil, err
	}
	if err := ensureDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return nil, nil, nil, err
	}
	if err := ensureSubDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return nil, nil, nil, err
	}
	if err := ensureHierarchyLoaded(svc, ctx, sourceID); err != nil {
		return nil, nil, nil, err
	}

	districtIDToName := buildDistrictIDNameMap(svc.districtCache)
	subIDToEntry := buildSubIDEntryMap(svc.subDistrictCache)

	cityKeys := matchCandidates(svc.cityCache, sourceID, normalized)
	if len(cityKeys) > 0 {
		cities := buildCityCandidatesDFS(cityKeys, svc.cityCache)
		for _, city := range cities {
			matchedDists, matchedSubs := dfsExpandCity(city, svc.hierarchyCache, districtIDToName, subIDToEntry, normalized)
			if len(matchedDists) > 0 || len(matchedSubs) > 0 {
				return []model.Candidate{city}, matchedDists, matchedSubs, nil
			}
		}
		return []model.Candidate{cities[0]}, nil, nil, nil
	}

	districtKeys := matchCandidates(svc.districtCache, sourceID, normalized)
	if len(districtKeys) > 0 {
		districts := buildDistrictCandidatesDFS(districtKeys, svc.districtCache)
		for _, dist := range districts {
			matchedSubs := dfsExpandDistrict(dist, svc.hierarchyCache, subIDToEntry, normalized)
			if len(matchedSubs) > 0 {
				return nil, []model.Candidate{dist}, matchedSubs, nil
			}
		}
		return nil, []model.Candidate{districts[0]}, nil, nil
	}

	subKeys := matchCandidates(svc.subDistrictCache, sourceID, normalized)
	if len(subKeys) > 0 {
		return nil, nil, buildSubCandidatesDFS(subKeys, svc.subDistrictCache), nil
	}

	return nil, nil, nil, nil
}

func isNameInInput(name, normalized string) bool {
	words := strings.Fields(normalized)
	nameWords := strings.Fields(name)
	if len(nameWords) == 0 || len(words) == 0 || len(nameWords) > len(words) {
		return false
	}
	for i := 0; i <= len(words)-len(nameWords); i++ {
		match := true
		for j := 0; j < len(nameWords); j++ {
			if words[i+j] != nameWords[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func buildDistrictIDNameMap(cache map[string][]*districtEntry) map[int64]string {
	m := make(map[int64]string)
	for _, entries := range cache {
		for _, e := range entries {
			m[e.ID] = e.Name
		}
	}
	return m
}

func buildSubIDEntryMap(cache map[string][]*subDistrictEntry) map[int64]*subDistrictEntry {
	m := make(map[int64]*subDistrictEntry)
	for _, entries := range cache {
		for _, e := range entries {
			m[e.ID] = e
		}
	}
	return m
}

func buildCityCandidatesDFS(keys []string, cache map[string][]*cityEntry) []model.Candidate {
	candidates := make([]model.Candidate, 0, len(keys))
	for _, key := range keys {
		for _, entry := range cache[key] {
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
	return candidates
}

func buildDistrictCandidatesDFS(keys []string, cache map[string][]*districtEntry) []model.Candidate {
	candidates := make([]model.Candidate, 0, len(keys))
	for _, key := range keys {
		for _, entry := range cache[key] {
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
	return candidates
}

func buildSubCandidatesDFS(keys []string, cache map[string][]*subDistrictEntry) []model.Candidate {
	candidates := make([]model.Candidate, 0, len(keys))
	for _, key := range keys {
		for _, entry := range cache[key] {
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
	return candidates
}

func dfsExpandCity(city model.Candidate, hierarchy *database.HierarchyMap, districtIDToName map[int64]string, subIDToEntry map[int64]*subDistrictEntry, normalized string) ([]model.Candidate, []model.Candidate) {
	childDistIDs := hierarchy.CityChildren[city.LocationID]

	var matchedDists []model.Candidate
	for _, distID := range childDistIDs {
		name, ok := districtIDToName[distID]
		if !ok {
			continue
		}
		if isNameInInput(normalizer.Normalize(name), normalized) {
			distCand := model.Candidate{
				LocationID: distID,
				Name:       name,
				Level:      "DISTRICT",
				Score:      1.0,
				Source:     "cache",
				MatchType:  "EXACT",
			}

			subs := dfsExpandDistrict(distCand, hierarchy, subIDToEntry, normalized)
			if len(subs) > 0 {
				return []model.Candidate{distCand}, subs
			}

			matchedDists = append(matchedDists, distCand)
		}
	}

	if len(matchedDists) > 0 {
		return matchedDists, nil
	}

	for _, distID := range childDistIDs {
		childSubIDs := hierarchy.DistrictChildren[distID]
		for _, subID := range childSubIDs {
			entry, ok := subIDToEntry[subID]
			if !ok {
				continue
			}
			if isNameInInput(normalizer.Normalize(entry.Name), normalized) {
				return nil, []model.Candidate{
					{
						LocationID: entry.ID,
						Name:       entry.Name,
						Level:      "SUBDISTRICT",
						Score:      1.0,
						Source:     "cache",
						MatchType:  "EXACT",
						PostalCode: entry.PostalCode,
					},
				}
			}
		}
	}

	return nil, nil
}

func dfsExpandDistrict(dist model.Candidate, hierarchy *database.HierarchyMap, subIDToEntry map[int64]*subDistrictEntry, normalized string) []model.Candidate {
	childSubIDs := hierarchy.DistrictChildren[dist.LocationID]
	var matchedSubs []model.Candidate
	for _, subID := range childSubIDs {
		entry, ok := subIDToEntry[subID]
		if !ok {
			continue
		}
		if isNameInInput(normalizer.Normalize(entry.Name), normalized) {
			matchedSubs = append(matchedSubs, model.Candidate{
				LocationID: entry.ID,
				Name:       entry.Name,
				Level:      "SUBDISTRICT",
				Score:      1.0,
				Source:     "cache",
				MatchType:  "EXACT",
				PostalCode: entry.PostalCode,
			})
		}
	}
	return matchedSubs
}

func (svc *Service) inferProvinceCandidates(cityCands, districtCands, subCands []model.Candidate) ([]model.Candidate, []model.Candidate, []model.Candidate) {
	provinceIDToName := make(map[int64]string)
	for _, entries := range svc.provinceCache {
		for _, e := range entries {
			provinceIDToName[e.ID] = e.Name
		}
	}
	cityIDToName := make(map[int64]string)
	for _, entries := range svc.cityCache {
		for _, e := range entries {
			cityIDToName[e.ID] = e.Name
		}
	}
	districtIDToName := make(map[int64]string)
	for _, entries := range svc.districtCache {
		for _, e := range entries {
			districtIDToName[e.ID] = e.Name
		}
	}

	provinceIDs := make(map[int64]bool)
	enhancedCityIDs := make(map[int64]bool)
	enhancedDistrictIDs := make(map[int64]bool)

	for _, c := range cityCands {
		enhancedCityIDs[c.LocationID] = true
		provinceID := svc.hierarchyCache.CityToProvince[c.LocationID]
		if provinceID > 0 {
			provinceIDs[provinceID] = true
		}
	}
	for _, d := range districtCands {
		enhancedDistrictIDs[d.LocationID] = true
		cityID := svc.hierarchyCache.DistrictToCity[d.LocationID]
		if cityID > 0 {
			enhancedCityIDs[cityID] = true
			provinceID := svc.hierarchyCache.CityToProvince[cityID]
			if provinceID > 0 {
				provinceIDs[provinceID] = true
			}
		}
	}
	for _, s := range subCands {
		distID := svc.hierarchyCache.SubDistrictToDist[s.LocationID]
		if distID > 0 {
			enhancedDistrictIDs[distID] = true
			cityID := svc.hierarchyCache.DistrictToCity[distID]
			if cityID > 0 {
				enhancedCityIDs[cityID] = true
				provinceID := svc.hierarchyCache.CityToProvince[cityID]
				if provinceID > 0 {
					provinceIDs[provinceID] = true
				}
			}
		}
	}

	provinceCands := make([]model.Candidate, 0, len(provinceIDs))
	for pid := range provinceIDs {
		if name, ok := provinceIDToName[pid]; ok {
			provinceCands = append(provinceCands, model.Candidate{
				LocationID: pid,
				Name:       name,
				Level:      "PROVINCE",
				Score:      1.0,
				Source:     "inferred",
				MatchType:  "INFERRED",
			})
		}
	}

	existingCityIDs := make(map[int64]bool)
	for _, c := range cityCands {
		existingCityIDs[c.LocationID] = true
	}
	for cid := range enhancedCityIDs {
		if !existingCityIDs[cid] {
			if name, ok := cityIDToName[cid]; ok {
				cityCands = append(cityCands, model.Candidate{
					LocationID: cid,
					Name:       name,
					Level:      "CITY",
					Score:      1.0,
					Source:     "inferred",
					MatchType:  "INFERRED",
				})
			}
		}
	}

	existingDistrictIDs := make(map[int64]bool)
	for _, d := range districtCands {
		existingDistrictIDs[d.LocationID] = true
	}
	for did := range enhancedDistrictIDs {
		if !existingDistrictIDs[did] {
			if name, ok := districtIDToName[did]; ok {
				districtCands = append(districtCands, model.Candidate{
					LocationID: did,
					Name:       name,
					Level:      "DISTRICT",
					Score:      1.0,
					Source:     "inferred",
					MatchType:  "INFERRED",
				})
			}
		}
	}

	return provinceCands, cityCands, districtCands
}
