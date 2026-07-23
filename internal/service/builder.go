// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"sort"

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
		candidates[i].Evidence = matchEvidenceToCandidate(&candidates[i], resolved)
	}

	return candidates
}

func (b *pathBuilder) collectUniqueEntities() map[string][]model.Entity {
	seen := make(map[int64]bool)
	entities := make(map[string][]model.Entity)

	for _, re := range b.evidence {
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
			Location: model.AdminLocation{
				SubDistrict: &model.SubDistrict{ID: e.ID, Name: e.Name, PostalCode: e.PostalCode, NormalizedName: normalizer.Normalize(e.Name)},
			},
		})
	}
	return candidates
}

func BuildConclusions(flat []model.AdminCandidate, hierarchy *database.HierarchyMap) []model.AdminCandidate {
	if hierarchy == nil {
		return flat
	}

	provinceSet := make(map[int64]model.Entity)
	citySet := make(map[int64]model.Entity)
	districtSet := make(map[int64]model.Entity)
	subdistrictSet := make(map[int64]model.Entity)

	for _, c := range flat {
		if c.Location.Province != nil {
			provinceSet[c.Location.Province.ID] = model.Entity{
				ID: c.Location.Province.ID, Name: c.Location.Province.Name, Level: "PROVINCE",
			}
		}
		if c.Location.City != nil {
			citySet[c.Location.City.ID] = model.Entity{
				ID: c.Location.City.ID, Name: c.Location.City.Name, Level: "CITY", PostalCode: c.Location.City.PostalCode,
			}
		}
		if c.Location.District != nil {
			districtSet[c.Location.District.ID] = model.Entity{
				ID: c.Location.District.ID, Name: c.Location.District.Name, Level: "DISTRICT",
			}
		}
		if c.Location.SubDistrict != nil {
			subdistrictSet[c.Location.SubDistrict.ID] = model.Entity{
				ID: c.Location.SubDistrict.ID, Name: c.Location.SubDistrict.Name, Level: "SUBDISTRICT", PostalCode: c.Location.SubDistrict.PostalCode,
			}
		}
	}

	var combined []model.AdminCandidate
	seen := make(map[string]bool)

	makeLoc := func(prov *model.Province, city *model.City, dist *model.District, sub *model.SubDistrict) model.AdminLocation {
		return model.AdminLocation{Province: prov, City: city, District: dist, SubDistrict: sub}
	}

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

	add := func(loc model.AdminLocation) {
		k := key(loc)
		if seen[k] {
			return
		}
		seen[k] = true
		combined = append(combined, model.AdminCandidate{Location: loc})
	}

	for _, s := range subdistrictSet {
		sub := &model.SubDistrict{ID: s.ID, Name: s.Name, PostalCode: s.PostalCode, NormalizedName: normalizer.Normalize(s.Name)}
		loc := makeLoc(nil, nil, nil, sub)

		if distID, ok := hierarchy.SubDistrictToDist[s.ID]; ok {
			if d, ok2 := districtSet[distID]; ok2 {
				loc.District = &model.District{ID: d.ID, Name: d.Name, NormalizedName: normalizer.Normalize(d.Name)}
				if cityID, ok3 := hierarchy.DistrictToCity[distID]; ok3 {
					if c, ok4 := citySet[cityID]; ok4 {
						loc.City = &model.City{ID: c.ID, Name: c.Name, PostalCode: c.PostalCode, NormalizedName: normalizer.Normalize(c.Name)}
						if provID, ok5 := hierarchy.CityToProvince[cityID]; ok5 {
							if p, ok6 := provinceSet[provID]; ok6 {
								loc.Province = &model.Province{ID: p.ID, Name: p.Name, NormalizedName: normalizer.Normalize(p.Name)}
							}
						}
					}
				}
			}
		}
		add(loc)
	}

	for _, d := range districtSet {
		loc := makeLoc(nil, nil, &model.District{ID: d.ID, Name: d.Name, NormalizedName: normalizer.Normalize(d.Name)}, nil)

		if cityID, ok := hierarchy.DistrictToCity[d.ID]; ok {
			if c, ok2 := citySet[cityID]; ok2 {
				loc.City = &model.City{ID: c.ID, Name: c.Name, PostalCode: c.PostalCode, NormalizedName: normalizer.Normalize(c.Name)}
				if provID, ok3 := hierarchy.CityToProvince[cityID]; ok3 {
					if p, ok4 := provinceSet[provID]; ok4 {
						loc.Province = &model.Province{ID: p.ID, Name: p.Name, NormalizedName: normalizer.Normalize(p.Name)}
					}
				}
			}
		}
		add(loc)

		for _, s := range subdistrictSet {
			if hierarchy.SubDistrictToDist[s.ID] == d.ID {
				subLoc := loc
				subLoc.SubDistrict = &model.SubDistrict{ID: s.ID, Name: s.Name, PostalCode: s.PostalCode, NormalizedName: normalizer.Normalize(s.Name)}
				add(subLoc)
			}
		}
	}

	for _, c := range citySet {
		loc := makeLoc(nil, &model.City{ID: c.ID, Name: c.Name, PostalCode: c.PostalCode, NormalizedName: normalizer.Normalize(c.Name)}, nil, nil)

		if provID, ok := hierarchy.CityToProvince[c.ID]; ok {
			if p, ok2 := provinceSet[provID]; ok2 {
				loc.Province = &model.Province{ID: p.ID, Name: p.Name, NormalizedName: normalizer.Normalize(p.Name)}
			}
		}
		add(loc)

		for _, d := range districtSet {
			if hierarchy.DistrictToCity[d.ID] == c.ID {
				distLoc := loc
				distLoc.District = &model.District{ID: d.ID, Name: d.Name, NormalizedName: normalizer.Normalize(d.Name)}
				add(distLoc)

				for _, s := range subdistrictSet {
					if hierarchy.SubDistrictToDist[s.ID] == d.ID {
						subLoc := distLoc
						subLoc.SubDistrict = &model.SubDistrict{ID: s.ID, Name: s.Name, PostalCode: s.PostalCode, NormalizedName: normalizer.Normalize(s.Name)}
						add(subLoc)
					}
				}
			}
		}
	}

	for _, p := range provinceSet {
		loc := makeLoc(&model.Province{ID: p.ID, Name: p.Name, NormalizedName: normalizer.Normalize(p.Name)}, nil, nil, nil)
		add(loc)

		for _, c := range citySet {
			if hierarchy.CityToProvince[c.ID] == p.ID {
				cityLoc := loc
				cityLoc.City = &model.City{ID: c.ID, Name: c.Name, PostalCode: c.PostalCode, NormalizedName: normalizer.Normalize(c.Name)}
				add(cityLoc)

				for _, d := range districtSet {
					if hierarchy.DistrictToCity[d.ID] == c.ID {
						distLoc := cityLoc
						distLoc.District = &model.District{ID: d.ID, Name: d.Name, NormalizedName: normalizer.Normalize(d.Name)}
						add(distLoc)

						for _, s := range subdistrictSet {
							if hierarchy.SubDistrictToDist[s.ID] == d.ID {
								subLoc := distLoc
								subLoc.SubDistrict = &model.SubDistrict{ID: s.ID, Name: s.Name, PostalCode: s.PostalCode, NormalizedName: normalizer.Normalize(s.Name)}
								add(subLoc)
							}
						}
					}
				}
			}
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
				break
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
