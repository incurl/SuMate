package philosophy

import (
	"testing"

	"github.com/incurl/sumate/internal/curated"
	"github.com/incurl/sumate/internal/model"
)

func TestStanceLexicon(t *testing.T) {
	stances := curated.Stances()
	if len(stances) < 10 {
		t.Fatalf("expected a substantial stance lexicon, got %d", len(stances))
	}
	for k, s := range stances {
		if s.Key != k {
			t.Errorf("stance %s has mismatched key %s", k, s.Key)
		}
		if s.Claim == "" || s.Premise == "" || s.Shadow == "" || s.Earned == "" {
			t.Errorf("stance %s has an empty required field", k)
		}
		if len(s.Supports) == 0 {
			t.Errorf("stance %s supports nothing", k)
		}
		if len(s.Undermines) == 0 {
			t.Errorf("stance %s has no shadow condition", k)
		}
		if len(s.Concepts) == 0 {
			t.Errorf("stance %s is carried by no concept", k)
		}
	}
}

func TestStanceOrderCoversLexicon(t *testing.T) {
	stances := curated.Stances()
	if len(curated.StanceOrder) != len(stances) {
		t.Errorf("StanceOrder has %d entries but the lexicon has %d",
			len(curated.StanceOrder), len(stances))
	}
	for _, k := range curated.StanceOrder {
		if _, ok := stances[k]; !ok {
			t.Errorf("StanceOrder names an unknown stance %q", k)
		}
	}
}

func TestPremiseFailureIsDetected(t *testing.T) {
	stances := curated.Stances()

	// A reader who is still able to act should NOT receive the
	// "there is nothing more to do" stance as unqualified comfort.
	s := stances["fate.limits"]
	r := Reader{Tags: map[model.RootKey][]string{
		model.RootMood:     {"mood.joy.hope"},
		model.RootOccasion: {"occasion.career.appointed"},
	}}
	a := Assess(s, r)
	if a.Fit == FitHeld {
		t.Errorf("the 'fate.limits' stance was judged as held for an active reader")
	}
	if a.Note == "" {
		t.Error("no note was produced")
	}
}

func TestPremiseHoldingIsDetected(t *testing.T) {
	stances := curated.Stances()
	s := stances["fate.limits"]
	r := Reader{Tags: map[model.RootKey][]string{
		model.RootMood:     {"mood.low.helpless"},
		model.RootOccasion: {"occasion.exile.hardship.solitude"},
	}}
	a := Assess(s, r)
	if a.Fit != FitHeld {
		t.Errorf("the 'fate.limits' stance was judged %q for a reader it fits", a.Fit)
	}
}

func TestAssessAllIsOrdered(t *testing.T) {
	r := Reader{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low.helpless"},
	}}
	all := AssessAll(r)
	if len(all) == 0 {
		t.Fatal("no stances assessed")
	}
	for i := 1; i < len(all); i++ {
		if fitRank(all[i].Fit) < fitRank(all[i-1].Fit) {
			t.Errorf("assessment %d (%s) is ordered before %d (%s)",
				i, all[i].Fit, i-1, all[i-1].Fit)
		}
	}
}
