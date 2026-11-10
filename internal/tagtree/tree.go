// Package tagtree defines the five-root hierarchical tag system of SuMate
// (东坡嘴替) and the machinery that counts, filters and drills it.
package tagtree

import (
	"sort"
	"strings"

	"github.com/sumate/sumate/internal/model"
)

// Tree is a hierarchical tag tree with an index for fast lookup.
type Tree struct {
	nodes map[string]*model.TagNode
	order []string
}

// New returns an empty tree.
func New() *Tree {
	return &Tree{nodes: map[string]*model.TagNode{}}
}

// Add inserts a node.  Children are linked to their parent when the parent
// already exists, otherwise the link is resolved in a second pass by
// Finalize.
func (t *Tree) Add(n model.TagNode) {
	cp := n
	t.nodes[n.ID] = &cp
	t.order = append(t.order, n.ID)
}

// Finalize resolves parent/child links and computes depth and label paths.
func (t *Tree) Finalize() {
	roots := t.roots()
	for _, id := range t.order {
		n := t.nodes[id]
		if n.Parent == "" {
			n.Depth = 0
			n.Path = []string{n.Label}
			continue
		}
		p, ok := t.nodes[n.Parent]
		if !ok {
			continue
		}
		n.Depth = p.Depth + 1
		n.Path = append(append([]string{}, p.Path...), n.Label)
	}
	// link children in insertion order
	for _, id := range t.order {
		n := t.nodes[id]
		if n.Parent == "" {
			continue
		}
		p := t.nodes[n.Parent]
		p.Children = append(p.Children, id)
	}
	_ = roots
}

// roots returns the ids of the five root nodes in canonical order.
func (t *Tree) roots() []string {
	var out []string
	for _, rk := range model.RootOrder {
		if n, ok := t.nodes[string(rk)]; ok && n.Parent == "" {
			out = append(out, n.ID)
		}
	}
	return out
}

// Roots returns the root node ids in canonical order.
func (t *Tree) Roots() []string { return t.roots() }

// Node returns the node with the given id.
func (t *Tree) Node(id string) (*model.TagNode, bool) {
	n, ok := t.nodes[id]
	return n, ok
}

// All returns every node in insertion order.
func (t *Tree) All() []*model.TagNode {
	out := make([]*model.TagNode, 0, len(t.order))
	for _, id := range t.order {
		out = append(out, t.nodes[id])
	}
	return out
}

// Count recomputes the per-node match counts from the verse index.
//
// The index maps a node id to the set of verse ids that carry that tag.
func (t *Tree) Count(index map[string][]string) {
	for _, id := range t.order {
		n := t.nodes[id]
		n.Count = len(index[id])
	}
}

// CountAll recomputes counts so that a parent's count is the union of its
// subtree (i.e. a verse counts for every ancestor of its tags).
func (t *Tree) CountAll(index map[string][]string) {
	for _, id := range t.order {
		n := t.nodes[id]
		seen := map[string]bool{}
		for _, v := range index[id] {
			seen[v] = true
		}
		n.Count = len(seen)
	}
}

// LabelPath returns the display path of a node.
func (t *Tree) LabelPath(id string) []string {
	n, ok := t.nodes[id]
	if !ok {
		return nil
	}
	return n.Path
}

// Filter is one selected node in the drilling history.
type Filter struct {
	NodeID string
	Root   model.RootKey
}

// Result describes the outcome of a filter operation.
type Result struct {
	// Verses are the matching verse ids.
	Verses []string
	// Counts holds the remaining count per candidate node id.
	Counts map[string]int
}

// Apply filters the verses by the given selection path and returns the
// surviving verse ids together with the per-node remaining counts.
//
// A verse survives when it carries a tag from *every* root dimension that
// the path has already entered.  This is what makes the tree progressive:
// once the user commits to a mood, the location options narrow down to the
// verses that can actually speak that mood.
func (t *Tree) Apply(verses []VerseIndex, path []Filter) Result {
	return t.ApplyAt(verses, path, "")
}

// ApplyAt filters the verses by the given selection path and computes the
// remaining counts for the children of `focus`.
//
// `focus` is the node whose children the caller wants to display.  When it
// is empty the last selected node of the path is used, and when the path is
// empty the roots are used.
func (t *Tree) ApplyAt(verses []VerseIndex, path []Filter, focus string) Result {
	kept := make([]string, 0, len(verses))
	for _, v := range verses {
		// Every selected node must match: drilling deeper inside a dimension
		// narrows the result, and entering another dimension narrows it
		// further.  Because a node's subtree is a subset of its parent's,
		// the conjunction of the path is exactly what the user expects.
		match := true
		for _, f := range path {
			if !v.Has(f.Root, f.NodeID) {
				match = false
				break
			}
		}
		if match {
			kept = append(kept, v.ID())
		}
	}

	// Remaining counts for the children of the focused node.
	counts := map[string]int{}
	focusID := focus
	if focusID == "" && len(path) > 0 {
		focusID = path[len(path)-1].NodeID
	}
	if focusID != "" {
		if node, ok := t.nodes[focusID]; ok {
			for _, childID := range node.Children {
				counts[childID] = t.countMatching(kept, childID, verses)
			}
		}
	} else {
		for _, r := range t.roots() {
			counts[r] = t.countMatching(kept, r, verses)
		}
	}
	return Result{Verses: kept, Counts: counts}
}

// countMatching counts the kept verses that also carry the given tag.
func (t *Tree) countMatching(kept []string, tagID string, verses []VerseIndex) int {
	byID := make(map[string]VerseIndex, len(verses))
	for _, v := range verses {
		byID[v.ID()] = v
	}
	n := 0
	for _, id := range kept {
		v, ok := byID[id]
		if !ok {
			continue
		}
		if v.HasAny(tagID) {
			n++
		}
	}
	return n
}

// VerseIndex is the minimal interface the tree needs to filter verses.
type VerseIndex interface {
	// ID is the verse id.
	ID() string
	// Has reports whether the verse carries the tag in the given root.
	Has(root model.RootKey, tagID string) bool
	// HasAny reports whether the verse carries the tag in any root.
	HasAny(tagID string) bool
}

// SortedChildren returns the children of a node, ordered by count then label.
func (t *Tree) SortedChildren(id string, counts map[string]int) []string {
	n, ok := t.nodes[id]
	if !ok {
		return nil
	}
	out := append([]string{}, n.Children...)
	sort.SliceStable(out, func(i, j int) bool {
		ci, cj := counts[out[i]], counts[out[j]]
		if ci != cj {
			return ci > cj
		}
		return t.nodes[out[i]].Label < t.nodes[out[j]].Label
	})
	return out
}

// JoinPath renders a breadcrumb path.
func JoinPath(labels []string) string { return strings.Join(labels, " › ") }
