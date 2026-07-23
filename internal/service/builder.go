// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"address-quality/internal/database"
	"address-quality/internal/model"
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
	candidates := b.buildPaths(entities)

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

func (b *pathBuilder) buildPaths(entities map[string][]model.Entity) []model.AdminCandidate {
	var candidates []model.AdminCandidate

	provinces := entities["PROVINCE"]
	cities := entities["CITY"]
	districts := entities["DISTRICT"]
	subdistricts := entities["SUBDISTRICT"]

	if len(provinces) == 0 && len(cities) == 0 && len(districts) == 0 && len(subdistricts) == 0 {
		return nil
	}

	if len(provinces) > 0 {
		candidates = b.buildWithProvinces(provinces, cities, districts, subdistricts)
		if len(candidates) > 0 {
			return candidates
		}
	}

	candidates = b.buildWithoutProvinces(cities, districts, subdistricts)

	return candidates
}

func (b *pathBuilder) buildWithProvinces(provinces, cities, districts, subdistricts []model.Entity) []model.AdminCandidate {
	var candidates []model.AdminCandidate

	for _, prov := range provinces {
		p := &model.Province{ID: prov.ID, Name: prov.Name}
		matchedCities := filterChildren(cities, b.hierarchy.CityToProvince, prov.ID)
		if len(matchedCities) == 0 {
			candidates = append(candidates, model.AdminCandidate{
				Location: model.AdminLocation{Province: p},
			})
			continue
		}

		for _, city := range matchedCities {
			ct := &model.City{ID: city.ID, Name: city.Name, PostalCode: city.PostalCode}
			matchedDists := filterChildren(districts, b.hierarchy.DistrictToCity, city.ID)
			if len(matchedDists) == 0 {
				candidates = append(candidates, model.AdminCandidate{
					Location: model.AdminLocation{
						Province: p,
						City:     ct,
					},
				})
				continue
			}

			for _, dist := range matchedDists {
				d := &model.District{ID: dist.ID, Name: dist.Name}
				matchedSubs := filterChildren(subdistricts, b.hierarchy.SubDistrictToDist, dist.ID)
				if len(matchedSubs) == 0 {
					candidates = append(candidates, model.AdminCandidate{
						Location: model.AdminLocation{
							Province: p,
							City:     ct,
							District: d,
						},
					})
					continue
				}

				for _, sub := range matchedSubs {
					s := &model.SubDistrict{ID: sub.ID, Name: sub.Name, PostalCode: sub.PostalCode}
					candidates = append(candidates, model.AdminCandidate{
						Location: model.AdminLocation{
							Province:    p,
							City:        ct,
							District:    d,
							SubDistrict: s,
						},
					})
				}
			}
		}
	}

	return candidates
}

