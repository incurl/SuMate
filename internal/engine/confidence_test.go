package engine

import (
	"testing"

	"github.com/incurl/sumate/internal/model"
)

// TestConfidenceIsReported checks that the product always states how sure it
// is, and says so honestly when it is unsure.
func TestConfidenceIsReported(t *testing.T) {
	e := loadEngine(t)

	// A precise selection.
	precise := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood:     {"mood.low.helpless"},
		model.RootOccasion: {"occasion.exile.hardship.solitude"},
		model.RootWeather:  {"weather.rain.storm"},
	}})
	t.Logf("precise: confidence=%d confident=%v", precise.Confidence, precise.Confident)
	t.Logf("  verse: %s", precise.Primary.Text)
	if precise.Confidence < 0 || precise.Confidence > 100 {
		t.Errorf("confidence out of range: %d", precise.Confidence)
	}

	// A vague selection: only a root, no detail.
	vague := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood"},
	}})
	t.Logf("vague:   confidence=%d confident=%v", vague.Confidence, vague.Confident)
	if vague.Confidence < 0 || vague.Confidence > 100 {
		t.Errorf("confidence out of range: %d", vague.Confidence)
	}

	// An empty selection should be the least confident of all.
	empty := e.Match(Selection{Tags: map[model.RootKey][]string{}})
	t.Logf("empty:   confidence=%d confident=%v", empty.Confidence, empty.Confident)
	if empty.Confidence > vague.Confidence {
		t.Errorf("an empty selection scored higher (%d) than a vague one (%d)",
			empty.Confidence, vague.Confidence)
	}
}

// TestHonestNoteAppearsWhenUnsure checks that an unsure answer explains
// itself instead of pretending.
func TestHonestNoteAppearsWhenUnsure(t *testing.T) {
	e := loadEngine(t)
	res := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood"},
	}})
	if !res.Confident {
		if res.Honest == "" {
			t.Error("an unconfident answer carries no honest note")
		}
		t.Logf("honest note: %s", res.Honest)
	}
}

// TestAMoreDetailedSelectionIsMoreConfident checks that the product gets
// surer as the user says more.
func TestAMoreDetailedSelectionIsMoreConfident(t *testing.T) {
	e := loadEngine(t)

	broad := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low"},
	}})
	narrow := e.Match(Selection{Tags: map[model.RootKey][]string{
		model.RootMood:     {"mood.low.helpless"},
		model.RootOccasion: {"occasion.exile.hardship.solitude"},
	}})
	t.Logf("broad:  %d (confident=%v)", broad.Confidence, broad.Confident)
	t.Logf("narrow: %d (confident=%v)", narrow.Confidence, narrow.Confident)

	if narrow.Confidence < broad.Confidence {
		t.Errorf("a more detailed selection was less confident (%d) than a broad one (%d)",
			narrow.Confidence, broad.Confidence)
	}
}
