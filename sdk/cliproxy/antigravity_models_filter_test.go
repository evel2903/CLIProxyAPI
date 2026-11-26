package cliproxy

import "testing"

func TestFilterAntigravityModelsNilCatalogPreservesModels(t *testing.T) {
	models := []*ModelInfo{
		{ID: "gemini-3-flash"},
		{ID: "gemini-3.1-flash-lite"},
	}

	filtered := filterAntigravityModels(models, antigravityModelCapabilityHints{})
	if len(filtered) != len(models) {
		t.Fatalf("nil catalog filtered %d of %d models", len(models)-len(filtered), len(models))
	}

	for i := range models {
		if filtered[i] != models[i] {
			t.Fatalf("model %d changed when catalog was unknown", i)
		}
	}
}

// Fork: the static catalog is never pruned, even by an empty known catalog.
func TestFilterAntigravityModelsEmptyCatalogKeepsStaticModels(t *testing.T) {
	models := []*ModelInfo{{ID: "gemini-3-flash"}}

	filtered := filterAntigravityModels(models, antigravityModelCapabilityHints{
		ModelIDs: map[string]struct{}{},
	})
	if len(filtered) != len(models) {
		t.Fatalf("empty known catalog pruned %d of %d static models", len(models)-len(filtered), len(models))
	}
}
