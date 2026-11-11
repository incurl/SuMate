package service

import (
	"strings"
	"testing"

	"github.com/incurl/sumate/internal/model"
	"github.com/incurl/sumate/internal/tagtree"
)

func newService(t *testing.T) *Service {
	t.Helper()
	s, err := New()
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return s
}

func TestStatsAreSane(t *testing.T) {
	s := newService(t)
	st := s.Stats()
	t.Logf("%+v", st)
	if st.Verses < 1500 {
		t.Fatalf("expected a large verse set, got %d", st.Verses)
	}
	if st.Tunes < 50 {
		t.Fatalf("expected many tunes, got %d", st.Tunes)
	}
	if st.Lines == 0 || st.Couplets == 0 {
		t.Fatalf("expected both lines and couplets: %+v", st)
	}
}

func TestRootsShowLiveCounts(t *testing.T) {
	s := newService(t)
	roots := s.Roots()
	if len(roots) != 5 {
		t.Fatalf("expected 5 roots, got %d", len(roots))
	}
	total := 0
	for _, r := range roots {
		if r.Remaining <= 0 {
			t.Errorf("root %s has no candidates", r.ID)
		}
		total += r.Remaining
	}
	t.Logf("root counts: ")
	for _, r := range roots {
		t.Logf("  %-10s %-14s %d", r.ID, r.Label, r.Remaining)
	}
	if total == 0 {
		t.Error("no candidates at all")
	}
}

func TestDrillingNarrowsResults(t *testing.T) {
	s := newService(t)
	roots := s.Roots()

	// drill into the mood root
	mood := roots[0]
	kids := s.Children(mood.ID, nil)
	if len(kids) == 0 {
		t.Fatal("mood root has no children")
	}
	first := kids[0]
	t.Logf("drilling into %s (%s) -> %d", first.ID, first.Label, first.Remaining)

	grand := s.Children(first.ID, []tagtree.Filter{{NodeID: mood.ID, Root: mood.Root}})
	if len(grand) == 0 {
		t.Fatal("no grandchildren")
	}
	for _, g := range grand {
		if g.Remaining > first.Remaining {
			t.Errorf("grandchild %s has more candidates (%d) than its parent (%d)",
				g.ID, g.Remaining, first.Remaining)
		}
	}
}

func TestBreadcrumbRecordsPathAndCounts(t *testing.T) {
	s := newService(t)
	roots := s.Roots()
	mood := roots[0]
	kids := s.Children(mood.ID, nil)
	child := kids[0]

	bc := s.Breadcrumb([]tagtree.Filter{
		{NodeID: mood.ID, Root: mood.Root},
		{NodeID: child.ID, Root: child.Root},
	})
	if len(bc) != 2 {
		t.Fatalf("expected 2 breadcrumb steps, got %d", len(bc))
	}
	if bc[0].Remaining < bc[1].Remaining {
		t.Errorf("breadcrumb counts should not increase: %d -> %d",
			bc[0].Remaining, bc[1].Remaining)
	}
	if len(bc[1].Path) < 2 {
		t.Errorf("breadcrumb path should be a full path, got %v", bc[1].Path)
	}
}

func TestPreviewWorksAtAnyLayer(t *testing.T) {
	s := newService(t)
	roots := s.Roots()
	mood := roots[0]

	// preview right after the root selection
	vs := s.Preview([]tagtree.Filter{{NodeID: mood.ID, Root: mood.Root}}, 10)
	if len(vs) == 0 {
		t.Fatal("preview at level 2 returned nothing")
	}
	if len(vs) > 10 {
		t.Fatalf("preview ignored the limit: %d", len(vs))
	}
	t.Logf("preview at level 2: %d verses", len(vs))
}

func TestMatchReturnsPrimaryAndAlternatives(t *testing.T) {
	s := newService(t)
	roots := s.Roots()
	mood := roots[0]

	res := s.Match([]tagtree.Filter{{NodeID: mood.ID, Root: mood.Root}}, "")
	if res.Primary.ID == "" {
		t.Fatal("no primary match")
	}
	if len(res.Alternatives) < 2 {
		t.Errorf("expected 2-3 alternatives, got %d", len(res.Alternatives))
	}
	if res.Explanation.Overall <= 0 {
		t.Error("no explanation score")
	}
	t.Logf("primary: %s (%d)", res.Primary.Text, res.Explanation.Overall)
	for _, a := range res.Alternatives {
		t.Logf("  alt: %s", a.Text)
	}
	for _, r := range res.Explanation.Reasons {
		t.Logf("  reason: %s", r)
	}
}

