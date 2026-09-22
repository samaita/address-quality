// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"errors"
	"math"
	"sort"
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

	if err := svc.ensureEntitiesCachesLoaded(ctx, sourceID); err != nil {
		log.Error().Err(err).Msg("load caches")
		return nil, err
	}

	evidence := ExtractEvidence(normalized)
	log.Debug().Int("evidence_count", len(evidence)).Msg("evidence extraction")

	roadTokens := detectRoadContextTokens(sanitized)
	compactMatchText := normalizer.Normalize(stripRoadContext(sanitized))
	resolved := svc.ResolveEvidence(ctx, sourceID, evidence, normalized, compactMatchText, roadTokens)

	buildCandidates := func(res []model.ResolvedEvidence) []model.AdminCandidate {
		cands := svc.DiscoverCandidates(res, []model.DiscoveryStrategy{model.DiscoveryTopDown, model.DiscoveryAnyLevel})
		cands = DeduplicateCandidates(cands)
		cands = svc.EnrichCandidates(cands)
		return BuildConclusions(cands, svc.hierarchyCache, res)
	}

	candidates := buildCandidates(resolved)

	// Second-pass contextual recovery: exact-resolution candidates provide the
	// context; unexplained input spans are fuzzy-matched against their
	// in-memory neighborhood; recovered evidence triggers a single rebuild.
	recovered, fuzzyCorrections, fuzzyExplained := svc.RecoverContextualEvidence(candidates, resolved, normalized, roadTokens)
	if len(recovered) > 0 {
		resolved = append(resolved, recovered...)
		candidates = buildCandidates(resolved)
	}
	log.Debug().Int("resolved_count", len(resolved)).Int("fuzzy_corrections", len(fuzzyCorrections)).Msg("entity resolution")

	var scored []scoredCandidate
	evalEvidence := evidenceWithoutValues(evidence, fuzzyExplained)
	for _, c := range candidates {
		eval := EvaluateCandidate(&c, svc.hierarchyCache, evalEvidence)
		scored = append(scored, scoredCandidate{candidate: c, eval: eval})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].eval.Confidence != scored[j].eval.Confidence {
			return scored[i].eval.Confidence > scored[j].eval.Confidence
		}
		iCount := countNonNil(scored[i].candidate.Location)
		jCount := countNonNil(scored[j].candidate.Location)
		if iCount != jCount {
			return iCount > jCount
		}
		return len(scored[i].eval.Conflicts) < len(scored[j].eval.Conflicts)
	})

	status := model.StatusUnknown
	if len(scored) > 0 {
		status = scored[0].eval.Status
		if len(scored) > 1 {
			top := scored[0].eval.Confidence
			second := scored[1].eval.Confidence
			if status == model.StatusValid && math.Abs(top-second) < 0.1 {
				status = model.StatusAmbiguous
			}
		}
	}

	var winner *scoredCandidate
	var eval model.CandidateEvaluation
	if len(scored) > 0 {
		winner = &scored[0]
	}
	location := resolveLocationFromCandidate(winner)

	loc, applied, err := svc.resolveLocationByPostalCode(ctx, location, normalized, sourceID)
	if err != nil {
		log.Error().Err(err).Msg("find by postal code")
		return nil, err
	}
	if applied {
		location = loc
		if winner != nil {
			winner.eval.Reasons = append(winner.eval.Reasons, model.ReasonPostalCodeLookup)
		} else {
			eval.Reasons = append(eval.Reasons, model.ReasonPostalCodeInferred)
		}
	}

	resolutionCands := make([]model.ResolutionCandidate, 0, len(scored))
	for _, s := range scored {
		if s.candidate.Location.Province != nil {
			loc := resolveLocationFromCandidate(&s)
			reasons := make([]string, len(s.eval.Reasons))
			for i, r := range s.eval.Reasons {
				reasons[i] = string(r)
			}
			resolutionCands = append(resolutionCands, model.ResolutionCandidate{
				UUID:     s.candidate.UUID,
				Score:    s.eval.Confidence,
				Location: loc,
				Reasons:  reasons,
			})
		}
	}

	formattedAddr := formatLocation(location)

	if winner != nil {
		eval = winner.eval
	}

	matchedStrs := make([]string, len(eval.Matched))
	for i, c := range eval.Matched {
		matchedStrs[i] = string(c)
	}
	missingStrs := make([]string, len(eval.Missing))
	for i, c := range eval.Missing {
		missingStrs[i] = string(c)
	}

	reasons := make([]string, len(eval.Reasons))
	for i, r := range eval.Reasons {
		reasons[i] = string(r)
	}

	unusedStrs := make([]string, len(eval.UnusedEvidence))
	for i, u := range eval.UnusedEvidence {
		unusedStrs[i] = u.Value
	}

	fuzzyMetadata := sortedFuzzyCorrections(fuzzyCorrections)

	data := model.ResponseData{
		AddressID:       addressID,
		Status:          status,
		Confidence:      eval.Confidence,
		RawInput:        req.Address,
		NormalizedInput: normalized,
		FormattedAddr:   formattedAddr,
		Location:        location,
		Assessment: model.Assessment{
			Matched:   matchedStrs,
			Missing:   missingStrs,
			Conflicts: eval.Conflicts,
			Ambiguous: unusedStrs,
		},
		Resolution: model.Resolution{
			Strategy:       reasons,
			CandidateCount: len(resolutionCands),
			Candidates:     resolutionCands,
		},
		Metadata: model.Metadata{
			LocationSource:   sourceCode,
			LocationVersion:  sourceVersion,
			FuzzyCorrections: fuzzyMetadata,
		},
	}

	log.Debug().
		Int("candidates", len(scored)).
		Str("resolved_province", location.Province).
		Str("resolved_city", location.City).
		Str("resolved_district", location.District).
		Str("resolved_subdistrict", location.SubDistrict).
		Str("resolved_postal_code", location.PostalCode).
		Float64("confidence", eval.Confidence).
		Str("status", string(status)).
		Msg("address resolution")

	record := buildAddressRecord(requestID, data, now)
	if svc.enableStoreRequest {
		if err := svc.storeQueue.Submit(record); err != nil {
			log.Warn().Err(err).Str("request_id", requestID).Msg("store queue rejected record")
		}
	}

	resp := &model.AddressResponse{
		Timestamp: now.Format(time.RFC3339),
		RequestID: requestID,
		Data:      data,
	}

	return resp, nil
}

