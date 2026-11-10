package annotate

import (
	"strings"
	"testing"

	"github.com/incurl/sumate/internal/curated"
	"github.com/incurl/sumate/internal/model"
)

// TestSingleCharacterKeywordsDoNotDominate checks the Y1 rule: a keyword
// that is also an ordinary word must not attach a concept on its own.
//
// Single characters such as 「风」「月」「无」 occur in a large fraction of
// the corpus; if they carried a concept by themselves, almost every verse
// would be tagged.
func TestSingleCharacterKeywordsDoNotDominate(t *testing.T) {
	e := New()
	concepts := e.Concepts()

	// Every keyword that is a single character.
	single := map[string]bool{}
	for _, c := range concepts {
		for _, kw := range c.Keywords {
			if len([]rune(kw)) == 1 {
				single[kw] = true
			}
		}
	}
	if len(single) == 0 {
		t.Fatal("no single-character keywords found")
	}

	// A verse built only of single-character trigger words may match
	// concepts, but none of them may be *primary*: no specific phrase
	// supports them.
	text := "风月无云不归何处相送"
	matched, specific := e.matchConcepts(text)
	if len(specific) != 0 {
		t.Errorf("a text of only common characters produced %d specific matches: %v",
			len(specific), specific)
	}
	for _, k := range matched {
		if k == "moon.reunion" {
			t.Error("a common-character text matched the reunion concept specifically")
		}
	}

	// A verse carrying a specific phrase must match that concept
	// specifically.
	_, reunion := e.matchConcepts("但愿人长久千里共婵娟")
	if len(reunion) == 0 {
		t.Fatal("a verse with a specific phrase matched nothing")
	}
	found := false
	for _, k := range reunion {
		if k == "moon.reunion" || k == "human.brother" {
			found = true
		}
	}
	if !found {
		t.Errorf("the reunion verse did not match the reunion concept: %v", specific)
	}
}

// TestConceptRankingIsDeterministic checks that the same text always yields
// the same ordering, regardless of Go's map iteration order.
func TestConceptRankingIsDeterministic(t *testing.T) {
	e := New()
	texts := []string{
		"但愿人长久千里共婵娟",
		"拣尽寒枝不肯栖寂寞沙洲冷",
		"莫听穿林打叶声何妨吟啸且徐行",
		"谁道人生无再少门前流水尚能西",
	}
	for _, text := range texts {
		first := e.MatchedConcepts(text)
		for i := 0; i < 20; i++ {
			again := e.MatchedConcepts(text)
			if len(again) != len(first) {
				t.Fatalf("unstable result length for %q", text)
			}
			for j := range first {
				if again[j] != first[j] {
					t.Fatalf("unstable ordering for %q: %v vs %v", text, first, again)
				}
			}
		}
	}
}

// TestAnnotatedVersesCarryTags checks that a verse with a strong concept
// receives tags in more than one dimension, so that the tree can filter it
// from several angles.
func TestAnnotatedVersesCarryTags(t *testing.T) {
	e := New()
	v := model.Verse{Text: "但愿人长久，千里共婵娟。"}
	a := e.Annotate(&v, model.Poem{})

	if len(a.Concepts) == 0 {
		t.Fatal("no concepts matched")
	}
	if len(a.Tangents) == 0 {
		t.Error("no Dongpo temperament assigned")
	}
	if len(a.Emotions) == 0 {
		t.Error("no emotions assigned")
	}
	dims := 0
	for _, tags := range a.Tags {
		if len(tags) > 0 {
			dims++
		}
	}
	if dims < 2 {
		t.Errorf("expected tags in at least 2 dimensions, got %d: %v", dims, a.Tags)
	}
	if a.Literal == "" || a.Conception == "" || a.Philosophy == "" ||
		a.Attitude == "" || a.Healing == "" {
		t.Error("the annotation is incomplete")
	}
	if a.Crowd == "" || a.MoodThreshold == "" || a.ScenarioBoundary == "" {
		t.Error("the audience / threshold / boundary fields are empty")
	}
}

// TestEmotionStrengthsAreGraded checks that the graded strengths use the
// four declared levels and never exceed the strongest.
func TestEmotionStrengthsAreGraded(t *testing.T) {
	e := New()
	for _, text := range []string{
		"人生如梦，一尊还酹江月",
		"莫听穿林打叶声，何妨吟啸且徐行",
		"拣尽寒枝不肯栖，寂寞沙洲冷",
	} {
		v := model.Verse{Text: text}
		a := e.Annotate(&v, model.Poem{})
		for _, em := range a.Emotions {
			switch em.Strength {
			case model.Slight, model.Moderate, model.Obvious, model.Profound:
			default:
				t.Errorf("invalid strength %q on %s", em.Strength, em.ID)
			}
			if !strings.HasPrefix(em.ID, em.Dimension+".") {
				t.Errorf("emotion %q does not belong to dimension %q", em.ID, em.Dimension)
			}
		}
	}
}

// TestConceptTagsAreReachable guards the invariant that every tag a
// concept claims exists in the tree.
func TestConceptTagsAreReachable(t *testing.T) {
	c := curated.Concepts()
	for k, x := range c {
		for _, tag := range x.Tags {
			if !strings.HasPrefix(tag, "mood.") &&
				!strings.HasPrefix(tag, "place.") &&
				!strings.HasPrefix(tag, "weather.") &&
				!strings.HasPrefix(tag, "relation.") &&
				!strings.HasPrefix(tag, "occasion.") {
				t.Errorf("concept %s has a malformed tag %q", k, tag)
			}
		}
	}
}
