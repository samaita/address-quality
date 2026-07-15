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

func matchFromCache(cache map[string]string, sourceID int64, normalized string) (string, bool) {
	words := strings.Fields(normalized)
	for n := len(words); n >= 1; n-- {
		for i := 0; i <= len(words)-n; i++ {
			ngram := strings.Join(words[i:i+n], " ")
			key := fmt.Sprintf("%d:%s", sourceID, ngram)
			if val, ok := cache[key]; ok {
				return val, true
			}
		}
	}
	return "", false
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
	cache := make(map[string]string)
	kodeToName := make(map[string]string)
	for _, r := range rows {
		key := fmt.Sprintf("%d:%s", r.SourceID, r.LowercaseNormalized)
		cache[key] = r.Name
		kodeKey := fmt.Sprintf("%d:%s", r.SourceID, r.Kode)
		kodeToName[kodeKey] = r.Name
	}
	svc.provinceCache = cache
	svc.provinceKodeToName = kodeToName
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
			Name:       r.Name,
			Kode:       r.Kode,
			PostalCode: r.PostalCode,
		}
	}
	svc.cityCache = cache
}

func (svc *Service) getProvinceOutput(ctx context.Context, sourceID int64, normalized string) (string, error) {
	svc.provinceOnce.Do(func() { svc.loadProvinces(ctx) })
	if svc.provinceErr != nil {
		return "", svc.provinceErr
	}
	name, _ := matchFromCache(svc.provinceCache, sourceID, normalized)
	return name, nil
}

func (svc *Service) getCityLocation(ctx context.Context, sourceID int64, normalized string) (*model.Location, error) {
	svc.cityOnce.Do(func() { svc.loadCities(ctx) })
	if svc.cityErr != nil {
		return nil, svc.cityErr
	}

	var entry *cityEntry
	words := strings.Fields(normalized)
	for n := len(words); n >= 1; n-- {
		for i := 0; i <= len(words)-n; i++ {
			ngram := strings.Join(words[i:i+n], " ")
			key := fmt.Sprintf("%d:%s", sourceID, ngram)
			if e, ok := svc.cityCache[key]; ok {
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

	return &model.Location{
		Province:   svc.provinceKodeToName[provinceKodeKey],
		City:       entry.Name,
		PostalCode: entry.PostalCode,
	}, nil
}
