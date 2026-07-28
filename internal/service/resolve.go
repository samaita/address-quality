// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"

	"address-quality/internal/model"
)

func (svc *Service) ResolveEvidence(ctx context.Context, sourceID int64, evidence []model.Evidence, normalizedText string) []model.ResolvedEvidence {
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

	return resolved
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
	return nil
}
