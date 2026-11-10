// Package engine implements the SuMate (东坡嘴替) matching algorithm.
//
// The scoring is deliberately *not* keyword based.  A verse is scored on
// four independent axes:
//
//   - semantic fit: how close its artistic conception is to the user's
//     described situation;
//   - emotional resonance: how well its graded emotion profile matches the
//     emotional tags the user selected;
//   - scenario consistency: how well its situational tags overlap the
//     environment the user described;
//   - Dongpo temperament match: how strongly it carries the healing,
//     open-minded, transcending attitude that defines the product.
//
// A fifth factor, the product's own bias, gently prefers verses that carry
// Su Dongpo's characteristic transcendence, so that the answer always sounds
// like Dongpo rather than like a sentiment database.
package engine

import (
	"math"
	"sort"
	"strings"

	"github.com/incurl/sumate/internal/curated"
	"github.com/incurl/sumate/internal/model"
	"github.com/incurl/sumate/internal/philosophy"
)

// Weights of the four scoring axes.
const (
	WeightSemantic    = 0.30
	WeightEmotional   = 0.28
	WeightScenario    = 0.24
	WeightTemperament = 0.18
)

// Candidate is a verse with its computed score.
type Candidate struct {
	Verse  model.Verse
	Score  float64
	Parts  model.MatchExplanation
	Hits   []string
	Reason []string
}

// Engine matches verses against a selection.
type Engine struct {
	verses []model.Verse
}

// New returns an engine over the given verse set.
func New(verses []model.Verse) *Engine {
	return &Engine{verses: verses}
}

// Verses returns the full verse set.
func (e *Engine) Verses() []model.Verse { return e.verses }

// Selection is the user's state: the tags chosen per dimension plus the
// optional free-text description.
type Selection struct {
	// Tags maps a root dimension to the selected leaf tag ids.
	Tags map[model.RootKey][]string
	// Description is an optional free-text hint from the user.
	Description string
}

// scored holds the intermediate result of scoring one verse.
type scored struct {
	v          model.Verse
	semantic   float64
	emotional  float64
	scenario   float64
	temperam   float64
	philosophy float64
	hits       []string
	reasons    []string
	caveats    []model.PhilosophicalCaveat
}

// Match ranks the verses against the selection and returns the top result
// together with alternatives.
func (e *Engine) Match(sel Selection) model.MatchResult {
	cands := e.Rank(sel, 0)
	result := model.MatchResult{}
	if len(cands) == 0 {
		return result
	}
	result.Candidates = len(cands)
	result.Primary = cands[0].Verse
	result.Explanation = cands[0].Parts
	limit := 3
	if len(cands)-1 < limit {
		limit = len(cands) - 1
	}
	for i := 1; i <= limit; i++ {
		result.Alternatives = append(result.Alternatives, cands[i].Verse)
	}
	return result
}

