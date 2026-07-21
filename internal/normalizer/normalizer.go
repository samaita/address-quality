// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2025 Samaita

package normalizer

import (
	"regexp"
	"strings"
)

var adminSet = map[string]struct{}{
	"kabupaten": {},
	"kab.":      {},
	"kab":       {},
	"kota":      {},
	"kecamatan": {},
	"kec.":      {},
	"kec":       {},
	"kelurahan": {},
	"kel.":      {},
	"kel":       {},
	"provinsi":  {},
	"prov.":     {},
	"prov":      {},
}

var rePunctuation = regexp.MustCompile(`[^a-zA-Z\s]`)
var reMultipleSpaces = regexp.MustCompile(`\s+`)

func Normalize(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	words := strings.Fields(lower)
	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if _, ok := adminSet[w]; !ok {
			filtered = append(filtered, w)
		}
	}
	joined := strings.Join(filtered, " ")
	noPunct := rePunctuation.ReplaceAllString(joined, "")
	result := reMultipleSpaces.ReplaceAllString(noPunct, " ")
	return strings.TrimSpace(result)
}
