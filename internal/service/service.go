// Package service exposes the SuMate (东坡嘴替) domain to the Wails
// frontend.  It is the single entry point the UI talks to.
package service

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/incurl/sumate/internal/annotate"
	"github.com/incurl/sumate/internal/corpus"
	"github.com/incurl/sumate/internal/curated"
	"github.com/incurl/sumate/internal/engine"
	"github.com/incurl/sumate/internal/model"
	"github.com/incurl/sumate/internal/philosophy"
	"github.com/incurl/sumate/internal/tagtree"
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

// Rank ranks the verses that match the given path, using the same pool
// and the same rule as Match.
//
// It exists so that the audit can measure the axes over the candidates the
// user actually sees.  Measuring over the whole corpus -- which is what the
// engine alone allows -- answers a different question: it says how the axes
// behave among verses the user has already excluded.
func (s *Service) Rank(path []tagtree.Filter, description string, limit int) []engine.Candidate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sel := engine.Selection{
		Description: description,
		Tags:        map[model.RootKey][]string{},
	}
	for _, f := range path {
		sel.Tags[f.Root] = append(sel.Tags[f.Root], f.NodeID)
	}

	strict := s.tree.Apply(s.index, path)
	if len(strict.Verses) > 0 {
		return s.engine.RankIn(s.versesIn(strict.Verses), sel, limit)
	}
	relaxed := s.tree.Relaxed(s.index, path)
	return s.engine.RankIn(s.versesIn(relaxed), sel, limit)
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

	// The candidate pool is the strict intersection of the selected tags.
	// When that pool is empty -- which happens when the user over-specifies
	// the selection across several dimensions -- the pool is relaxed to the
	// union of the per-dimension matches, so that the user always receives a
	// verse instead of an empty screen.
	strict := s.tree.Apply(s.index, path)
	if len(strict.Verses) > 0 {
		return s.engine.MatchIn(s.versesIn(strict.Verses), sel)
	}

	relaxed := s.tree.Relaxed(s.index, path)
	return s.engine.MatchIn(s.versesIn(relaxed), sel)
}

// versesIn returns the verses with the given ids, in corpus order.
//
// The order is the corpus order on purpose: the engine's ranking is
// order-independent (see internal/engine/arbitrary_test.go), so the pool's
// order cannot change which verse wins.
func (s *Service) versesIn(ids []string) []model.Verse {
	keep := make(map[string]bool, len(ids))
	for _, id := range ids {
		keep[id] = true
	}
	pool := make([]model.Verse, 0, len(ids))
	for _, v := range s.verses {
		if keep[v.ID] {
			pool = append(pool, v)
		}
	}
	return pool
}

// ProductStance returns the position the product itself takes.
//
// A tool that promises to find your voice has a duty to declare what it
// believes and when it cannot serve.
func (s *Service) ProductStance() curated.ProductStance {
	return curated.TheProductStance()
}

// PhilosophicalReading returns the philosophical assessment of the reader's
// current situation.
//
// It answers the question the tree cannot: not "which verses mention this",
// but "which positions on how to live are available to you right now, and
// on what condition".
func (s *Service) PhilosophicalReading(path []tagtree.Filter) []philosophy.Assessment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tags := map[model.RootKey][]string{}
	for _, f := range path {
		tags[f.Root] = append(tags[f.Root], f.NodeID)
	}
	return philosophy.AssessAll(philosophy.Reader{Tags: tags})
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
