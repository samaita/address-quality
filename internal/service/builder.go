// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"sort"

	"github.com/google/uuid"

	"address-quality/internal/database"
	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

type pathBuilder struct {
	hierarchy *database.HierarchyMap
	evidence  []model.ResolvedEvidence
}

func (svc *Service) DiscoverCandidates(resolved []model.ResolvedEvidence, strategies []model.DiscoveryStrategy) []model.AdminCandidate {
	b := &pathBuilder{
		hierarchy: svc.hierarchyCache,
		evidence:  resolved,
	}

	entities := b.collectUniqueEntities()

	var candidates []model.AdminCandidate
	candidates = append(candidates, b.buildFromProvince(entities["PROVINCE"])...)
	candidates = append(candidates, b.buildFromCity(entities["CITY"])...)
	candidates = append(candidates, b.buildFromDistrict(entities["DISTRICT"])...)
	candidates = append(candidates, b.buildFromSubdistrict(entities["SUBDISTRICT"])...)

	for i := range candidates {
		candidates[i].DiscoveryStrategies = strategies
		candidates[i].UUID = uuid.Must(uuid.NewV7()).String()
	}

	return candidates
}

func (b *pathBuilder) collectUniqueEntities() map[string][]model.Entity {
	seen := make(map[int64]bool)
	entities := make(map[string][]model.Entity)

	for _, re := range b.evidence {
		// postal code is the least reliable evidence, so we skip it when collecting unique entities
		// it only has reliable use at three first digit as people can easily mis-type, not remember, and/or knowing actual postal code
		if re.Evidence.Type == model.EvidencePostalCode {
			continue
		}

		for _, c := range re.Candidates {
			if seen[c.ID] {
				continue
			}
			seen[c.ID] = true
			entities[c.Level] = append(entities[c.Level], c)
		}
	}
	return entities
}

func (b *pathBuilder) buildFromProvince(entities []model.Entity) []model.AdminCandidate {
	var candidates []model.AdminCandidate
	for _, e := range entities {
		candidates = append(candidates, model.AdminCandidate{
			OriginLevel: model.OriginProvince,
			Location: model.AdminLocation{
				Province: &model.Province{ID: e.ID, Name: e.Name, NormalizedName: normalizer.Normalize(e.Name)},
			},
		})
	}
	return candidates
}

func (b *pathBuilder) buildFromCity(entities []model.Entity) []model.AdminCandidate {
	var candidates []model.AdminCandidate
	for _, e := range entities {
		candidates = append(candidates, model.AdminCandidate{
			OriginLevel: model.OriginCity,
			Location: model.AdminLocation{
				City: &model.City{ID: e.ID, Name: e.Name, PostalCode: e.PostalCode, NormalizedName: normalizer.Normalize(e.Name)},
			},
		})
	}
	return candidates
}

func (b *pathBuilder) buildFromDistrict(entities []model.Entity) []model.AdminCandidate {
	var candidates []model.AdminCandidate
	for _, e := range entities {
		candidates = append(candidates, model.AdminCandidate{
			OriginLevel: model.OriginDistrict,
			Location: model.AdminLocation{
				District: &model.District{ID: e.ID, Name: e.Name, NormalizedName: normalizer.Normalize(e.Name)},
			},
		})
	}
	return candidates
}

func (b *pathBuilder) buildFromSubdistrict(entities []model.Entity) []model.AdminCandidate {
	var candidates []model.AdminCandidate
	for _, e := range entities {
		candidates = append(candidates, model.AdminCandidate{
			OriginLevel: model.OriginSubDistrict,
			Location: model.AdminLocation{
				SubDistrict: &model.SubDistrict{ID: e.ID, Name: e.Name, PostalCode: e.PostalCode, NormalizedName: normalizer.Normalize(e.Name)},
			},
		})
	}
	return candidates
}

func BuildConclusions(flat []model.AdminCandidate, hierarchy *database.HierarchyMap, resolved []model.ResolvedEvidence) []model.AdminCandidate {
	if hierarchy == nil {
		return flat
	}

	var combined []model.AdminCandidate
	seen := make(map[string]bool)

	key := func(loc model.AdminLocation) string {
		var p, c, d, s int64
		if loc.Province != nil {
			p = loc.Province.ID
		}
		if loc.City != nil {
			c = loc.City.ID
		}
		if loc.District != nil {
			d = loc.District.ID
		}
		if loc.SubDistrict != nil {
			s = loc.SubDistrict.ID
		}
		return itoa(p) + ":" + itoa(c) + ":" + itoa(d) + ":" + itoa(s)
	}

	for _, c := range flat {
		k := key(c.Location)
		if seen[k] {
			continue
		}
		seen[k] = true
		combined = append(combined, c)
	}

	for i := range combined {
		if resolved != nil {
			combined[i].Evidence = matchEvidenceToCandidate(&combined[i], resolved)
		}
	}

	sort.Slice(combined, func(i, j int) bool {
		iCount := countNonNil(combined[i].Location)
		jCount := countNonNil(combined[j].Location)
		if iCount != jCount {
			return iCount > jCount
		}
		return false
	})

	return combined
}

func countNonNil(loc model.AdminLocation) int {
	var n int
	if loc.Province != nil {
		n++
	}
	if loc.City != nil {
		n++
	}
	if loc.District != nil {
		n++
	}
	if loc.SubDistrict != nil {
		n++
	}
	return n
}

func matchEvidenceToCandidate(candidate *model.AdminCandidate, resolved []model.ResolvedEvidence) []model.MatchedEvidence {
	var matched []model.MatchedEvidence

	for _, re := range resolved {
		for _, entity := range re.Candidates {
			if entityInCandidate(candidate, entity) {
				e := entity
				matched = append(matched, model.MatchedEvidence{
					Evidence: re.Evidence,
					Resolved: &e,
				})
			}
		}
	}

	return matched
}

func entityInCandidate(candidate *model.AdminCandidate, entity model.Entity) bool {
	switch entity.Level {
	case "PROVINCE":
		return candidate.Location.Province != nil && candidate.Location.Province.ID == entity.ID
	case "CITY":
		return candidate.Location.City != nil && candidate.Location.City.ID == entity.ID
	case "DISTRICT":
		return candidate.Location.District != nil && candidate.Location.District.ID == entity.ID
	case "SUBDISTRICT":
		return candidate.Location.SubDistrict != nil && candidate.Location.SubDistrict.ID == entity.ID
	}
	return false
}
