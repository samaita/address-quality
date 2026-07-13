package normalizer

import (
	"regexp"
	"strings"
)

var abbreviationMap = map[string]string{
	"kab.": "kabupaten",
	"kab":  "kabupaten",
	"kec.": "kecamatan",
	"kec":  "kecamatan",
	"kel.": "kelurahan",
	"kel":  "kelurahan",
	"prov.": "provinsi",
	"prov": "provinsi",
}

var rePunctuation = regexp.MustCompile(`[^\w\s]`)
var reMultipleSpaces = regexp.MustCompile(`\s+`)

func Normalize(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	words := strings.Fields(lower)
	for i, w := range words {
		if expanded, ok := abbreviationMap[w]; ok {
			words[i] = expanded
		}
	}
	expanded := strings.Join(words, " ")
	noPunct := rePunctuation.ReplaceAllString(expanded, "")
	result := reMultipleSpaces.ReplaceAllString(noPunct, " ")
	return strings.TrimSpace(result)
}
