package normalizer

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Kab. Bandung", "kabupaten bandung"},
		{"Kota Jakarta", "kota jakarta"},
		{"Kec. Cibeunying Kidul", "kecamatan cibeunying kidul"},
		{"Kel. Sukamaju", "kelurahan sukamaju"},
		{"Prov. Jawa Barat", "provinsi jawa barat"},
		{"KAB. BANDUNG", "kabupaten bandung"},
		{"Kota Jakarta Barat", "kota jakarta barat"},
		{"Kab Bandung", "kabupaten bandung"},
		{"  Kota   Jakarta  ", "kota jakarta"},
		{"Jakarta", "jakarta"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Normalize(tt.input)
			if got != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
