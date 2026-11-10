package tagtree

import (
	"testing"

	"github.com/sumate/sumate/internal/model"
)

func TestBuildShape(t *testing.T) {
	tree := Build()
	all := tree.All()
	if len(all) < 100 {
		t.Fatalf("expected a substantial tag tree, got %d nodes", len(all))
	}
	roots := tree.Roots()
	if len(roots) != 5 {
		t.Fatalf("expected 5 roots, got %d: %v", len(roots), roots)
	}
	depths := map[int]int{}
	for _, n := range all {
		depths[n.Depth]++
		if n.Depth > 3 {
			t.Fatalf("node %s deeper than expected (%d)", n.ID, n.Depth)
		}
	}
	t.Logf("nodes=%d depth1=%d depth2=%d depth3=%d",
		len(all), depths[1], depths[2], depths[3])
	if depths[3] == 0 {
		t.Fatal("expected at least one node at depth 3")
	}
}

func TestEveryNodeHasLabelAndPath(t *testing.T) {
	tree := Build()
	for _, n := range tree.All() {
		if n.Label == "" {
			t.Errorf("node %s has no label", n.ID)
		}
		if len(n.Path) == 0 {
			t.Errorf("node %s has no path", n.ID)
		}
		if n.Depth == 0 && n.Root != model.RootKey(n.ID) {
			t.Errorf("root %s has mismatched root %s", n.ID, n.Root)
		}
		if n.Depth > 0 && n.Root == "" {
			t.Errorf("child %s has no root", n.ID)
		}
	}
}
