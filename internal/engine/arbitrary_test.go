package engine

import (
	"fmt"
	"testing"

	"github.com/incurl/sumate/internal/model"
)

// TestAnswerDoesNotDependOnCorpusOrder is the regression test for the
// deepest defect found in the metaphysical evaluation.
//
// The product promises "the single most resonant verse", and the
// superlative presupposes a total order.  But the scoring axes used to
// saturate in the decision region: the semantic axis took one value and the
// scenario axis took one value, so the top thirty verses tied at 82 and the
// winner was decided by sort stability -- that is, by the position of the
// verse in the corpus file.  Reversing the corpus changed the answer.
//
// The answer must now be a property of the verses, not of their order.
func TestAnswerDoesNotDependOnCorpusOrder(t *testing.T) {
	e := loadEngine(t)

	// A reversed copy of the same verses.
	rev := make([]model.Verse, 0, len(e.Verses()))
	for i := len(e.Verses()) - 1; i >= 0; i-- {
		rev = append(rev, e.Verses()[i])
	}
	revEngine := New(rev)

	for _, tags := range []map[model.RootKey][]string{
		{model.RootMood: {"mood.low"}},
		{model.RootMood: {"mood.low.longing"}},
		{model.RootMood: {"mood.joy.casual"}},
		{model.RootOccasion: {"occasion.exile.hardship"}},
	} {
		a := e.Match(Selection{Tags: tags})
		b := revEngine.Match(Selection{Tags: tags})
		if a.Primary.ID == "" {
			t.Fatal("no primary verse")
		}
		if a.Primary.ID != b.Primary.ID {
			t.Errorf("the answer depends on corpus order for %v:\n  forward:  %s\n  reversed: %s",
				tags, a.Primary.Text, b.Primary.Text)
		}
	}
}

// TestTheTopIsNotATie checks that the winner is actually decided: if many
// verses tie at the top, the choice among them is arbitrary again.
func TestTheTopIsNotATie(t *testing.T) {
	e := loadEngine(t)
	cands := e.Rank(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low"},
	}}, 30)
	if len(cands) == 0 {
		t.Fatal("no candidates")
	}
	top := cands[0].Score
	tied := 0
	for _, c := range cands {
		if c.Score == top {
			tied++
		}
	}
	fmt.Printf("  前 30 名中与第 1 名同分者：%d 句\n", tied)
	if tied > 3 {
		t.Errorf("%d of the top 30 tie at the top score; the winner is arbitrary", tied)
	}
}

// TestSemanticAxisVaries checks that the semantic axis measures something.
//
// It used to be a constant over the decision region: standard deviation
// 0.00, one distinct value.  A third of the total weight was producing no
// difference at all.
func TestSemanticAxisVaries(t *testing.T) {
	e := loadEngine(t)
	cands := e.Rank(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low.longing"},
	}}, 50)
	if len(cands) < 10 {
		t.Fatalf("not enough candidates: %d", len(cands))
	}
	distinct := map[int]bool{}
	for _, c := range cands {
		distinct[c.Parts.Semantic] = true
	}
	fmt.Printf("  前 50 名语义分取值：%d 种\n", len(distinct))
	if len(distinct) < 3 {
		t.Errorf("the semantic axis takes %d distinct values in the decision region", len(distinct))
	}
}
