// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"context"
	"fmt"

	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

func (svc *Service) ResolveEvidence(ctx context.Context, sourceID int64, evidence []model.Evidence) []model.ResolvedEvidence {
	resolved := make([]model.ResolvedEvidence, 0, len(evidence))

	for _, ev := range evidence {
		re := model.ResolvedEvidence{Evidence: ev}

		switch ev.Type {
		case model.EvidencePostalCode:
			re.Candidates = svc.resolvePostalCodeEntity(ev)
		case model.EvidencePlaceName:
			re.Candidates = svc.resolvePlaceNameEntity(ctx, sourceID, ev)
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

func (svc *Service) resolvePlaceNameEntity(ctx context.Context, sourceID int64, ev model.Evidence) []model.Entity {
	var entities []model.Entity
	normalized := normalizer.Normalize(ev.Value)

	if err := ensureProvincesLoaded(svc, ctx); err != nil {
		return nil
	}

	provKeys := matchCandidates(svc.provinceCache, sourceID, ev.Value)
	for _, k := range provKeys {
		for _, entry := range svc.provinceCache[k] {
			entities = append(entities, model.Entity{
				ID:    entry.ID,
				Name:  entry.Name,
				Level: "PROVINCE",
			})
		}
	}

	provinceKey := fmt.Sprintf("%d:%s", sourceID, normalized)
	if entries, ok := svc.provinceCache[provinceKey]; ok {
		for _, e := range entries {
			if !entityIDExists(entities, e.ID) {
				entities = append(entities, model.Entity{
					ID:    e.ID,
					Name:  e.Name,
					Level: "PROVINCE",
				})
			}
		}
	}

	if err := ensureCitiesLoaded(svc, ctx); err != nil {
		return entities
	}

	cityKeys := matchCandidates(svc.cityCache, sourceID, ev.Value)
	for _, k := range cityKeys {
		for _, entry := range svc.cityCache[k] {
			entities = append(entities, model.Entity{
				ID:         entry.ID,
				Name:       entry.Name,
				Level:      "CITY",
				PostalCode: entry.PostalCode,
			})
		}
	}

	if err := ensureDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return entities
	}

	distKeys := matchCandidates(svc.districtCache, sourceID, ev.Value)
	for _, k := range distKeys {
		for _, entry := range svc.districtCache[k] {
			entities = append(entities, model.Entity{
				ID:    entry.ID,
				Name:  entry.Name,
				Level: "DISTRICT",
			})
		}
	}

	if err := ensureSubDistrictsLoaded(svc, ctx, sourceID); err != nil {
		return entities
	}

	subKeys := matchCandidates(svc.subDistrictCache, sourceID, ev.Value)
	for _, k := range subKeys {
		for _, entry := range svc.subDistrictCache[k] {
			entities = append(entities, model.Entity{
				ID:         entry.ID,
				Name:       entry.Name,
				Level:      "SUBDISTRICT",
				PostalCode: entry.PostalCode,
			})
		}
	}

	return entities
}

func entityIDExists(entities []model.Entity, id int64) bool {
	for _, e := range entities {
		if e.ID == id {
			return true
		}
	}
	return false
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