// Rank scores every verse against the selection and returns them sorted.
func (e *Engine) Rank(sel Selection, limit int) []Candidate {
	scoredAll := make([]scored, 0, len(e.verses))
	for _, v := range e.verses {
		s := e.score(v, sel)
		scoredAll = append(scoredAll, s)
	}
	sort.SliceStable(scoredAll, func(i, j int) bool {
		return scoredAll[i].total() > scoredAll[j].total()
	})
	out := make([]Candidate, 0, len(scoredAll))
	for _, s := range scoredAll {
		parts := model.MatchExplanation{
			Overall:     int(math.Round(s.total() * 100)),
			Semantic:    int(math.Round(s.semantic * 100)),
			Emotional:   int(math.Round(s.emotional * 100)),
			Scenario:    int(math.Round(s.scenario * 100)),
			Temperament: int(math.Round(s.temperam * 100)),
			Reasons:     s.reasons,
			MatchedTags: s.hits,
			Caveats:     s.caveats,
			Philosophy:  philosophySummary(s.caveats),
		}
		out = append(out, Candidate{
			Verse:  s.v,
			Score:  s.total(),
			Parts:  parts,
			Hits:   s.hits,
			Reason: s.reasons,
		})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

// total computes the composite score of a scored verse.
func (s scored) total() float64 {
	base := WeightSemantic*s.semantic +
		WeightEmotional*s.emotional +
		WeightScenario*s.scenario +
		WeightTemperament*s.temperam
	// The product's own bias: prefer verses that carry Dongpo's healing
	// attitude, so the answer always sounds like Dongpo.
	base += 0.04 * s.temperam
	// The philosophical fit.  A verse that would mislead is demoted rather
	// than hidden: the user still sees the caveat, but a verse whose
	// premise holds wins.
	base += 0.06 * s.philosophy
	if base > 1 {
		base = 1
	}
	return base
}

// score computes the four axes for one verse.
func (e *Engine) score(v model.Verse, sel Selection) scored {
	s := scored{v: v}
	a := v.Annotation

	// --- emotional axis: the user selected emotional tags.
	selEmo := e.selectedEmotions(sel)
	if len(selEmo) == 0 {
		s.emotional = 0.55
	} else {
		var hit, total float64
		for key, strength := range selEmo {
			total++
			for _, em := range a.Emotions {
				if em.ID != key {
					continue
				}
				hit += overlap(strength, em.Strength)
				s.hits = append(s.hits, "情绪·"+em.Name)
				break
			}
		}
		if total > 0 {
			s.emotional = hit / total
		}
		// A verse with no emotional tags at all cannot resonate.
		if len(a.Emotions) == 0 {
			s.emotional *= 0.5
		}
	}

	// --- scenario axis: the user selected situational / environmental tags.
	selTags := e.selectedTags(sel)
	if len(selTags) == 0 {
		s.scenario = 0.55
	} else {
		var hit, total float64
		for _, tag := range selTags {
			total++
			if verseHasTag(v, tag) {
				hit++
				s.hits = append(s.hits, tagLabel(tag))
			}
		}
		if total > 0 {
			s.scenario = hit / total
		}
		// Sub-tags count as a partial match for their parent.
		s.scenario = math.Min(1, s.scenario*1.15)
	}

	// --- semantic axis: concept overlap + free text.
	s.semantic = semanticScore(v, sel)

	// --- temperament axis: Dongpo's own spiritual signature.
	s.temperam = temperamentScore(a)

	// --- philosophical axis: does this verse actually apply to them?
	//
	// This is the product's honesty check.  A verse that says 「此处心安即吾乡」
	// is comfort to a reader who has nowhere else to stand; to a reader who
	// still has a road to travel it is an excuse not to walk.  The verse
	// is not demoted out of sight -- it is demoted *with its caveat*, so the
	// reader can judge for themselves.
	reader := philosophy.Reader{Tags: sel.Tags}
	s.philosophy, s.caveats = stanceScore(v, reader)

	if len(s.reasons) == 0 {
		s.reasons = e.buildReasons(v, sel)
	}
	return s
}

// selectedEmotions resolves the selected mood tags to emotion keys.
//
// A mood tag may stand for several emotions; the strongest of them wins.
func (e *Engine) selectedEmotions(sel Selection) map[string]model.Strength {
	out := map[string]model.Strength{}
	moodMap := curated.MoodEmotionMap()
	for _, tag := range sel.Tags[model.RootMood] {
		me, ok := moodMap[tag]
		if !ok {
			continue
		}
		for _, key := range me.Emotions {
			if cur, ok := out[key]; !ok || cur.Rank() < me.Strength.Rank() {
				out[key] = me.Strength
			}
		}
	}
	return out
}

// selectedTags returns every selected tag across all dimensions.
func (e *Engine) selectedTags(sel Selection) []string {
	var out []string
	for _, list := range sel.Tags {
		out = append(out, list...)
	}
	return out
}

// verseHasTag reports whether a verse carries the given tag, either directly
// or through one of its ancestors.
func verseHasTag(v model.Verse, tag string) bool {
	for _, tags := range v.Annotation.Tags {
		for _, t := range tags {
			if t == tag || strings.HasPrefix(t, tag+".") {
				return true
			}
		}
	}
	return false
}

// tagLabel renders a tag id for display.
func tagLabel(tag string) string {
	i := strings.Index(tag, ".")
	if i < 0 {
		return tag
	}
	return tag[i+1:]
}

// overlap returns the similarity of two graded strengths.
func overlap(a, b model.Strength) float64 {
	r, s := a.Rank(), b.Rank()
	if r == 0 || s == 0 {
		return 0
	}
	d := r - s
	if d < 0 {
		d = -d
	}
	return 1 - float64(d)/4
}

// temperamentScore measures how strongly a verse carries Dongpo's healing
// attitude.
func temperamentScore(a model.Annotation) float64 {
	if len(a.Tangents) == 0 {
		return 0.25
	}
	score := 0.0
	for _, t := range a.Tangents {
		switch t {
		case model.TangentSorrowTranscend, model.TangentHardshipOptimism:
			// the two signature temperaments carry the most weight
			score += 0.26
		case model.TangentAcceptFate, model.TangentDetach:
			score += 0.20
		case model.TangentReconcile:
			score += 0.18
		case model.TangentTolerance:
			score += 0.16
		}
	}
	if score > 1 {
		score = 1
	}
	return score
}

// semanticScore measures the artistic-conception fit of a verse.
func semanticScore(v model.Verse, sel Selection) float64 {
	a := v.Annotation
	score := 0.35
	if len(a.Concepts) > 0 {
		score += 0.15
	}
	if sel.Description != "" {
		score = math.Max(score, descriptionScore(v, sel.Description))
	}
	// A verse with a rich philosophical layer is a better "mouthpiece".
	switch {
	case a.Philosophy != "" && a.Conception != "":
		score += 0.1
	}
	return math.Min(score, 1)
}

// descriptionScore measures the overlap between the user's free text and the
// verse's own vocabulary.
func descriptionScore(v model.Verse, desc string) float64 {
	descRunes := map[rune]bool{}
	for _, r := range desc {
		if r >= 0x4e00 && r <= 0x9fff {
			descRunes[r] = true
		}
	}
	if len(descRunes) == 0 {
		return 0
	}
	hit := 0
	for _, kw := range v.Annotation.Keywords {
		for _, r := range kw {
			if descRunes[r] {
				hit++
				break
			}
		}
	}
	if hit == 0 {
		return 0
	}
	return math.Min(1, float64(hit)/float64(len(v.Annotation.Keywords)+1))
}

// buildReasons explains why a verse matches.
func (e *Engine) buildReasons(v model.Verse, sel Selection) []string {
	var reasons []string
	a := v.Annotation

	if a.Conception != "" {
		reasons = append(reasons, a.Conception)
	}
	if len(a.Philosophy) > 0 {
		reasons = append(reasons, a.Philosophy)
	}
	if len(a.Tangents) > 0 {
		var names []string
		for _, t := range a.Tangents {
			names = append(names, model.TangentName[t])
		}
		reasons = append(reasons, "东坡气质："+strings.Join(names, "、"))
	}
	if len(a.Emotions) > 0 {
		var names []string
		for i, em := range a.Emotions {
			if i >= 3 {
				break
			}
			names = append(names, em.Name+"("+strengthLabel(em.Strength)+")")
		}
		reasons = append(reasons, "情绪："+strings.Join(names, "、"))
	}
	if a.Attitude != "" {
		reasons = append(reasons, "态度："+a.Attitude)
	}
	if a.Healing != "" {
		reasons = append(reasons, "疗愈："+a.Healing)
	}
	return reasons
}

// strengthLabel renders a strength in Chinese.
func strengthLabel(s model.Strength) string {
	switch s {
	case model.Slight:
		return "轻微"
	case model.Moderate:
		return "中等"
	case model.Obvious:
		return "明显"
	case model.Profound:
		return "深重"
	}
	return ""
}

// MatchIn ranks a given pool of verses against the selection.
//
// It is used when the strict intersection of the selected tags is empty and
// the product has to fall back to a relaxed pool, so that the user always
// receives a verse.
func (e *Engine) MatchIn(pool []model.Verse, sel Selection) model.MatchResult {
	sub := &Engine{verses: pool}
	return sub.Match(sel)
}

// stanceScore measures how well a verse's philosophical stance fits
// the reader, and collects the caveats.
//
// The score rewards a verse whose stance holds for the reader and penalises
// one whose premise has failed.  A verse with no stance at all gets a
// middling score: it is not wrong, it simply does not speak to the
// question.
func stanceScore(v model.Verse, reader philosophy.Reader) (float64, []model.PhilosophicalCaveat) {
	stances := v.Annotation.Stances
	if len(stances) == 0 {
		return 0.5, nil
	}
	lex := curated.Stances()

	var sum float64
	var caveats []model.PhilosophicalCaveat
	for _, key := range stances {
		st, ok := lex[key]
		if !ok {
			continue
		}
		a := philosophy.Assess(st, reader)
		switch a.Fit {
		case philosophy.FitHeld:
			sum += 1.0
		case philosophy.FitStrained:
			sum += 0.6
		case philosophy.FitUnknown:
			sum += 0.5
		case philosophy.FitFailed:
			sum += 0.15
			caveats = append(caveats, model.PhilosophicalCaveat{
				Stance:  a.Stance,
				Name:    a.Name,
				Fit:     string(a.Fit),
				Premise: a.Premise,
				Shadow:  a.Shadow,
				Earned:  a.Earned,
				Note:    a.Note,
			})
		}
	}
	if sum == 0 {
		return 0.5, caveats
	}
	return sum / float64(len(stances)), caveats
}

// philosophySummary composes the prose shown above the match.
//
// The product's promise is that Su Dongpo speaks *for* the user.  That
// promise is only kept if the product also says when it cannot: a verse
// whose premise does not hold for this reader is shown together with the
// condition under which it would.
func philosophySummary(caveats []model.PhilosophicalCaveat) string {
	if len(caveats) == 0 {
		return ""
	}
	var b strings.Builder
	for i, c := range caveats {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(c.Note)
	}
	return b.String()
}