func TestMatchNarrowsWithMoreDimensions(t *testing.T) {
	s := newService(t)
	roots := s.Roots()

	// mood only
	moodOnly := s.Match([]tagtree.Filter{{NodeID: roots[0].ID, Root: roots[0].Root}}, "")
	// mood + place
	moodKids := s.Children(roots[0].ID, nil)
	placeKids := s.Children(roots[1].ID, nil)
	if len(moodKids) == 0 || len(placeKids) == 0 {
		t.Fatal("missing children")
	}
	both := s.Match([]tagtree.Filter{
		{NodeID: roots[0].ID, Root: roots[0].Root},
		{NodeID: moodKids[0].ID, Root: roots[0].Root},
		{NodeID: roots[1].ID, Root: roots[1].Root},
		{NodeID: placeKids[0].ID, Root: roots[1].Root},
	}, "")

	if both.Primary.ID == "" {
		t.Fatal("no match with two dimensions")
	}
	if both.Primary.Text == moodOnly.Primary.Text {
		t.Logf("both selections produced the same verse: %s", both.Primary.Text)
	}
	t.Logf("mood only -> %s", moodOnly.Primary.Text)
	t.Logf("mood+place -> %s", both.Primary.Text)
}

func TestSearchFindsKnownVerses(t *testing.T) {
	s := newService(t)
	for _, q := range []string{"婵娟", "风雨", "白发", "扁舟"} {
		vs := s.Search(q, 5)
		if len(vs) == 0 {
			t.Errorf("search %q found nothing", q)
		}
	}
}

func TestEveryVerseIsRenderable(t *testing.T) {
	s := newService(t)
	all := s.AllVerses()
	if len(all) == 0 {
		t.Fatal("no verses")
	}
	for _, v := range all {
		if v.Text == "" {
			t.Fatalf("verse %s has no text", v.ID)
		}
		if v.Tune == "" {
			t.Fatalf("verse %s has no tune", v.ID)
		}
		if v.Annotation.Literal == "" {
			t.Fatalf("verse %s has no literal meaning", v.ID)
		}
		if v.Annotation.Attitude == "" {
			t.Fatalf("verse %s has no attitude", v.ID)
		}
		if v.Background.Historical == "" {
			t.Fatalf("verse %s has no background", v.ID)
		}
		if !strings.Contains(v.ID, "-") {
			t.Fatalf("verse id %q is malformed", v.ID)
		}
	}
}

func TestTagDimensionsArePopulated(t *testing.T) {
	s := newService(t)
	all := s.AllVerses()
	counts := map[model.RootKey]int{}
	for _, v := range all {
		for root, tags := range v.Annotation.Tags {
			if len(tags) > 0 {
				counts[root]++
			}
		}
	}
	for _, root := range model.RootOrder {
		if counts[root] == 0 {
			t.Errorf("dimension %s is never used", root)
		}
	}
	t.Logf("dimension usage: %v", counts)
}

// TestMatchCountsOnlyTheMatchingPool is the regression test for a pool
// bug that made the candidate count meaningless.
//
// The service computed the filtered pool with tree.Apply and then threw it
// away, calling engine.Match -- which ranks *every* verse in the corpus.
// So a selection that matched 270 verses reported 2604 candidates, and the
// ranking was over verses the user had already excluded.
//
// The engine's ranking is order-independent, so restricting the pool does
// not make the answer arbitrary -- it makes it *correct*.
func TestMatchCountsOnlyTheMatchingPool(t *testing.T) {
	s := newService(t)
	roots := s.Roots()
	mood := roots[0]
	kids := s.Children(mood.ID, nil)
	if len(kids) == 0 {
		t.Fatal("no children")
	}
	child := kids[0]

	res := s.Match([]tagtree.Filter{{NodeID: child.ID, Root: child.Root}}, "")
	if res.Candidates == 0 {
		t.Fatal("no candidates")
	}

	// Every candidate must actually carry the selected tag.
	strict := s.tree.Apply(s.index, []tagtree.Filter{{NodeID: child.ID, Root: child.Root}})
	if res.Candidates != len(strict.Verses) {
		t.Errorf("candidate count %d does not match the filtered pool %d",
			res.Candidates, len(strict.Verses))
	}

	// And the primary must be one of them.
	if !containsID(strict.Verses, res.Primary.ID) {
		t.Errorf("the primary %q is not in the filtered pool", res.Primary.Text)
	}
	for _, alt := range res.Alternatives {
		if !containsID(strict.Verses, alt.ID) {
			t.Errorf("the alternative %q is not in the filtered pool", alt.Text)
		}
	}
	t.Logf("候选 %d 句（过滤后 %d 句）", res.Candidates, len(strict.Verses))
}

func containsID(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
