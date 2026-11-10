// Package service exposes the SuMate (东坡嘴替) domain to the Wails
// frontend.  It is the single entry point the UI talks to.
package service

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/sumate/sumate/internal/annotate"
	"github.com/sumate/sumate/internal/corpus"
	"github.com/sumate/sumate/internal/engine"
	"github.com/sumate/sumate/internal/model"
	"github.com/sumate/sumate/internal/tagtree"
)

// Service is the application service.
type Service struct {
	mu sync.RWMutex

	tree   *tagtree.Tree
	verses []model.Verse
	engine *engine.Engine

	// verseIndex lets the tag tree filter the verse set.
	index []tagtree.VerseIndex
}

// verseRef adapts a verse to the tagtree.VerseIndex interface.
type verseRef struct {
	v model.Verse
}

// ID returns the verse id.
func (r verseRef) ID() string { return r.v.ID }

// Has reports whether the verse carries a tag in the given root.
func (r verseRef) Has(root model.RootKey, tagID string) bool {
	for _, t := range r.v.Annotation.Tags[root] {
		if t == tagID || strings.HasPrefix(t, tagID+".") {
			return true
		}
	}
	return false
}

// HasAny reports whether the verse carries the tag in any root.
func (r verseRef) HasAny(tagID string) bool {
	for _, tags := range r.v.Annotation.Tags {
		for _, t := range tags {
			if t == tagID || strings.HasPrefix(t, tagID+".") {
				return true
			}
		}
	}
	return false
}

// New builds the service from the embedded corpus.
func New() (*Service, error) {
	c, err := corpus.Load()
	if err != nil {
		return nil, err
	}
	eng := annotate.New()
	verses, err := corpus.NewPipeline(c, eng).Build()
	if err != nil {
		return nil, err
	}
	s := &Service{
		tree:   tagtree.Build(),
		verses: verses,
		engine: engine.New(verses),
	}
	s.index = make([]tagtree.VerseIndex, len(verses))
	for i, v := range verses {
		s.index[i] = verseRef{v: v}
	}
	return s, nil
}

// Stats returns a summary of the loaded corpus.
func (s *Service) Stats() model.Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lines, couplets := 0, 0
	tunes := map[string]bool{}
	concepts := map[string]bool{}
	for _, v := range s.verses {
		if v.Kind == "couplet" {
			couplets++
		} else {
			lines++
		}
		tunes[v.Tune] = true
		for _, c := range v.Annotation.Concepts {
			concepts[c] = true
		}
	}
	return model.Stats{
		Poems:    len(s.verses) / 2, // approximate; refined below
		Verses:   len(s.verses),
		Tunes:    len(tunes),
		Lines:    lines,
		Couplets: couplets,
		Concepts: len(concepts),
	}
}

// Roots returns the five root nodes with their live candidate counts.
func (s *Service) Roots() []model.TreeNode {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := s.tree.Apply(s.index, nil)
	out := make([]model.TreeNode, 0, 5)
	for _, id := range s.tree.Roots() {
		n, ok := s.tree.Node(id)
		if !ok {
			continue
		}
		out = append(out, model.TreeNode{
			TagNode:   *n,
			Remaining: res.Counts[id],
			IsLeaf:    len(n.Children) == 0,
		})
	}
	return out
}

// Children returns the children of a node with their live counts.
func (s *Service) Children(nodeID string, path []tagtree.Filter) []model.TreeNode {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := s.tree.ApplyAt(s.index, path, nodeID)
	out := make([]model.TreeNode, 0)
	for _, id := range s.tree.SortedChildren(nodeID, res.Counts) {
		n, ok := s.tree.Node(id)
		if !ok {
			continue
		}
		out = append(out, model.TreeNode{
			TagNode:   *n,
			Remaining: res.Counts[id],
			IsLeaf:    len(n.Children) == 0,
		})
	}
	return out
}

// Preview returns the verses matching the current selection, without
// requiring the user to drill down to a leaf.
func (s *Service) Preview(path []tagtree.Filter, limit int) []model.Verse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := s.tree.Apply(s.index, path)
	if limit <= 0 || limit > len(res.Verses) {
		limit = len(res.Verses)
	}
	byID := map[string]model.Verse{}
	for _, v := range s.verses {
		byID[v.ID] = v
	}
	out := make([]model.Verse, 0, limit)
	for _, id := range res.Verses[:limit] {
		out = append(out, byID[id])
	}
	return out
}

// Breadcrumb builds the breadcrumb trail for the current selection.
func (s *Service) Breadcrumb(path []tagtree.Filter) []model.Breadcrumb {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// The counts are recomputed for every prefix of the path so that each
	// step shows how many candidates remained at that moment.
	out := make([]model.Breadcrumb, 0, len(path))
	for i := range path {
		res := s.tree.Apply(s.index, path[:i+1])
		n, ok := s.tree.Node(path[i].NodeID)
		if !ok {
			continue
		}
		out = append(out, model.Breadcrumb{
			NodeID:    n.ID,
			Label:     n.Label,
			Root:      n.Root,
			Path:      n.Path,
			Remaining: len(res.Verses),
		})
	}
	return out
}

// Match runs the final convergence matching rule.
func (s *Service) Match(path []tagtree.Filter, description string) model.MatchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sel := engine.Selection{
		Description: description,
		Tags:        map[model.RootKey][]string{},
	}
	for _, f := range path {
		if sel.Tags[f.Root] == nil {
			sel.Tags[f.Root] = []string{}
		}
		sel.Tags[f.Root] = append(sel.Tags[f.Root], f.NodeID)
	}
	return s.engine.Match(sel)
}

// Verse returns a single verse by id.
//
// The zero verse is returned when the id is unknown; the frontend renders
// that as an empty card rather than having to deal with an error channel.
func (s *Service) Verse(id string) model.Verse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.verses {
		if v.ID == id {
			return v
		}
	}
	return model.Verse{}
}

// Search returns verses whose text or concepts match the query.
func (s *Service) Search(query string, limit int) []model.Verse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 20
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}
	var out []model.Verse
	for _, v := range s.verses {
		if strings.Contains(v.Text, q) ||
			strings.Contains(v.Annotation.Literal, q) ||
			strings.Contains(v.Annotation.Conception, q) ||
			containsAny(v.Annotation.Keywords, q) {
			out = append(out, v)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

// containsAny reports whether any of the keywords contains the query.
func containsAny(keywords []string, q string) bool {
	for _, k := range keywords {
		if strings.Contains(k, q) {
			return true
		}
	}
	return false
}

// AllVerses returns every verse, for the browse view.
func (s *Service) AllVerses() []model.Verse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]model.Verse{}, s.verses...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// String makes the service printable for debugging.
func (s *Service) String() string {
	return fmt.Sprintf("Service{verses:%d}", len(s.verses))
}
