// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2025 Samaita

package sanitizer

import "github.com/microcosm-cc/bluemonday"

// Sanitizer is a struct that holds a bluemonday.Policy for sanitizing input strings.
// The principle is to clean user input to prevent XSS attacks and other malicious content from being processed or stored.
// Do not put any business logic in this package.
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
