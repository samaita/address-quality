package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"address-quality/internal/logger"
	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

func (svc *Service) ValidateAddressV1(ctx context.Context, req *model.AddressRequest, requestID string) (*model.AddressResponse, error) {
	if err := req.Validate(svc.maxAddressLength); err != nil {
		return nil, errors.Join(ErrValidation, err)
	}

	now := time.Now().UTC()
	addressID := uuid.Must(uuid.NewV7()).String()
	sanitized := svc.sanitize(req.Address)
	normalized := normalizer.Normalize(sanitized)
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

	inputPostalCode := extractPostalCode(normalized)

	provinceCandidates, err := svc.findProvinceCandidates(ctx, sourceID, normalized)
	if err != nil {
		log.Error().Err(err).Msg("find province candidates")
		return nil, err
	}

	cityCandidates, err := svc.findCityCandidates(ctx, sourceID, normalized, provinceCandidates)
	if err != nil {
		log.Error().Err(err).Msg("find city candidates")
		return nil, err
	}

	districtCandidates, err := svc.findDistrictCandidates(ctx, sourceID, normalized, cityCandidates)
	if err != nil {
		log.Error().Err(err).Msg("find district candidates")
		return nil, err
	}

	subDistrictCandidates, err := svc.findSubDistrictCandidates(ctx, sourceID, normalized, districtCandidates)
	if err != nil {
		log.Error().Err(err).Msg("find subdistrict candidates")
		return nil, err
	}

	winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID, pathOK := resolveWinner(
		provinceCandidates, cityCandidates, districtCandidates, subDistrictCandidates, svc.hierarchyCache,
	)

	location := model.Location{}

	if winnerProvinceID > 0 {
		for _, c := range provinceCandidates {
			if c.LocationID == winnerProvinceID {
				location.Province = c.Name
				break
			}
		}
	}

	if winnerCityID > 0 {
		for _, c := range cityCandidates {
			if c.LocationID == winnerCityID {
				location.City = c.Name
				break
			}
		}
	}

	if winnerDistrictID > 0 {
		for _, c := range districtCandidates {
			if c.LocationID == winnerDistrictID {
				location.District = c.Name
				break
			}
		}
	}

	if winnerSubDistrictID > 0 {
		for _, c := range subDistrictCandidates {
			if c.LocationID == winnerSubDistrictID {
				location.SubDistrict = c.Name
				location.PostalCode = c.PostalCode
				break
			}
		}
	}

	if location == (model.Location{}) {
		if inputPostalCode != "" {
			loc, locErr := svc.locationRepo.FindByPostalCode(ctx, inputPostalCode, sourceID)
			if locErr != nil {
				log.Error().Err(locErr).Msg("find by postal code")
				return nil, locErr
			}
			if loc != nil {
				location = *loc
			}
		}
	}

	postalCodeMatched := false
	if inputPostalCode != "" && winnerSubDistrictID > 0 {
		for _, c := range subDistrictCandidates {
			if c.LocationID == winnerSubDistrictID {
				if c.PostalCode == inputPostalCode {
					postalCodeMatched = true
				}
				break
			}
		}
	}

	if winnerSubDistrictID > 0 {
		for _, c := range subDistrictCandidates {
			if c.LocationID == winnerSubDistrictID {
				postalCodeMatched = true
				break
			}
		}
	}

	confidence := calculateConfidence(
		provinceCandidates, cityCandidates, districtCandidates, subDistrictCandidates,
		winnerProvinceID, winnerCityID, winnerDistrictID, winnerSubDistrictID,
		postalCodeMatched, svc.hierarchyCache,
	)

	explainability := model.Explainability{}
	if len(provinceCandidates) > 0 {
		inputProv := ""
		reasonsProv := []string{"province_match"}
		if pathOK {
			reasonsProv = append(reasonsProv, "hierarchy_valid")
		}
		explainability.Province = buildExplainability("province", inputProv, provinceCandidates, winnerProvinceID, reasonsProv)
	}
	if len(cityCandidates) > 0 {
		inputCity := ""
		reasonsCity := []string{"city_match"}
		if pathOK {
			reasonsCity = append(reasonsCity, "parent_valid")
		}
		if winnerDistrictID > 0 {
			reasonsCity = append(reasonsCity, "district_match")
		}
		explainability.City = buildExplainability("city", inputCity, cityCandidates, winnerCityID, reasonsCity)
	}
	if len(districtCandidates) > 0 {
		inputDist := ""
		reasonsDist := []string{"district_match"}
		if pathOK {
			reasonsDist = append(reasonsDist, "parent_valid")
		}
		explainability.District = buildExplainability("district", inputDist, districtCandidates, winnerDistrictID, reasonsDist)
	}
	if len(subDistrictCandidates) > 0 {
		inputVil := ""
		reasonsVil := []string{"subdistrict_match"}
		if pathOK {
			reasonsVil = append(reasonsVil, "parent_valid")
		}
		explainability.SubDistrict = buildExplainability("subdistrict", inputVil, subDistrictCandidates, winnerSubDistrictID, reasonsVil)
	}

	log.Debug().
		Int("province_candidates", len(provinceCandidates)).
		Int("city_candidates", len(cityCandidates)).
		Int("district_candidates", len(districtCandidates)).
		Int("subdistrict_candidates", len(subDistrictCandidates)).
		Str("resolved_province", location.Province).
		Str("resolved_city", location.City).
		Str("resolved_district", location.District).
		Str("resolved_subdistrict", location.SubDistrict).
		Str("resolved_postal_code", location.PostalCode).
		Float64("confidence", confidence).
		Msg("candidate resolution")

	formParts := []string{}
	for _, s := range []string{location.SubDistrict, location.District, location.City, location.Province} {
		if s != "" {
			formParts = append(formParts, s)
		}
	}
	formattedOutput := strings.Join(formParts, ", ")
	if location.PostalCode != "" {
		formattedOutput += " " + location.PostalCode
	}

	quality := model.Quality{
		AddressID:       addressID,
		Confidence:      confidence,
		Location:        location,
		NormalizedInput: normalized,
		FormattedOutput: formattedOutput,
		LocationVersion: sourceVersion,
		LocationSource:  sourceCode,
		RawInput:        req.Address,
		Explainability:  explainability,
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