func (b *pathBuilder) buildWithoutProvinces(cities, districts, subdistricts []model.Entity) []model.AdminCandidate {
	var candidates []model.AdminCandidate

	if len(cities) > 0 {
		for _, city := range cities {
			ct := &model.City{ID: city.ID, Name: city.Name, PostalCode: city.PostalCode}
			p := b.inferProvince(city.ID, "CITY")
			if p != nil {
				matchedDists := filterChildren(districts, b.hierarchy.DistrictToCity, city.ID)
				if len(matchedDists) > 0 {
					for _, dist := range matchedDists {
						d := &model.District{ID: dist.ID, Name: dist.Name}
						matchedSubs := filterChildren(subdistricts, b.hierarchy.SubDistrictToDist, dist.ID)
						if len(matchedSubs) > 0 {
							for _, sub := range matchedSubs {
								s := &model.SubDistrict{ID: sub.ID, Name: sub.Name, PostalCode: sub.PostalCode}
								candidates = append(candidates, model.AdminCandidate{
									Location: model.AdminLocation{
										Province:    p,
										City:        ct,
										District:    d,
										SubDistrict: s,
									},
								})
							}
						} else {
							candidates = append(candidates, model.AdminCandidate{
								Location: model.AdminLocation{
									Province: p,
									City:     ct,
									District: d,
								},
							})
						}
					}
				} else {
					candidates = append(candidates, model.AdminCandidate{
						Location: model.AdminLocation{
							Province: p,
							City:     ct,
						},
					})
				}
			} else {
				candidates = append(candidates, model.AdminCandidate{
					Location: model.AdminLocation{
						City: ct,
					},
				})
			}
		}
		return candidates
	}

	if len(districts) > 0 {
		for _, dist := range districts {
			d := &model.District{ID: dist.ID, Name: dist.Name}
			ct := b.inferCity(dist.ID, "DISTRICT")
			p := b.inferProvince(dist.ID, "DISTRICT")
			if ct != nil {
				matchedSubs := filterChildren(subdistricts, b.hierarchy.SubDistrictToDist, dist.ID)
				if len(matchedSubs) > 0 {
					for _, sub := range matchedSubs {
						s := &model.SubDistrict{ID: sub.ID, Name: sub.Name, PostalCode: sub.PostalCode}
						if p != nil {
							candidates = append(candidates, model.AdminCandidate{
								Location: model.AdminLocation{
									Province:    p,
									City:        ct,
									District:    d,
									SubDistrict: s,
								},
							})
						} else {
							candidates = append(candidates, model.AdminCandidate{
								Location: model.AdminLocation{
									City:        ct,
									District:    d,
									SubDistrict: s,
								},
							})
						}
					}
				} else {
					if p != nil {
						candidates = append(candidates, model.AdminCandidate{
							Location: model.AdminLocation{
								Province: p,
								City:     ct,
								District: d,
							},
						})
					} else {
						candidates = append(candidates, model.AdminCandidate{
							Location: model.AdminLocation{
								City:     ct,
								District: d,
							},
						})
					}
				}
			} else {
				candidates = append(candidates, model.AdminCandidate{
					Location: model.AdminLocation{
						District: d,
					},
				})
			}
		}
		return candidates
	}

	if len(subdistricts) > 0 {
		for _, sub := range subdistricts {
			s := &model.SubDistrict{ID: sub.ID, Name: sub.Name, PostalCode: sub.PostalCode}
			d := b.inferDistrict(sub.ID)
			ct := b.inferCity(sub.ID, "SUBDISTRICT")
			p := b.inferProvince(sub.ID, "SUBDISTRICT")
			if d != nil {
				if ct != nil {
					if p != nil {
						candidates = append(candidates, model.AdminCandidate{
							Location: model.AdminLocation{
								Province:    p,
								City:        ct,
								District:    d,
								SubDistrict: s,
							},
						})
					} else {
						candidates = append(candidates, model.AdminCandidate{
							Location: model.AdminLocation{
								City:        ct,
								District:    d,
								SubDistrict: s,
							},
						})
					}
				} else {
					candidates = append(candidates, model.AdminCandidate{
						Location: model.AdminLocation{
							District:    d,
							SubDistrict: s,
						},
					})
				}
			} else {
				candidates = append(candidates, model.AdminCandidate{
					Location: model.AdminLocation{
						SubDistrict: s,
					},
				})
			}
		}
		return candidates
	}

	return nil
}

func (b *pathBuilder) inferProvince(entityID int64, level string) *model.Province {
	var cityID int64
	switch level {
	case "CITY":
		cityID = entityID
	case "DISTRICT":
		cityID = b.hierarchy.DistrictToCity[entityID]
	case "SUBDISTRICT":
		distID := b.hierarchy.SubDistrictToDist[entityID]
		cityID = b.hierarchy.DistrictToCity[distID]
	}
	if cityID == 0 {
		return nil
	}
	provID := b.hierarchy.CityToProvince[cityID]
	if provID == 0 {
		return nil
	}
	return &model.Province{ID: provID}
}

func (b *pathBuilder) inferCity(entityID int64, level string) *model.City {
	switch level {
	case "DISTRICT":
		cityID := b.hierarchy.DistrictToCity[entityID]
		if cityID == 0 {
			return nil
		}
		return &model.City{ID: cityID}
	case "SUBDISTRICT":
		distID := b.hierarchy.SubDistrictToDist[entityID]
		if distID == 0 {
			return nil
		}
		cityID := b.hierarchy.DistrictToCity[distID]
		if cityID == 0 {
			return nil
		}
		return &model.City{ID: cityID}
	}
	return nil
}

func (b *pathBuilder) inferDistrict(subID int64) *model.District {
	distID := b.hierarchy.SubDistrictToDist[subID]
	if distID == 0 {
		return nil
	}
	return &model.District{ID: distID}
}

func filterChildren(entities []model.Entity, parentMap map[int64]int64, parentID int64) []model.Entity {
	var result []model.Entity
	for _, e := range entities {
		if parentMap[e.ID] == parentID {
			result = append(result, e)
		}
	}
	return result
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
