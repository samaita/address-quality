package service

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"

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

func (svc *Service) ValidateAddressV1(ctx context.Context, req *model.AddressRequest, requestID string, src model.SourceData) (*model.AddressResponse, error) {
	if err := req.Validate(svc.maxAddressLength); err != nil {
		return nil, errors.Join(ErrValidation, err)
	}

	now := time.Now().UTC()
	addressID := uuid.Must(uuid.NewV7()).String()
	sanitized := svc.sanitize(req.Address)

	location := model.Location{}
	sourceID, sourceVersion, err := svc.locationRepo.FindSourceByCode(ctx, src.Code)
	if err == nil {
		if postalCode := extractPostalCode(sanitized); postalCode != "" {
			if loc, locErr := svc.locationRepo.FindByPostalCode(ctx, postalCode, sourceID); locErr == nil {
				location = *loc
			}
		}
	}

	quality := model.Quality{
		AddressID:       addressID,
		Confidence:      0.0,
		Location:        location,
		NormalizedInput: sanitized,
		Output:          sanitized,
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
