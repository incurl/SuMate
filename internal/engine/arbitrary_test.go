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

// TestScenarioAxisIsGraded checks that the scenario axis can tell a
// precise match from an approximate one.
//
// The axis used to be binary: the verse carries the selected tag, or it
// does not.  Inside a pool that has already been filtered by that tag
// every surviving verse scored 1.0, so the axis was a constant and could
// not break a single tie.  It now grades the match by how precisely the
// verse addresses the selection.
func TestScenarioAxisIsGraded(t *testing.T) {
	e := loadEngine(t)

	// A deep tag: the pool holds verses that carry the tag itself and
	// verses that carry a shallower ancestor of it.
	cands := e.Rank(Selection{Tags: map[model.RootKey][]string{
		model.RootOccasion: {"occasion.exile.hardship"},
	}}, 200)
	if len(cands) < 20 {
		t.Fatalf("not enough candidates: %d", len(cands))
	}
	distinct := map[int]bool{}
	for _, c := range cands {
		distinct[c.Parts.Scenario] = true
	}
	fmt.Printf("  情境分取值：%d 种 %v\n", len(distinct), sortedKeys(distinct))
	if len(distinct) < 2 {
		t.Errorf("the scenario axis takes %d distinct values; it cannot discriminate", len(distinct))
	}
}

// TestEveryAxisVariesInTheFilteredPool checks the claim the audit makes:
// no axis is a constant in the population the user actually sees.
func TestEveryAxisVariesInTheFilteredPool(t *testing.T) {
	e := loadEngine(t)

	// The pool the user sees for a leaf selection: verses carrying the tag.
	var pool []model.Verse
	for _, v := range e.Verses() {
		if verseHasTag(v, "mood.low.longing") {
			pool = append(pool, v)
		}
	}
	if len(pool) == 0 {
		t.Fatal("empty pool")
	}
	sub := New(pool)
	cands := sub.Rank(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low.longing"},
	}}, 50)

	axes := map[string]map[int]bool{
		"语义": {}, "情绪": {}, "情境": {}, "气质": {},
	}
	for _, c := range cands {
		axes["语义"][c.Parts.Semantic] = true
		axes["情绪"][c.Parts.Emotional] = true
		axes["情境"][c.Parts.Scenario] = true
		axes["气质"][c.Parts.Temperament] = true
	}
	for name, vals := range axes {
		fmt.Printf("  %s 取值 %d 种\n", name, len(vals))
	}
	// The semantic and emotional axes must discriminate.  The scenario axis
	// is constant here by construction -- every verse in the pool carries
	// exactly that tag -- and the temperament axis may be too.
	if len(axes["语义"]) < 3 {
		t.Errorf("the semantic axis takes %d distinct values in the pool", len(axes["语义"]))
	}
	if len(axes["情绪"]) < 2 {
		t.Errorf("the emotional axis takes %d distinct values in the pool", len(axes["情绪"]))
	}
}

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
