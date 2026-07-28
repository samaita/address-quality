// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import (
	"address-quality/internal/model"
	"log"
)

func DeduplicateCandidates(candidates []model.AdminCandidate) []model.AdminCandidate {
	if len(candidates) <= 1 {
		return candidates
	}

	seen := make(map[string]int)
	var deduped []model.AdminCandidate

	for _, c := range candidates {
		if c.Location.SubDistrict != nil {
			log.Printf("dedup %+v %+v", c.UUID, c.Location.SubDistrict.Name)
		}
		key := candidateKey(c)
		if idx, ok := seen[key]; ok {
			deduped[idx] = mergeCandidates(deduped[idx], c)
		} else {
			seen[key] = len(deduped)
			deduped = append(deduped, c)
		}
	}

	return deduped
}

func candidateKey(c model.AdminCandidate) string {
	var provID, cityID, distID, subID int64
	if c.Location.Province != nil {
		provID = c.Location.Province.ID
	}
	if c.Location.City != nil {
		cityID = c.Location.City.ID
	}
	if c.Location.District != nil {
		distID = c.Location.District.ID
	}
	if c.Location.SubDistrict != nil {
		subID = c.Location.SubDistrict.ID
	}
	return key(provID, cityID, distID, subID)
}

func key(ids ...int64) string {
	var s string
	for i, id := range ids {
		if i > 0 {
			s += ":"
		}
		s += itoa(id)
	}
	return s
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func mergeCandidates(a, b model.AdminCandidate) model.AdminCandidate {
	strategySet := make(map[model.DiscoveryStrategy]bool)
	for _, s := range a.DiscoveryStrategies {
		strategySet[s] = true
	}
	for _, s := range b.DiscoveryStrategies {
		strategySet[s] = true
	}

	merged := a
	merged.DiscoveryStrategies = nil
	for s := range strategySet {
		merged.DiscoveryStrategies = append(merged.DiscoveryStrategies, s)
	}

	evidenceSet := make(map[string]model.MatchedEvidence)
	for _, me := range a.Evidence {
		evidenceSet[evidenceKey(me)] = me
	}
	for _, me := range b.Evidence {
		evidenceSet[evidenceKey(me)] = me
	}
	merged.Evidence = nil
	for _, me := range evidenceSet {
		merged.Evidence = append(merged.Evidence, me)
	}

	return merged
}

func evidenceKey(me model.MatchedEvidence) string {
	if me.Resolved != nil {
		return me.Value + ":" + me.Resolved.Level + ":" + itoa(me.Resolved.ID)
	}
	return me.Value
}
