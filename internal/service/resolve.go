// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"sort"
	"strings"

	"address-quality/internal/model"
)

func (svc *Service) ResolveEvidence(ctx context.Context, sourceID int64, evidence []model.Evidence, normalizedText string, roadTokens map[string]bool) []model.ResolvedEvidence {
	if err := ensurePhraseDictLoaded(svc, ctx); err != nil {
		return nil
	}

	wordEntities := svc.matchPhrases(sourceID, normalizedText)

	resolved := make([]model.ResolvedEvidence, 0, len(evidence))

	for _, ev := range evidence {
		re := model.ResolvedEvidence{Evidence: ev}

		switch ev.Type {
		case model.EvidencePostalCode:
			re.Candidates = svc.resolvePostalCodeEntity(ev)
		case model.EvidencePlaceName:
			re.Candidates = svc.resolvePlaceNameEntity(ev, wordEntities)
		case model.EvidenceRoadName:
			re.Candidates = svc.resolveRoadNameEntity(ev)
		}

		resolved = append(resolved, re)
	}

	svc.applyCityPriority(resolved, roadTokens)

	return resolved
}

// applyCityPriority suppresses SUBDISTRICT/DISTRICT candidates for tokens in the
// city priority set (normalized city names that also exist as subdistrict/district),
// but ONLY when the token is the sole resolved location evidence. Other place-name
// evidence or a resolved postal code pins the location and disables the priority.
// Tokens that follow a road prefix (jl/jalan/gg) are road names, not location
// evidence, and never block the priority.
func (svc *Service) applyCityPriority(resolved []model.ResolvedEvidence, roadTokens map[string]bool) {
	if len(svc.cityPrioritySet) == 0 {
		return
	}

	for i := range resolved {
		re := &resolved[i]
		if re.Evidence.Type != model.EvidencePlaceName {
			continue
		}
		cityType, isPriority := svc.cityPrioritySet[re.Evidence.Value]
		if !isPriority {
			continue
		}
		if svc.hasOtherLocationEvidence(resolved, re.Evidence.Value, roadTokens) {
			continue // other evidence pins the location — keep existing behavior
		}

		// only the ambiguous token is evidence -> suppress lower levels, keep CITY
		filtered := re.Candidates[:0]
		for _, c := range re.Candidates {
			if c.Level == "CITY" {
				filtered = append(filtered, c)
			}
		}
		re.Candidates = filtered

		// prefer the city_type city (KOTA over KABUPATEN) when ranking
		sort.SliceStable(re.Candidates, func(a, b int) bool {
			aType := cityTypeOf(re.Candidates[a].Name)
			bType := cityTypeOf(re.Candidates[b].Name)
			if aType != bType {
				return aType == cityType
			}
			return false
		})
	}
}

// hasOtherLocationEvidence reports whether any evidence other than the priority
// token resolved to a location entity. Place-name evidence and resolved postal
// codes count; road names and unresolved tokens do not.
func (svc *Service) hasOtherLocationEvidence(resolved []model.ResolvedEvidence, priorityValue string, roadTokens map[string]bool) bool {
	for _, re := range resolved {
		if re.Evidence.Value == priorityValue {
			continue // the ambiguous token itself
		}
		if len(re.Candidates) == 0 {
			continue // unresolved token (e.g. road name) is not evidence
		}
		// a token that follows a road prefix (jl/jalan/gg) is a road name,
		// not location evidence — it must not block the city priority
		if roadTokens[re.Evidence.Value] {
			continue
		}
		switch re.Evidence.Type {
		case model.EvidencePlaceName, model.EvidencePostalCode:
			return true
		}
	}
	return false
}

// cityTypeOf classifies an entity name as KOTA or KABUPATEN based on its prefix.
func cityTypeOf(name string) string {
	if strings.HasPrefix(name, "Kota ") {
		return "KOTA"
	}
	return "KABUPATEN"
}

func (svc *Service) resolvePostalCodeEntity(ev model.Evidence) []model.Entity {
	var entities []model.Entity
	for _, entries := range svc.subDistrictCache {
		for _, entry := range entries {
			if entry.PostalCode == ev.Value {
				entities = append(entities, model.Entity{
					ID:         entry.ID,
					Name:       entry.Name,
					Level:      "SUBDISTRICT",
					PostalCode: entry.PostalCode,
				})
			}
		}
	}
	return entities
}

func (svc *Service) resolvePlaceNameEntity(ev model.Evidence, wordEntities map[string][]model.Entity) []model.Entity {
	entities, ok := wordEntities[ev.Value]
	if !ok {
		return nil
	}

	seen := make(map[int64]bool)
	var unique []model.Entity
	for _, e := range entities {
		if seen[e.ID] {
			continue
		}
		seen[e.ID] = true
		unique = append(unique, e)
	}
	return unique
}

func (svc *Service) resolveRoadNameEntity(ev model.Evidence) []model.Entity {
	return nil
}

func (svc *Service) ensureEntitiesCachesLoaded(ctx context.Context, sourceID int64) error {
	if err := ensureProvincesLoaded(svc, ctx); err != nil {
		return err
	}
	if err := ensureCitiesLoaded(svc, ctx); err != nil {
		return err
	}
	if err := ensureDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return err
	}
	if err := ensureSubDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return err
	}
	if err := ensureHierarchyLoaded(svc, ctx, sourceID); err != nil {
		return err
	}
	if err := ensureCityPriorityLoaded(svc, ctx, sourceID); err != nil {
		return err
	}
	return nil
}
