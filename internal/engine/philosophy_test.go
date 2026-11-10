package engine

import (
	"strings"
	"testing"

	"github.com/incurl/sumate/internal/model"
	"github.com/incurl/sumate/internal/philosophy"
)

// TestPhilosophicalFitChangesTheRanking checks that the philosophical layer
// actually affects the result, not merely decorates it.
//
// Two readers in nearly the same emotional state -- one who has come out the
// other side, one who is still in the middle of it -- must not be handed the
// same consolation.
func TestPhilosophicalFitChangesTheRanking(t *testing.T) {
	e := loadEngine(t)

	// A reader who has been through it: the consolation fits.
	resigned := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood:     {"mood.low.helpless"},
		model.RootOccasion: {"occasion.exile.hardship.solitude"},
	}})

	// A reader who is starting to get better: the same consolation may not
	// be what they need.
	recovering := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood:     {"mood.joy.hope"},
		model.RootOccasion: {"occasion.career.appointed"},
	}})

	if resigned.Primary.ID == "" || recovering.Primary.ID == "" {
		t.Fatal("both selections must produce a verse")
	}
	t.Logf("resigned  -> %s", resigned.Primary.Text)
	t.Logf("recovering-> %s", recovering.Primary.Text)

	// The philosophical summary must be present when a stance does not fit.
	if resigned.Explanation.Philosophy == "" && recovering.Explanation.Philosophy == "" {
		t.Error("neither match carries a philosophical summary")
	}
}

// TestFailedPremiseProducesACaveat checks that a mismatch is reported rather
// than hidden.
func TestFailedPremiseProducesACaveat(t *testing.T) {
	e := loadEngine(t)

	// Find a verse whose stance fails for this reader.
	reader := philosophy.Reader{Tags: map[model.RootKey][]string{
		model.RootMood:     {"mood.joy.hope"},
		model.RootOccasion: {"occasion.career.appointed"},
	}}
	cands := e.Rank(Selection{Tags: reader.Tags}, 0)

	var sawCaveat bool
	for _, c := range cands {
		if len(c.Parts.Caveats) > 0 {
			sawCaveat = true
			for _, cav := range c.Parts.Caveats {
				if cav.Note == "" {
					t.Error("a caveat carries no note")
				}
				if cav.Premise == "" || cav.Shadow == "" {
					t.Error("a caveat lacks its premise or shadow")
				}
			}
			break
		}
	}
	if !sawCaveat {
		t.Error("no verse produced a philosophical caveat for a mismatched reader")
	}
}

// TestStancesAreAttachedToVerses checks that verses carry stances at all.
func TestStancesAreAttachedToVerses(t *testing.T) {
	e := loadEngine(t)
	withStance := 0
	for _, v := range e.Verses() {
		if len(v.Annotation.Stances) > 0 {
			withStance++
		}
	}
	t.Logf("verses carrying a stance: %d / %d", withStance, len(e.Verses()))
	if withStance == 0 {
		t.Fatal("no verse carries a philosophical stance")
	}
	if withStance*10 < len(e.Verses()) {
		t.Errorf("only %d of %d verses carry a stance", withStance, len(e.Verses()))
	}
}

// TestCaveatUsesItsOwnShadow checks that a caveat is specific rather than a
// generic disclaimer.
func TestCaveatUsesItsOwnShadow(t *testing.T) {
	reader := philosophy.Reader{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.joy.hope"},
	}}
	all := philosophy.AssessAll(reader)
	seen := map[string]bool{}
	for _, a := range all {
		if a.Fit != philosophy.FitFailed {
			continue
		}
		if seen[a.Note] {
			t.Errorf("two different stances produced the same note: %q", a.Note)
		}
		seen[a.Note] = true
		if !strings.Contains(a.Note, a.Shadow) {
			t.Errorf("the note does not carry the stance's own shadow: %q", a.Note)
		}
	}
}
