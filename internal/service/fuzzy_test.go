package service

import (
	"context"
	"testing"

	"address-quality/internal/model"
)

type fakePostgresRepo struct {
	entities []model.Entity
	err      error
}

func (f *fakePostgresRepo) Ping(ctx context.Context) error { return nil }

func (f *fakePostgresRepo) FindSimilarLocations(ctx context.Context, sourceID int64, token string, limit int, minSimilarity float64) ([]model.Entity, error) {
	return f.entities, f.err
}

func kelurahanCihaurGeulis() model.Entity {
	return model.Entity{ID: 42, Name: "Cihaur Geulis", Level: "SUBDISTRICT", PostalCode: "40191"}
}

func TestAppendFuzzyEvidenceNilPostgres(t *testing.T) {
	svc := &Service{}
	resolved := []model.ResolvedEvidence{{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "cihuar"}}}
	got, corrections := svc.AppendFuzzyEvidence(context.Background(), 1, resolved)
	if len(got) != 1 || corrections != nil {
		t.Fatalf("expected no-op without postgres, got %d evidence, corrections %v", len(got), corrections)
	}
}

func TestAppendFuzzyEvidenceAddsResolvedEvidence(t *testing.T) {
	svc := &Service{postgresRepo: &fakePostgresRepo{entities: []model.Entity{kelurahanCihaurGeulis()}}}
	resolved := []model.ResolvedEvidence{
		{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "bandung"}, Candidates: []model.Entity{{ID: 1, Name: "Kota Bandung", Level: "CITY"}}},
		{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "cihuar"}},
	}
	got, corrections := svc.AppendFuzzyEvidence(context.Background(), 1, resolved)

	// original evidence untouched + one appended evidence
	if len(got) != 3 {
		t.Fatalf("expected 3 resolved evidence, got %d", len(got))
	}
	added := got[2]
	if added.Value != "Cihaur Geulis" || len(added.Candidates) != 1 || added.Candidates[0].ID != 42 {
		t.Fatalf("unexpected added evidence: %+v", added)
	}
	if got[1].Value != "cihuar" || len(got[1].Candidates) != 0 {
		t.Fatalf("original typo evidence must stay unresolved and untouched: %+v", got[1])
	}
	if corrections["cihuar"] != "Cihaur Geulis" {
		t.Fatalf("unexpected corrections: %v", corrections)
	}
}

func TestAppendFuzzyEvidenceDedupsSameCorrection(t *testing.T) {
	svc := &Service{postgresRepo: &fakePostgresRepo{entities: []model.Entity{kelurahanCihaurGeulis()}}}
	resolved := []model.ResolvedEvidence{
		{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "cihuar"}},
		{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "gelis"}},
	}
	got, corrections := svc.AppendFuzzyEvidence(context.Background(), 1, resolved)

	if len(got) != 3 {
		t.Fatalf("expected 3 resolved evidence (2 typos + 1 deduped correction), got %d", len(got))
	}
	added := got[2]
	if added.Value != "Cihaur Geulis" || len(added.Candidates) != 1 {
		t.Fatalf("expected deduped single entity, got %+v", added)
	}
	if len(corrections) != 2 || corrections["cihuar"] != "Cihaur Geulis" || corrections["gelis"] != "Cihaur Geulis" {
		t.Fatalf("unexpected corrections: %v", corrections)
	}
}

func TestAppendFuzzyEvidenceDedupsAgainstExistingEvidence(t *testing.T) {
	// fuzzy returns a name that is already evidence: merge candidates, add nothing
	svc := &Service{postgresRepo: &fakePostgresRepo{entities: []model.Entity{kelurahanCihaurGeulis()}}}
	resolved := []model.ResolvedEvidence{
		{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "Cihaur Geulis"}, Candidates: []model.Entity{kelurahanCihaurGeulis()}},
		{Evidence: model.Evidence{Type: model.EvidencePlaceName, Value: "cihuar"}},
	}
	got, corrections := svc.AppendFuzzyEvidence(context.Background(), 1, resolved)
	if len(got) != 2 {
		t.Fatalf("expected no new evidence, got %d", len(got))
	}
	if len(got[0].Candidates) != 1 {
		t.Fatalf("expected merged candidates to stay deduped, got %+v", got[0].Candidates)
	}
	if corrections["cihuar"] != "Cihaur Geulis" {
		t.Fatalf("unexpected corrections: %v", corrections)
	}
}

func TestEvidenceWithoutValues(t *testing.T) {
	evidence := []model.Evidence{
		{Type: model.EvidencePostalCode, Value: "40191"},
		{Type: model.EvidencePlaceName, Value: "bandung"},
		{Type: model.EvidencePlaceName, Value: "cihuar"},
		{Type: model.EvidencePlaceName, Value: "gelis"},
	}
	got := evidenceWithoutValues(evidence, map[string]string{"cihuar": "Cihaur Geulis", "gelis": "Cihaur Geulis"})
	if len(got) != 2 || got[0].Value != "40191" || got[1].Value != "bandung" {
		t.Fatalf("expected corrected tokens removed, got %+v", got)
	}
	if same := evidenceWithoutValues(evidence, nil); len(same) != len(evidence) {
		t.Fatalf("nil corrections must return evidence unchanged")
	}
}
