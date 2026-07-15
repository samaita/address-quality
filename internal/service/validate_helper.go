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

func normalize(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	words := strings.Fields(lower)
	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if _, ok := abbreviationSet[w]; !ok {
			filtered = append(filtered, w)
		}
	}
	return strings.Join(filtered, " ")
}

func buildAddressRecord(requestID, addressID string, quality model.Quality, now time.Time) *database.AddressRecord {
	outputJSON, _ := json.Marshal(quality)

	return &database.AddressRecord{
		ID:              requestID,
		AddressID:       addressID,
		RawInput:        quality.RawInput,
		NormalizedAddr:  quality.Output,
		Confidence:      quality.Confidence,
		PostalCode:      quality.Location.PostalCode,
		SubDistrict:     quality.Location.SubDistrict,
		District:        quality.Location.District,
		City:            quality.Location.City,
		Province:        quality.Location.Province,
		LocationVersion: quality.LocationVersion,
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
		normalizedKey := normalize(r.Name)
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
		normalizedKey := normalize(r.Name)
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
		normalizedKey := normalize(r.Name)
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
		normalizedKey := normalize(r.Name)
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
			if strings.TrimSpace(normalized) == entry.Kode || matchedNgram == normalize(entry.Name) {
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
			if matchedNgram == normalize(entry.Name) {
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
			if matchedNgram == normalize(entry.Name) {
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
			if matchedNgram == normalize(entry.Name) {
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
		provinceID int64
		cityID     int64
		districtID int64
		villageID  int64
		score      int
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
			for _, village := range allSubDistricts {
				if hierarchy != nil {
					if hierarchy.SubDistrictToDist[village.LocationID] != district.LocationID {
						continue
					}
				}
				p := &path{provinceID: provinceID, cityID: city.LocationID, districtID: district.LocationID, villageID: village.LocationID, score: 4}
				if bestPath == nil || p.score > bestPath.score {
					bestPath = p
				}
			}
		}
	}

	if bestPath == nil {
		return provinceID, 0, 0, 0, true
	}
	return bestPath.provinceID, bestPath.cityID, bestPath.districtID, bestPath.villageID, true
}

func calculateConfidence(provinceCands, cityCands, districtCands, subDistrictCands []model.Candidate, winnerProvinceID, winnerCityID, winnerDistrictID, winnerVillageID int64, postalCodeMatched bool, hierarchy *database.HierarchyMap) float64 {
	var score float64

	hasExact := false
	for _, c := range provinceCands {
		if c.MatchType == "EXACT" {
			hasExact = true
			break
		}
	}
	if !hasExact {
		for _, c := range cityCands {
			if c.MatchType == "EXACT" {
				hasExact = true
				break
			}
		}
	}
	if !hasExact {
		for _, c := range districtCands {
			if c.MatchType == "EXACT" {
				hasExact = true
				break
			}
		}
	}
	if !hasExact {
		for _, c := range subDistrictCands {
			if c.MatchType == "EXACT" {
				hasExact = true
				break
			}
		}
	}
	if hasExact {
		score += 0.40
	}

	parentValid := false
	if hierarchy != nil && winnerCityID > 0 {
		provID := hierarchy.CityToProvince[winnerCityID]
		if provID == winnerProvinceID {
			parentValid = true
		}
	}
	if winnerDistrictID > 0 && parentValid && hierarchy != nil {
		cityID := hierarchy.DistrictToCity[winnerDistrictID]
		if cityID == winnerCityID {
			parentValid = true
		} else {
			parentValid = false
		}
	}
	if winnerVillageID > 0 && parentValid && hierarchy != nil {
		distID := hierarchy.SubDistrictToDist[winnerVillageID]
		if distID == winnerDistrictID {
			parentValid = true
		} else {
			parentValid = false
		}
	}
	if parentValid {
		score += 0.30
	}

	if postalCodeMatched {
		score += 0.20
	}

	if winnerProvinceID > 0 {
		score += 0.10
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func buildExplainability(level string, input string, candidates []model.Candidate, winnerID int64, reasons []string) *model.LevelExplain {
	if len(candidates) == 0 {
		return nil
	}
	candNames := make([]string, len(candidates))
	for i, c := range candidates {
		candNames[i] = c.Name
	}
	winnerName := ""
	for _, c := range candidates {
		if c.LocationID == winnerID {
			winnerName = c.Name
			break
		}
	}
	return &model.LevelExplain{
		Input:      input,
		Candidates: candNames,
		Winner:     winnerName,
		Reasons:    reasons,
	}
}
