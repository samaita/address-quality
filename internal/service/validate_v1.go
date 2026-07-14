package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"address-quality/internal/model"
)

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
