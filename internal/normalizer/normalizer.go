// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package normalizer

import (
	"regexp"
	"strings"
)

var adminSet = map[string]struct{}{
	"kabupaten":    {},
	"kab.":         {},
	"kab":          {},
	"kota":         {},
	"kecamatan":    {},
	"kec.":         {},
	"kec":          {},
	"kelurahan":    {},
	"kel.":         {},
	"kel":          {},
	"provinsi":     {},
	"prov.":        {},
	"prov":         {},
	"administrasi": {},
	"kepulauan":    {},
}

var rePunctuation = regexp.MustCompile(`[^a-zA-Z\s]`)
var reMultipleSpaces = regexp.MustCompile(`\s+`)
var rePostalCode = regexp.MustCompile(`\b\d{5}\b`)

func Normalize(name string) string {
	name = strings.ReplaceAll(name, "\\n", " ")
	lower := strings.ToLower(strings.TrimSpace(name))
	words := strings.Fields(lower)
	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if _, ok := adminSet[w]; !ok {
			filtered = append(filtered, w)
		}
	}
	joined := strings.Join(filtered, " ")

	postalCodes := rePostalCode.FindAllString(joined, -1)
	noPunct := rePunctuation.ReplaceAllString(joined, "")
	result := reMultipleSpaces.ReplaceAllString(noPunct, " ")
	result = strings.TrimSpace(result)

	if len(postalCodes) > 0 {
		if result != "" {
			result += " "
		}
		result += strings.Join(postalCodes, " ")
	}
	return result
}
