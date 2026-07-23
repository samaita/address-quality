// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

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

	provinceCands, cityCands, districtCands, subCands, err := svc.findCandidatesByLevels(ctx, sourceID, normalized)
	if err != nil {
		log.Error().Err(err).Msg("find candidates by levels")
		return nil, err
	}
	provID, cityID, distID, subID, _ := resolveWinner(provinceCands, cityCands, districtCands, subCands, svc.hierarchyCache)

	bProvCands, bCityCands, bDistCands, bSubCands,
		bProvID, bCityID, bDistID, bSubID, err := svc.findBCandidates(ctx, sourceID, normalized)
	if err != nil {
		log.Error().Err(err).Msg("find B candidates")
		return nil, err
	}

	inputPostalCode := extractPostalCode(normalized)

	postalA := postalCodeMatches(subCands, subID, inputPostalCode)
	ctxA := &EvaluationContext{
		WinnerProvinceID:    provID,
		WinnerCityID:        cityID,
		WinnerDistrictID:    distID,
		WinnerSubDistrictID: subID,
		PostalCodeMatched:   postalA,
		InputPostalCode:     inputPostalCode,
		ExactMatchFound:     hasExactMatch(provinceCands, cityCands, districtCands, subCands),
	}
	evalA := EvaluateCandidate(ctxA, svc.hierarchyCache)
	evalA.Reasons = BuildExplainability(ctxA)

	postalB := postalCodeMatches(bSubCands, bSubID, inputPostalCode)
	ctxB := &EvaluationContext{
		WinnerProvinceID:    bProvID,
		WinnerCityID:        bCityID,
		WinnerDistrictID:    bDistID,
		WinnerSubDistrictID: bSubID,
		PostalCodeMatched:   postalB,
		InputPostalCode:     inputPostalCode,
		ExactMatchFound:     hasExactMatch(bProvCands, bCityCands, bDistCands, bSubCands),
	}
	evalB := EvaluateCandidate(ctxB, svc.hierarchyCache)
	evalB.Reasons = BuildExplainability(ctxB)

	useB := evalB.Confidence > evalA.Confidence && len(bProvCands) > 0
	if useB {
		provinceCands, cityCands, districtCands, subCands = bProvCands, bCityCands, bDistCands, bSubCands
		provID, cityID, distID, subID = bProvID, bCityID, bDistID, bSubID
	}

	postalMatched := postalCodeMatches(subCands, subID, inputPostalCode)
	evalCtx := &EvaluationContext{
		WinnerProvinceID:    provID,
		WinnerCityID:        cityID,
		WinnerDistrictID:    distID,
		WinnerSubDistrictID: subID,
		PostalCodeMatched:   postalMatched,
		InputPostalCode:     inputPostalCode,
		ExactMatchFound:     hasExactMatch(provinceCands, cityCands, districtCands, subCands),
	}
	eval := EvaluateCandidate(evalCtx, svc.hierarchyCache)
	eval.Reasons = BuildExplainability(evalCtx)

	location := resolveLocation(provinceCands, cityCands, districtCands, subCands, provID, cityID, distID, subID)

	if location == (model.Location{}) && inputPostalCode != "" {
		loc, locErr := svc.locationRepo.FindByPostalCode(ctx, inputPostalCode, sourceID)
		if locErr != nil {
			log.Error().Err(locErr).Msg("find by postal code")
			return nil, locErr
		}
		if loc != nil {
			location = *loc
		}
	}

	var resolutionCands []model.ResolutionCandidate
	if provID > 0 {
		resolutionCands = append(resolutionCands, model.ResolutionCandidate{
			Score:    eval.Confidence,
			Location: location,
			Reasons:  eval.Reasons,
		})
	}

	formParts := []string{}
	for _, s := range []string{location.SubDistrict, location.District, location.City, location.Province} {
		if s != "" {
			formParts = append(formParts, s)
		}
	}
	formattedAddr := strings.Join(formParts, ", ")
	if location.PostalCode != "" {
		formattedAddr += " " + location.PostalCode
	}

	matchedStrs := make([]string, len(eval.Matched))
	for i, c := range eval.Matched {
		matchedStrs[i] = string(c)
	}
	missingStrs := make([]string, len(eval.Missing))
	for i, c := range eval.Missing {
		missingStrs[i] = string(c)
	}

	data := model.ResponseData{
		AddressID:       addressID,
		Status:          eval.Status,
		Confidence:      eval.Confidence,
		RawInput:        req.Address,
		NormalizedInput: normalized,
		FormattedAddr:   formattedAddr,
		Location:        location,
		Assessment: model.Assessment{
			Matched:   matchedStrs,
			Missing:   missingStrs,
			Conflicts: eval.Conflicts,
			Ambiguous: []string{},
		},
		Resolution: model.Resolution{
			Strategy:       eval.Reasons,
			CandidateCount: len(resolutionCands),
			Candidates:     resolutionCands,
		},
		Metadata: model.Metadata{
			LocationSource:  sourceCode,
			LocationVersion: sourceVersion,
		},
	}

	log.Debug().
		Int("province_candidates", len(provinceCands)).
		Int("city_candidates", len(cityCands)).
		Int("district_candidates", len(districtCands)).
		Int("subdistrict_candidates", len(subCands)).
		Str("resolved_province", location.Province).
		Str("resolved_city", location.City).
		Str("resolved_district", location.District).
		Str("resolved_subdistrict", location.SubDistrict).
		Str("resolved_postal_code", location.PostalCode).
		Float64("confidence", eval.Confidence).
		Str("status", string(eval.Status)).
		Msg("candidate resolution")

	record := buildAddressRecord(requestID, data, now)
	if err := svc.repo.InsertAddressRequest(ctx, record); err != nil {
		return nil, err
	}

	resp := &model.AddressResponse{
		Timestamp: now.Format(time.RFC3339),
		RequestID: requestID,
		Data:      data,
	}

	return resp, nil
}

func resolveLocation(provinceCands, cityCands, districtCands, subCands []model.Candidate, provID, cityID, distID, subID int64) model.Location {
	loc := model.Location{}

	if provID > 0 {
		for _, c := range provinceCands {
			if c.LocationID == provID {
				loc.Province = c.Name
				break
			}
		}
	}

	if cityID > 0 {
		for _, c := range cityCands {
			if c.LocationID == cityID {
				loc.City = c.Name
				break
			}
		}
	}

	if distID > 0 {
		for _, c := range districtCands {
			if c.LocationID == distID {
				loc.District = c.Name
				break
			}
		}
	}

	if subID > 0 {
		for _, c := range subCands {
			if c.LocationID == subID {
				loc.SubDistrict = c.Name
				loc.PostalCode = c.PostalCode
				break
			}
		}
	}

	return loc
}

func postalCodeMatches(subCands []model.Candidate, subID int64, inputPostalCode string) bool {
	if inputPostalCode == "" || subID == 0 {
		return false
	}
	for _, c := range subCands {
		if c.LocationID == subID {
			return c.PostalCode == inputPostalCode
		}
	}
	return false
}
