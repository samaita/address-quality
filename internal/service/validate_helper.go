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

func (svc *Service) loadPhraseDict(ctx context.Context) {
	dict := make(map[string]map[string][]model.Entity)

	for cacheKey, entries := range svc.provinceCache {
		m, ok := dict[cacheKey]
		if !ok {
			m = make(map[string][]model.Entity)
			dict[cacheKey] = m
		}
		for _, e := range entries {
			m["PROVINCE"] = append(m["PROVINCE"], model.Entity{ID: e.ID, Name: e.Name, Level: "PROVINCE"})
		}
	}

	for cacheKey, entries := range svc.cityCache {
		m, ok := dict[cacheKey]
		if !ok {
			m = make(map[string][]model.Entity)
			dict[cacheKey] = m
		}
		for _, e := range entries {
			m["CITY"] = append(m["CITY"], model.Entity{ID: e.ID, Name: e.Name, Level: "CITY", PostalCode: e.PostalCode})
		}
	}

	for cacheKey, entries := range svc.districtCache {
		m, ok := dict[cacheKey]
		if !ok {
			m = make(map[string][]model.Entity)
			dict[cacheKey] = m
		}
		for _, e := range entries {
			m["DISTRICT"] = append(m["DISTRICT"], model.Entity{ID: e.ID, Name: e.Name, Level: "DISTRICT"})
		}
	}

	for cacheKey, entries := range svc.subDistrictCache {
		m, ok := dict[cacheKey]
		if !ok {
			m = make(map[string][]model.Entity)
			dict[cacheKey] = m
		}
		for _, e := range entries {
			m["SUBDISTRICT"] = append(m["SUBDISTRICT"], model.Entity{ID: e.ID, Name: e.Name, Level: "SUBDISTRICT", PostalCode: e.PostalCode})
		}
	}

	svc.phraseDict = dict
}

func ensurePhraseDictLoaded(svc *Service, ctx context.Context) error {
	svc.phraseDictOnce.Do(func() { svc.loadPhraseDict(ctx) })
	return svc.phraseDictErr
}

func (svc *Service) matchPhrases(sourceID int64, normalizedText string) map[string][]model.Entity {
	words := strings.Fields(normalizedText)
	wordEntities := make(map[string][]model.Entity)

	i := 0
	for i < len(words) {

		longestEnd := -1
		for j := len(words); j > i; j-- {
			candidate := strings.Join(words[i:j], " ")
			key := fmt.Sprintf("%d:%s", sourceID, candidate)
			if _, ok := svc.phraseDict[key]; ok {
				longestEnd = j
				break
			}
		}

		if longestEnd == -1 {
			i++
			continue
		}

		phrase := strings.Join(words[i:longestEnd], " ")
		key := fmt.Sprintf("%d:%s", sourceID, phrase)
		var entities []model.Entity
		if byLevel, ok := svc.phraseDict[key]; ok {
			for _, levelEntities := range byLevel {
				entities = append(entities, levelEntities...)
			}
		}

		for k := i; k < longestEnd; k++ {
			wordEntities[words[k]] = append(wordEntities[words[k]], entities...)
		}

		i = longestEnd
	}

	// log.Printf(">> %+v\n", wordEntities)

	return wordEntities
}
