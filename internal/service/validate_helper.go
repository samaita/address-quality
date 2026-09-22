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

func (svc *Service) resolveLocationByPostalCode(ctx context.Context, location model.Location, normalized string, sourceID int64) (model.Location, bool, error) {
	inputPostalCode := extractPostalCode(normalized)
	if inputPostalCode == "" {
		return location, false, nil
	}

	// No winner found — infer full location from postal code DB
	if location == (model.Location{}) {
		results, err := svc.locationRepo.FindByPostalCode(ctx, inputPostalCode, sourceID)
		if err != nil {
			return model.Location{}, false, err
		}
		if len(results) == 0 {
			return location, false, nil
		}

		districtGroups := make(map[string][]model.Location)
		for _, r := range results {
			districtGroups[r.District] = append(districtGroups[r.District], r)
		}
		var best []model.Location
		bestDistrict := ""
		for district, locs := range districtGroups {
			if len(locs) > len(best) || (len(locs) == len(best) && district < bestDistrict) {
				best = locs
				bestDistrict = district
			}
		}
		return best[0], true, nil
	}

	// Winner exists but missing sub-district/postal code — only fill postal code
	if location.PostalCode == "" && location.SubDistrict == "" {
		location.PostalCode = inputPostalCode
		return location, true, nil
	}

	return location, false, nil
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

func (svc *Service) loadCityPriority(ctx context.Context, sourceID int64) {
	rows, err := svc.locationRepo.FindAllCityPriority(ctx, sourceID)
	if err != nil {
		svc.cityPriorityErr = err
		return
	}
	set := make(map[string]string, len(rows))
	for _, r := range rows {
		set[r.LowercaseNormalized] = r.CityType
	}
	svc.cityPrioritySet = set
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

func ensureCityPriorityLoaded(svc *Service, ctx context.Context, sourceID int64) error {
	svc.cityPriorityOnce.Do(func() { svc.loadCityPriority(ctx, sourceID) })
	return svc.cityPriorityErr
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
	svc.compactPhraseDict = buildCompactPhraseDict(dict)
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

		// A repeated-word phrase ("lembang lembang") is ambiguous: it may be
		// one mention of a reduplicated official name (Bone Bone, Fak Fak),
		// but in addresses it is usually two separate mentions of the same
		// name — kelurahan and kecamatan sharing it ("lembang kec lembang").
		// Keep the phrase match, but also add the word's own matches so the
		// single-name entities stay reachable; scoring resolves the ambiguity.
		if repeatedPhrase(words[i:longestEnd]) {
			singleKey := fmt.Sprintf("%d:%s", sourceID, words[i])
			if byLevel, ok := svc.phraseDict[singleKey]; ok {
				var single []model.Entity
				for _, levelEntities := range byLevel {
					single = append(single, levelEntities...)
				}
				for k := i; k < longestEnd; k++ {
					wordEntities[words[k]] = append(wordEntities[words[k]], single...)
				}
			}
		}

		i = longestEnd
	}

	return wordEntities
}

// buildCompactPhraseDict indexes administrative names with whitespace removed.
// This is deterministic normalization, not fuzzy matching: no characters may
// be inserted, deleted, substituted, or transposed.
func buildCompactPhraseDict(dict map[string]map[string][]model.Entity) map[string]map[string][]model.Entity {
	compact := make(map[string]map[string][]model.Entity, len(dict))
	for key, byLevel := range dict {
		sep := strings.IndexByte(key, ':')
		if sep < 0 {
			continue
		}
		compactKey := key[:sep+1] + compactSpaces(key[sep+1:])
		levels := compact[compactKey]
		if levels == nil {
			levels = make(map[string][]model.Entity)
			compact[compactKey] = levels
		}
		for level, entities := range byLevel {
			levels[level] = appendUniqueEntities(levels[level], entities...)
		}
	}
	return compact
}

func appendUniqueEntities(dst []model.Entity, entities ...model.Entity) []model.Entity {
	seen := make(map[int64]bool, len(dst)+len(entities))
	for _, e := range dst {
		seen[e.ID] = true
	}
	for _, e := range entities {
		if seen[e.ID] {
			continue
		}
		seen[e.ID] = true
		dst = append(dst, e)
	}
	return dst
}

// matchCompactPhrases mirrors longest-match phrase resolution, but requires
// exact equality after whitespace removal. Results are kept separate from the
// normal phrase pass so literal, hierarchy-compatible subdistrict evidence can
// retain priority (for example Cikeruh versus Mekargalih).
func (svc *Service) matchCompactPhrases(sourceID int64, normalizedText string) map[string][]model.Entity {
	if svc.compactPhraseDict == nil {
		svc.compactPhraseDict = buildCompactPhraseDict(svc.phraseDict)
	}

	words := strings.Fields(normalizedText)
	wordEntities := make(map[string][]model.Entity)
	for i := 0; i < len(words); {
		longestEnd := -1
		var entities []model.Entity
		for j := len(words); j > i; j-- {
			spacedCandidate := strings.Join(words[i:j], " ")
			candidate := compactSpaces(spacedCandidate)
			key := fmt.Sprintf("%d:%s", sourceID, candidate)
			byLevel, ok := svc.compactPhraseDict[key]
			if !ok {
				continue
			}
			// Remove entities already reachable through the ordinary exact
			// phrase key. The compact pass is only for a real spacing change;
			// replaying a one-word exact name such as "bandung" would bypass
			// city-priority handling and introduce unrelated lower-level peers.
			exactIDs := make(map[int64]bool)
			if exactByLevel := svc.phraseDict[fmt.Sprintf("%d:%s", sourceID, spacedCandidate)]; exactByLevel != nil {
				for _, exactEntities := range exactByLevel {
					for _, e := range exactEntities {
						exactIDs[e.ID] = true
					}
				}
			}
			var spacingOnly []model.Entity
			for _, levelEntities := range byLevel {
				for _, e := range levelEntities {
					if !exactIDs[e.ID] {
						spacingOnly = appendUniqueEntities(spacingOnly, e)
					}
				}
			}
			if len(spacingOnly) == 0 {
				continue
			}
			longestEnd = j
			entities = spacingOnly
			break
		}
		if longestEnd == -1 {
			i++
			continue
		}
		for k := i; k < longestEnd; k++ {
			wordEntities[words[k]] = appendUniqueEntities(wordEntities[words[k]], entities...)
		}
		i = longestEnd
	}
	return wordEntities
}

// repeatedPhrase reports whether a matched phrase is one word repeated
// ("lembang lembang", "fak fak").
func repeatedPhrase(words []string) bool {
	for _, w := range words {
		if w != words[0] {
			return false
		}
	}
	return len(words) > 1
}
