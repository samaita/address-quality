// SPDX-License-Identifier: BUSL-1.1
// Copyright (c) 2026 Samaita

package service

import "testing"

func TestCanonicalizeRequestBody(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"already trimmed single line", `{"address":"Sukarasa Sukasari Jawa Barat"}`, `{"address":"Sukarasa Sukasari Jawa Barat"}`},
		{"leading and trailing whitespace", "  {\"address\":\"Sukarasa\"}  ", `{"address":"Sukarasa"}`},
		{"newlines collapsed", "{\n  \"address\": \"Sukarasa\",\n  \"source_code\": \"kemendagri\"\n}", `{ "address": "Sukarasa", "source_code": "kemendagri" }`},
		{"tabs and multi spaces collapsed", "{\t\"address\":   \"Sukarasa\"}", `{ "address": "Sukarasa"}`},
		{"blank body", "   \n\t ", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := canonicalizeRequestBody(tc.in); got != tc.want {
				t.Errorf("canonicalizeRequestBody(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRequestHashStable(t *testing.T) {
	a := requestHash(canonicalizeRequestBody("{\n\"address\": \"Sukarasa Sukasari\"\n}"))
	b := requestHash(canonicalizeRequestBody(`{ "address": "Sukarasa Sukasari" }`))
	if a == "" || a != b {
		t.Errorf("request hash should be stable across whitespace variants, got %q and %q", a, b)
	}
}

func TestResponseIsNotFound(t *testing.T) {
	if responseIsNotFound(mockNotFoundResponse()) != true {
		t.Error("404 mock payload should be detected as not found")
	}
	if responseIsNotFound([]byte(`{"results":[]}`)) != true {
		t.Error("empty results payload should be detected as not found")
	}
	if responseIsNotFound([]byte(`{}`)) != true {
		t.Error("empty object payload should be detected as not found")
	}
	if responseIsNotFound([]byte(`{"results":[{"place":"places/x"}]}`)) != false {
		t.Error("results payload should not be detected as not found")
	}
}
