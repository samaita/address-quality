package normalizer

import (
	"encoding/csv"
	"os"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Kab. Bandung", "bandung"},
		{"Kota Jakarta", "jakarta"},
		{"Kec. Cibeunying Kidul", "cibeunying kidul"},
		{"Kel. Sukamaju", "sukamaju"},
		{"Prov. Jawa Barat", "jawa barat"},
		{"KAB. BANDUNG", "bandung"},
		{"Kota Jakarta Barat", "jakarta barat"},
		{"Kab Bandung", "bandung"},
		{"  Kota   Jakarta  ", "jakarta"},
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

func TestNormalizeCSV(t *testing.T) {
	f, err := os.Open("../../tests/api/cases/address.csv")
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}

	if len(rows) < 2 {
		t.Fatal("csv has no data rows")
	}

	maxRows := 21
	if len(rows) < maxRows {
		maxRows = len(rows)
	}

	for i, row := range rows[1:maxRows] {
		if len(row) < 3 {
			t.Fatalf("row %d: expected at least 3 columns, got %d", i+2, len(row))
		}
		raw := row[2]
		got := Normalize(raw)

		words := strings.Fields(got)
		for _, w := range words {
			if _, ok := adminSet[w]; ok {
				t.Errorf("Normalize(%q) = %q, contains admin word %q", raw, got, w)
			}
		}
		if got != strings.ToLower(got) {
			t.Errorf("Normalize(%q) = %q, expected all lowercase", raw, got)
		}
	}
}
