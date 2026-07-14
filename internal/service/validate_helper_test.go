package service

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Kab. Bandung", "bandung"},
		{"Kota Jakarta", "jakarta"},
		{"Kecamatan Cibeunying", "cibeunying"},
		{"Kelurahan Sukamaju", "sukamaju"},
		{"Provinsi Jawa Barat", "jawa barat"},
		{"Kab Bandung", "bandung"},
		{"KOTA BANDUNG", "bandung"},
		{"Jakarta Pusat", "jakarta pusat"},
		{"kab. bandung kota", "bandung"},
		{"", ""},
		{"Jl. Merdeka No. 10", "jl. merdeka no. 10"},
		{"Kec. Sukajadi Kota Bandung", "sukajadi bandung"},
		{"  Kabupaten   Bandung  ", "bandung"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalize(tt.input)
			if got != tt.expected {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
