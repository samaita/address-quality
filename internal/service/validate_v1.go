package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

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

func (svc *Service) ValidateAddressV1(ctx context.Context, req *model.AddressRequest, requestID string) (*model.AddressResponse, error) {
	if err := req.Validate(svc.maxAddressLength); err != nil {
		return nil, errors.Join(ErrValidation, err)
	}

	now := time.Now().UTC()
	addressID := uuid.Must(uuid.NewV7()).String()
	sanitized := svc.sanitize(req.Address)
	normalized := normalize(sanitized)

	sourceCode := req.SourceCode
	if sourceCode == "" {
		sourceCode = svc.sourceCode
	}
	sourceID, sourceVersion, err := svc.locationRepo.FindSourceByCode(ctx, sourceCode)

	output := sanitized
	location := model.Location{}
	if err == nil {
		provinces, provErr := svc.locationRepo.FindProvincesBySourceID(ctx, sourceID)
		if provErr == nil {
			if matched := matchProvince(normalized, provinces); matched != "" {
				output = matched
			}
		}
		if postalCode := extractPostalCode(normalized); postalCode != "" {
			if loc, locErr := svc.locationRepo.FindByPostalCode(ctx, postalCode, sourceID); locErr == nil {
				location = *loc
			}
		}
	}

	quality := model.Quality{
		AddressID:       addressID,
		Confidence:      0.0,
		Location:        location,
		NormalizedInput: normalized,
		Output:          output,
		LocationVersion: sourceVersion,
		RawInput:        req.Address,
	}

	record := buildAddressRecord(requestID, addressID, quality, now)

	if err := svc.repo.InsertAddressRequest(ctx, record); err != nil {
		return nil, err
	}

	resp := &model.AddressResponse{
		Timestamp: now.Format(time.RFC3339),
		RequestID: requestID,
		Quality:   quality,
	}

	return resp, nil
}
