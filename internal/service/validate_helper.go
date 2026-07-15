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

func matchFromCache(cache map[string]*provinceEntry, sourceID int64, normalized string) *provinceEntry {
	words := strings.Fields(normalized)
	for n := len(words); n >= 1; n-- {
		for i := 0; i <= len(words)-n; i++ {
			ngram := strings.Join(words[i:i+n], " ")
			key := fmt.Sprintf("%d:%s", sourceID, ngram)
			if val, ok := cache[key]; ok {
				return val
			}
		}
	}
	return nil
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
	cache := make(map[string]*provinceEntry)
	kodeToEntry := make(map[string]*provinceEntry)
	for _, r := range rows {
		entry := &provinceEntry{ID: r.ID, Name: r.Name, Kode: r.Kode}
		key := fmt.Sprintf("%d:%s", r.SourceID, r.LowercaseNormalized)
		cache[key] = entry
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
	cache := make(map[string]*cityEntry)
	for _, r := range rows {
		key := fmt.Sprintf("%d:%s", r.SourceID, r.LowercaseNormalized)
		cache[key] = &cityEntry{
			ID:         r.ID,
			Name:       r.Name,
			Kode:       r.Kode,
			PostalCode: r.PostalCode,
		}
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

func (svc *Service) getProvinceLocation(ctx context.Context, sourceID int64, normalized string) (*provinceEntry, error) {
	svc.provinceOnce.Do(func() { svc.loadProvinces(ctx) })
	if svc.provinceErr != nil {
		return nil, svc.provinceErr
	}
	return matchFromCache(svc.provinceCache, sourceID, normalized), nil
}

func (svc *Service) getCityLocation(ctx context.Context, sourceID int64, normalized string, provinceID int64) (*model.Location, error) {
	svc.cityOnce.Do(func() { svc.loadCities(ctx) })
	if svc.cityErr != nil {
		return nil, svc.cityErr
	}

	if provinceID > 0 {
		svc.cityProvinceOnce.Do(func() { svc.loadCityProvinceMapping(ctx, sourceID) })
	}

	var entry *cityEntry
	words := strings.Fields(normalized)
	for n := len(words); n >= 1; n-- {
		for i := 0; i <= len(words)-n; i++ {
			ngram := strings.Join(words[i:i+n], " ")
			key := fmt.Sprintf("%d:%s", sourceID, ngram)
			if e, ok := svc.cityCache[key]; ok {
				if provinceID > 0 && svc.cityProvinceMap[e.ID] != provinceID {
					continue
				}
				entry = e
				break
			}
		}
		if entry != nil {
			break
		}
	}
	if entry == nil {
		return nil, nil
	}

	parts := strings.Split(entry.Kode, ".")
	provinceKodeKey := fmt.Sprintf("%d:%s", sourceID, parts[0])

	provinceName := ""
	if e, ok := svc.provinceKodeToEntry[provinceKodeKey]; ok {
		provinceName = e.Name
	}

	return &model.Location{
		Province:   provinceName,
		City:       entry.Name,
		PostalCode: entry.PostalCode,
	}, nil
}
