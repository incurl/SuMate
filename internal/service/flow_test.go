package service

import (
	"strings"
	"testing"

	"github.com/incurl/sumate/internal/model"
	"github.com/incurl/sumate/internal/tagtree"
)

// TestProgressiveDrillingFlow walks the whole product loop the way a user
// would: pick a dimension, drill down, switch to another dimension, drill
// there too, backtrack through the breadcrumb, preview, and finally match.
func TestProgressiveDrillingFlow(t *testing.T) {
	s := newService(t)

	roots := s.Roots()
	if len(roots) != 5 {
		t.Fatalf("expected 5 roots, got %d", len(roots))
	}

	// --- step 1: enter the mood dimension.
	mood := roots[0]
	moodKids := s.Children(mood.ID, nil)
	if len(moodKids) == 0 {
		t.Fatal("mood dimension has no children")
	}
	t.Logf("mood children: %d, first = %s (%d)",
		len(moodKids), moodKids[0].Label, moodKids[0].Remaining)

	// --- step 2: drill to the second level.
	lown := moodKids[0]
	path := []tagtree.Filter{{NodeID: mood.ID, Root: mood.Root}}
	lvl2 := s.Children(lown.ID, path)
	if len(lvl2) == 0 {
		t.Fatal("no second level")
	}
	path = append(path, tagtree.Filter{NodeID: lown.ID, Root: lown.Root})

	// --- step 3: a second dimension on top of the first one.
	placeKids := s.Children(roots[1].ID, path)
	if len(placeKids) == 0 {
		t.Fatal("no place children under the mood selection")
	}
	path = append(path, tagtree.Filter{NodeID: roots[1].ID, Root: roots[1].Root})

	// The candidate count must have shrunk with every step.
	crumbs := s.Breadcrumb(path)
	if len(crumbs) != 3 {
		t.Fatalf("expected 3 breadcrumb steps, got %d", len(crumbs))
	}
	for i := 1; i < len(crumbs); i++ {
		if crumbs[i].Remaining > crumbs[i-1].Remaining {
			t.Errorf("step %d has more candidates (%d) than step %d (%d)",
				i, crumbs[i].Remaining, i-1, crumbs[i-1].Remaining)
		}
	}
	t.Logf("breadcrumb: ")
	for _, c := range crumbs {
		t.Logf("   %-28s %d", strings.Join(c.Path, " > "), c.Remaining)
	}

	// --- step 4: backtrack to the first step.
	//
	// Dropping the later steps must restore the wider candidate count that
	// the first step had.
	back := s.Breadcrumb(path[:1])
	if len(back) != 1 {
		t.Fatalf("expected 1 step after backtracking, got %d", len(back))
	}
	if back[0].Remaining != crumbs[0].Remaining {
		t.Errorf("backtracking should restore the count of step 1: %d != %d",
			back[0].Remaining, crumbs[0].Remaining)
	}

	// --- step 5: preview at an arbitrary layer.
	preview := s.Preview(path, 20)
	if len(preview) == 0 {
		t.Fatal("preview returned nothing")
	}
	for _, v := range preview {
		if !carriesTag(v, roots[0].ID) {
			t.Errorf("preview verse %s does not carry the selected mood", v.ID)
		}
	}
	t.Logf("preview at the combined selection: %d verses", len(preview))

	// --- step 6: the final match.
	res := s.Match(path, "有点累，但还想再撑一会儿")
	if res.Primary.ID == "" {
		t.Fatal("no primary match")
	}
	if len(res.Alternatives) < 2 {
		t.Errorf("expected alternatives, got %d", len(res.Alternatives))
	}
	if res.Explanation.Overall <= 0 || len(res.Explanation.Reasons) == 0 {
		t.Error("the match carries no explanation")
	}
	t.Logf("match: %s  (score %d)", res.Primary.Text, res.Explanation.Overall)
	for _, r := range res.Explanation.Reasons {
		t.Logf("   · %s", r)
	}
	for _, a := range res.Alternatives {
		t.Logf("   alt: %s", a.Text)
	}
}

// carriesTag reports whether a verse carries a tag or one of its descendants.
func carriesTag(v model.Verse, tag string) bool {
	for _, tags := range v.Annotation.Tags {
		for _, t := range tags {
			if t == tag || strings.HasPrefix(t, tag+".") {
				return true
			}
		}
	}
	return false
}