func resolveLocationFromCandidate(winner *scoredCandidate) model.Location {
	if winner == nil {
		return model.Location{}
	}

	loc := model.Location{}
	c := winner.candidate

	if c.Location.Province != nil {
		loc.Province = c.Location.Province.Name
	}
	if c.Location.City != nil {
		loc.City = c.Location.City.Name
	}
	if c.Location.District != nil {
		loc.District = c.Location.District.Name
	}
	if c.Location.SubDistrict != nil {
		loc.SubDistrict = c.Location.SubDistrict.Name
		loc.PostalCode = c.Location.SubDistrict.PostalCode
	}
	if c.Location.PostalCode != nil {
		loc.PostalCode = c.Location.PostalCode.Code
	}

	return loc
}

func formatLocation(location model.Location) string {
	parts := []string{}
	for _, s := range []string{location.SubDistrict, location.District, location.City, location.Province} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	formatted := strings.Join(parts, ", ")
	if location.PostalCode != "" {
		formatted += " " + location.PostalCode
	}
	return formatted
}

func evidenceAsSlice(evidence []model.Evidence) []model.Evidence {
	return evidence
}

// evidenceWithoutValues drops evidence tokens consumed by contextual fuzzy
// recovery, so typo tokens do not surface as unused evidence and drag the
// winning candidate's confidence down.
func evidenceWithoutValues(evidence []model.Evidence, explained map[string]bool) []model.Evidence {
	if len(explained) == 0 {
		return evidence
	}
	filtered := make([]model.Evidence, 0, len(evidence))
	for _, ev := range evidence {
		if explained[ev.Value] {
			continue
		}
		filtered = append(filtered, ev)
	}
	return filtered
}

func sortedFuzzyCorrections(corrections map[string]string) []model.FuzzyCorrection {
	if len(corrections) == 0 {
		return nil
	}
	out := make([]model.FuzzyCorrection, 0, len(corrections))
	for from, to := range corrections {
		out = append(out, model.FuzzyCorrection{From: from, To: to})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].From < out[j].From })
	return out
}

type scoredCandidate struct {
	candidate model.AdminCandidate
	eval      model.CandidateEvaluation
}
