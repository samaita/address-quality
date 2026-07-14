package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"address-quality/internal/logger"
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
	log := logger.L.With().Str("request_id", requestID).Logger()

	sourceCode := req.SourceCode
	if sourceCode == "" {
		sourceCode = svc.sourceCode
	}
	sourceID, sourceVersion, err := svc.locationRepo.FindSourceByCode(ctx, sourceCode)
	if err != nil {
		log.Error().Err(err).Msg("find source by code")
		return nil, err
	}

	output := sanitized
	location := model.Location{}

	provinces, provErr := svc.locationRepo.FindProvincesBySourceID(ctx, sourceID)
	if provErr != nil {
		log.Error().Err(provErr).Msg("find provinces by source")
		return nil, provErr
	}
	if matched := matchProvince(normalized, provinces); matched != "" {
		output = matched
	}

	if postalCode := extractPostalCode(normalized); postalCode != "" {
		loc, locErr := svc.locationRepo.FindByPostalCode(ctx, postalCode, sourceID)
		if locErr != nil {
			log.Error().Err(locErr).Msg("find by postal code")
			return nil, locErr
		}
		location = *loc
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
