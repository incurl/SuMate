package engine

import (
	"strings"
	"testing"

	"github.com/incurl/sumate/internal/annotate"
	"github.com/incurl/sumate/internal/corpus"
	"github.com/incurl/sumate/internal/model"
)

func loadEngine(t *testing.T) *Engine {
	t.Helper()
	c, err := corpus.Load()
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	verses, err := corpus.NewPipeline(c, annotate.New()).Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return New(verses)
}

// TestMatchProducesTopAndAlternatives checks the shape of the final output.
func TestMatchProducesTopAndAlternatives(t *testing.T) {
	e := loadEngine(t)
	res := e.Match(Selection{
		Tags: map[model.RootKey][]string{
			model.RootMood: {"mood.low.longing"},
		},
	})
	if res.Primary.ID == "" {
		t.Fatal("no primary verse")
	}
	if len(res.Alternatives) < 2 || len(res.Alternatives) > 3 {
		t.Errorf("expected 2-3 alternatives, got %d", len(res.Alternatives))
	}
	if res.Primary.ID == res.Alternatives[0].ID {
		t.Error("the primary and the first alternative are the same verse")
	}
	if res.Explanation.Overall <= 0 {
		t.Error("no overall score")
	}
	if len(res.Explanation.Reasons) == 0 {
		t.Error("no explanation")
	}
}

// TestScoresAreOrdered checks that the ranking is coherent.
func TestScoresAreOrdered(t *testing.T) {
	e := loadEngine(t)
	cands := e.Rank(Selection{
		Tags: map[model.RootKey][]string{
			model.RootMood:     {"mood.insight.zen"},
			model.RootOccasion: {"occasion.exile.hardship"},
		},
	}, 0)
	if len(cands) < 10 {
		t.Fatalf("expected many candidates, got %d", len(cands))
	}
	for i := 1; i < len(cands); i++ {
		if cands[i].Score > cands[i-1].Score+1e-9 {
			t.Fatalf("candidate %d scored %.4f above candidate %d (%.4f)",
				i, cands[i].Score, i-1, cands[i-1].Score)
		}
	}
}

// TestMoodSelectionChangesTheAnswer checks that the mood filter has teeth.
func TestMoodSelectionChangesTheAnswer(t *testing.T) {
	e := loadEngine(t)

	joy := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.joy.casual"},
	}})
	sad := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low.longing"},
	}})
	if joy.Primary.ID == sad.Primary.ID {
		t.Errorf("both selections produced %q", joy.Primary.Text)
	}
	t.Logf("joy -> %s", joy.Primary.Text)
	t.Logf("sad -> %s", sad.Primary.Text)
}

// TestDongpoTemperamentIsPreferred checks the product's healing bias: a
// verse carrying Dongpo's transcendent attitude must outrank a flat one.
func TestDongpoTemperamentIsPreferred(t *testing.T) {
	e := loadEngine(t)
	cands := e.Rank(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low.helpless"},
	}}, 0)

	var withTangent, withoutTangent *Candidate
	for i := range cands {
		c := &cands[i]
		if len(c.Verse.Annotation.Tangents) >= 3 && withTangent == nil {
			withTangent = c
		}
		if len(c.Verse.Annotation.Tangents) == 0 && withoutTangent == nil {
			withoutTangent = c
		}
	}
	if withTangent == nil || withoutTangent == nil {
		t.Skip("not enough contrasting candidates in this corpus")
	}
	if withTangent.Score < withoutTangent.Score {
		t.Errorf("a verse with Dongpo's temperaments (%.3f) scored below a flat one (%.3f)",
			withTangent.Score, withoutTangent.Score)
	}
}

// TestExplanationMentionsTheVerse checks that the explanation is about the
// verse that was chosen.
func TestExplanationMentionsTheVerse(t *testing.T) {
	e := loadEngine(t)
	res := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.insight.plain"},
	}})
	joined := strings.Join(res.Explanation.Reasons, " ")
	if joined == "" {
		t.Fatal("empty explanation")
	}
	if !strings.Contains(joined, res.Primary.Annotation.Philosophy[:4]) &&
		!strings.Contains(joined, res.Primary.Annotation.Attitude[:4]) {
		t.Errorf("the explanation does not mention the verse's own reading: %q", joined)
	}
}
