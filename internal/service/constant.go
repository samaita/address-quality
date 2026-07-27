// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

const (
	WeightExactMatch       = 0.10
	WeightHierarchy        = 0.15
	WeightPostalCode       = 0.05
	WeightProvince         = 0.15
	WeightCity             = 0.12
	WeightDistrict         = 0.08
	WeightSubDistrict      = 0.05
	WeightMultiProvince    = 0.20
	WeightMultiCity        = 0.15
	WeightMultiDistrict    = 0.10
	WeightMultiSubDistrict       = 0.05
	MaxMultiEvidenceBonus        = 0.40
	WeightPostalCodePrefix3      = 0.07
	WeightPostalCodePrefix4      = 0.10
	WeightPostalCodePrefix5      = 0.20
)
