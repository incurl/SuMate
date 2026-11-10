package service

import (
	"strings"
	"testing"

	"github.com/sumate/sumate/internal/tagtree"
)

// TestFlowMirrorsFrontend mirrors the state transitions of frontend/src/state.ts
// and checks that the backend produces the counts the UI expects.
func TestFlowMirrorsFrontend(t *testing.T) {
	s := newService(t)

	roots := s.Roots()
	moodRoot := roots[0]
	placeRoot := roots[1]

	// 1) pick the mood dimension -> show its children
	moodKids := s.Children(moodRoot.ID, nil)
	if len(moodKids) == 0 {
		t.Fatal("no mood children")
	}
	low := moodKids[0] // 低回与疗愈

	// 2) pick a sub-option of the mood dimension
	path := []tagtree.Filter{
		{NodeID: moodRoot.ID, Root: moodRoot.Root},
		{NodeID: low.ID, Root: low.Root},
	}
	moodLevel2 := s.Children(low.ID, path)
	if len(moodLevel2) == 0 {
		t.Fatal("no second-level mood children")
	}
	t.Logf("mood level 2: %d options, e.g. %s (%d)",
		len(moodLevel2), moodLevel2[0].Label, moodLevel2[0].Remaining)

	// 3) switch to the place dimension: the mood selection must survive
	placeKids := s.Children(placeRoot.ID, path)
	if len(placeKids) == 0 {
		t.Fatal("no place children under the mood selection")
	}
	for _, k := range placeKids {
		if k.Remaining > low.Remaining {
			t.Errorf("place option %s has more candidates (%d) than the mood branch (%d)",
				k.Label, k.Remaining, low.Remaining)
		}
	}
	t.Logf("place under the mood selection: %d options, e.g. %s (%d)",
		len(placeKids), placeKids[0].Label, placeKids[0].Remaining)

	// 4) the counts along the whole path must be monotonically decreasing
	path = append(path, tagtree.Filter{NodeID: placeRoot.ID, Root: placeRoot.Root})
	crumbs := s.Breadcrumb(path)
	for i := 1; i < len(crumbs); i++ {
		if crumbs[i].Remaining > crumbs[i-1].Remaining {
			t.Errorf("step %d (%s) has %d candidates, more than step %d (%d)",
				i, crumbs[i].Label, crumbs[i].Remaining, i-1, crumbs[i-1].Remaining)
		}
	}
	for _, c := range crumbs {
		t.Logf("  %-40s %5d", strings.Join(c.Path, " > "), c.Remaining)
	}

	// 5) backtracking to the first step restores the wider count
	back := s.Breadcrumb(path[:1])
	if back[0].Remaining != crumbs[0].Remaining {
		t.Errorf("backtrack count %d != step 1 count %d",
			back[0].Remaining, crumbs[0].Remaining)
	}

	// 6) the final match respects the whole selection
	res := s.Match(path, "")
	if res.Primary.ID == "" {
		t.Fatal("no match")
	}
	if !carriesTag(res.Primary, moodRoot.ID) {
		t.Errorf("the primary match %q does not carry the selected mood", res.Primary.Text)
	}
	t.Logf("primary: %s (%d)", res.Primary.Text, res.Explanation.Overall)
}
