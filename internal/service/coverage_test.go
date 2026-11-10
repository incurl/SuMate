package service

import (
	"strconv"
	"testing"

	"github.com/sumate/sumate/internal/model"
)

// TestTagCoverage reports how many verses reach every node of the tag tree.
// A node is reached when a verse carries one of its tags or of any of its
// descendants, because the tree filters by prefix.
func TestTagCoverage(t *testing.T) {
	s := newService(t)
	all := s.AllVerses()

	direct := map[string]int{}
	for _, v := range all {
		for _, tags := range v.Annotation.Tags {
			for _, tag := range tags {
				direct[tag]++
			}
		}
	}

	effective := map[string]int{}
	for _, n := range s.tree.All() {
		n := n
		for _, v := range all {
			for _, tags := range v.Annotation.Tags {
				for _, tag := range tags {
					if tag == n.ID || len(tag) > len(n.ID) && tag[:len(n.ID)+1] == n.ID+"." {
						effective[n.ID]++
						break
					}
				}
			}
		}
	}

	var orphans []string
	for _, n := range s.tree.All() {
		if effective[n.ID] == 0 {
			orphans = append(orphans, n.ID+" ("+n.Label+")")
		}
	}
	t.Logf("total nodes=%d, unreachable=%d", len(s.tree.All()), len(orphans))
	for _, o := range orphans {
		t.Logf("  unreachable: %s", o)
	}

	for _, root := range model.RootOrder {
		var parts []string
		for _, n := range s.tree.All() {
			if n.Parent == string(root) {
				parts = append(parts, n.Label+"="+strconv.Itoa(effective[n.ID]))
			}
		}
		t.Logf("%s: %v", root, parts)
	}
}
