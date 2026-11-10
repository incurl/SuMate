package curated

import (
	"testing"

	"github.com/incurl/sumate/internal/tagtree"
)

func TestConceptsAreWellFormed(t *testing.T) {
	c := Concepts()
	if len(c) < 25 {
		t.Fatalf("expected a substantial concept lexicon, got %d", len(c))
	}
	for k, x := range c {
		if x.Key != k {
			t.Errorf("concept %s has mismatched key %s", k, x.Key)
		}
		if x.Name == "" || x.Surface == "" || x.Deep == "" ||
			x.Philosophy == "" || x.Attitude == "" {
			t.Errorf("concept %s has an empty required field", k)
		}
		if x.Polarity != "positive" && x.Polarity != "neutral" &&
			x.Polarity != "melanchol" && x.Polarity != "mixed" {
			t.Errorf("concept %s has invalid polarity %q", k, x.Polarity)
		}
		if len(x.Keywords) == 0 {
			t.Errorf("concept %s has no keywords", k)
		}
		if len(x.Tags) == 0 {
			t.Errorf("concept %s has no tags", k)
		}
	}
}

// Keywords may be shared between concepts on purpose -- a verse like
// "休将白发唱黄鸡" legitimately carries both the ageing and the
// daily-life concept.  What must not happen is a concept listing the same
// keyword twice, or an empty keyword.
func TestConceptsHaveNoDuplicateKeywords(t *testing.T) {
	c := Concepts()
	for k, x := range c {
		seen := map[string]bool{}
		for _, kw := range x.Keywords {
			if kw == "" {
				t.Errorf("concept %s has an empty keyword", k)
				continue
			}
			if seen[kw] {
				t.Errorf("concept %s repeats keyword %q", k, kw)
			}
			seen[kw] = true
		}
	}
}

// Every tag referenced by a concept must exist in the tag tree.
func TestConceptTagsExistInTree(t *testing.T) {
	c := Concepts()
	tree := tagtree.Build()
	for k, x := range c {
		for _, tag := range x.Tags {
			if _, ok := tree.Node(tag); !ok {
				t.Errorf("concept %s references unknown tag %q", k, tag)
			}
		}
	}
}

// Every emotion and tangent must be a known key.
func TestConceptEmotionsAndTangentsAreKnown(t *testing.T) {
	c := Concepts()
	known := curatedEmotionKeys()
	for k, x := range c {
		for _, e := range x.Emotions {
			if !known[e] {
				t.Errorf("concept %s references unknown emotion %q", k, e)
			}
		}
		for _, tg := range x.Tangents {
			if !knownTangent(tg) {
				t.Errorf("concept %s references unknown tangent %q", k, tg)
			}
		}
	}
}
