package service

import (
	"context"
	"encoding/json"
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

func matchProvince(normalized string, provinces []database.ProvinceRow) string {
	var best string
	for _, p := range provinces {
		if strings.Contains(normalized, p.LowercaseNormalized) {
			if len(p.Name) > len(best) {
				best = p.Name
			}
		}
	}
	return best
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
	cache := make(map[int64][]database.ProvinceRow)
	for _, r := range rows {
		cache[r.SourceID] = append(cache[r.SourceID], r)
	}
	svc.provinceCache = cache
}

func (svc *Service) getProvinceOutput(ctx context.Context, sourceID int64, normalized string) (string, error) {
	svc.provinceOnce.Do(func() { svc.loadProvinces(ctx) })
	if svc.provinceErr != nil {
		return "", svc.provinceErr
	}
	return matchProvince(normalized, svc.provinceCache[sourceID]), nil
}
