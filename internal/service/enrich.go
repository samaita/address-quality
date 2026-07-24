// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"address-quality/internal/model"
	"address-quality/internal/normalizer"
)

func (svc *Service) EnrichCandidates(candidates []model.AdminCandidate) []model.AdminCandidate {
	if svc.hierarchyCache == nil {
		return candidates
	}

	for i := range candidates {
		c := &candidates[i]

		if c.Location.SubDistrict != nil && c.Location.District == nil {
			if distID, ok := svc.hierarchyCache.SubDistrictToDist[c.Location.SubDistrict.ID]; ok {
				if entry, ok2 := svc.districtByID[distID]; ok2 {
					c.Location.District = &model.District{
						ID:             entry.ID,
						Name:           entry.Name,
						NormalizedName: normalizer.Normalize(entry.Name),
					}
				}
			}
		}

		if c.Location.District != nil && c.Location.City == nil {
			if cityID, ok := svc.hierarchyCache.DistrictToCity[c.Location.District.ID]; ok {
				if entry, ok2 := svc.cityByID[cityID]; ok2 {
					c.Location.City = &model.City{
						ID:             entry.ID,
						Name:           entry.Name,
						NormalizedName: normalizer.Normalize(entry.Name),
						PostalCode:     entry.PostalCode,
					}
				}
			}
		}

		if c.Location.City != nil && c.Location.Province == nil {
			if provID, ok := svc.hierarchyCache.CityToProvince[c.Location.City.ID]; ok {
				if entry, ok2 := svc.provinceByID[provID]; ok2 {
					c.Location.Province = &model.Province{
						ID:             entry.ID,
						Name:           entry.Name,
						NormalizedName: normalizer.Normalize(entry.Name),
					}
				}
			}
		}
	}

	return candidates
}
