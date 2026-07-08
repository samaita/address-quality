package sanitizer

import "github.com/microcosm-cc/bluemonday"

var policy *bluemonday.Policy

func init() {
	policy = bluemonday.UGCPolicy()
}

func Sanitize(input string) string {
	return policy.Sanitize(input)
}
