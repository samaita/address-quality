package sanitizer

import "github.com/microcosm-cc/bluemonday"

type Sanitizer struct {
	policy *bluemonday.Policy
}

func DefaultPolicy() *bluemonday.Policy {
	return bluemonday.UGCPolicy()
}

func New(policy *bluemonday.Policy) *Sanitizer {
	return &Sanitizer{policy: policy}
}

func (s *Sanitizer) Sanitize(input string) string {
	return s.policy.Sanitize(input)
}
