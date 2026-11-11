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
	result.Confidence, result.Confident, result.Honest = confidence(cands, sel)
	return result
}

// Confidence thresholds.
//
// Below the floor the product does not claim to have found the verse; it
// says so.  Between floor and good the answer is offered with a caveat.
const (
	// Below the floor the product does not claim to have found the verse.
	confidenceFloor = 32
	// At or above the good threshold it stands behind the answer.
	confidenceGood = 58
)

// confidence measures how well the top verse actually fits.
//
// The measure is deliberately conservative.  It looks at three things:
//
//   - the absolute score of the winner;
//   - the *margin* over the runner-up, because a verse that barely beats
//     its neighbour is not a confident answer;
//   - whether the winner's philosophical premise holds.
//
// The margin matters most.  A product that only looked at the absolute
// score would call a verse a perfect match whenever the whole corpus
// happened to be mediocre.
func confidence(cands []Candidate, sel Selection) (int, bool, string) {
	if len(cands) == 0 {
		return 0, false, ""
	}
	top := cands[0]
	score := top.Score * 100

	// The margin over the runner-up, expressed in points.
	margin := 0.0
	if len(cands) > 1 {
		margin = (top.Score - cands[1].Score) * 100
	}

	// A philosophical mismatch costs the answer its confidence, because a
	// consolation whose premise does not hold is not an answer.
	philo := top.Parts
	philosophyPenalty := 0.0
	if len(philo.Caveats) > 0 {
		philosophyPenalty = 12
	}

	// The score of the winner alone is a poor measure: the axes have
	// baselines (an unmatched verse still scores 0.35 on semantics), so a
	// bare selection would look confident.  The margin over the runner-up
	// and the *specificity* of the selection carry the real information.
	//
	// Specificity is the fraction of the four axes that the user actually
	// constrained: a mood alone constrains one, a mood plus a place plus a
	// weather constrains three.
	specificity := 0.0
	if len(sel.Tags) > 0 {
		for _, tags := range sel.Tags {
			if len(tags) > 0 {
				specificity++
			}
		}
		specificity /= 3.0 // three dimensions is a full description
		if specificity > 1 {
			specificity = 1
		}
	}

	conf := score*0.35 + margin*0.45 + specificity*30 - philosophyPenalty
	if conf > 100 {
		conf = 100
	}
	if conf < 0 {
		conf = 0
	}
	c := int(conf)

	if c >= confidenceGood {
		return c, true, ""
	}
	if c < confidenceFloor {
		// Below the floor the product says plainly that it is reaching.
		return c, false, "这一句未必就是那一句。你的处境还没有被说到足够具体，" +
			"我只能给出一个方向。再往细处选一层，或者写下你此刻的感觉，会更准。"
	}

	// The honest note: why the answer is weak, and what would fix it.
	var reasons []string
	if margin < 6 {
		reasons = append(reasons, "几句与这一句其实差不多近")
	}
	if len(philo.Caveats) > 0 {
		reasons = append(reasons, "这一句的前提与你的处境并不完全相合")
	}
	if specificity < 0.4 {
		reasons = append(reasons, "你还没有说到具体的处境")
	}
	if score < 55 {
		reasons = append(reasons, "现有的选择还不足以指向某一句")
	}
	note := "这一句未必就是那一句。"
	if len(reasons) > 0 {
		note += "（" + strings.Join(reasons, "；") + "）"
	}
	note += "再往细处选一层，或者写下你此刻的感觉，会更准。"
	return c, false, note
}

// Rank scores every verse against the selection and returns them sorted.
func (e *Engine) Rank(sel Selection, limit int) []Candidate {
	scoredAll := make([]scored, 0, len(e.verses))
	for _, v := range e.verses {
		s := e.score(v, sel)
		scoredAll = append(scoredAll, s)
	}
	sort.SliceStable(scoredAll, func(i, j int) bool {
		return betterVerse(scoredAll[i], scoredAll[j])
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
	//
	// The score is *graded* rather than binary.  A verse that carries the
	// selected tag exactly is a perfect match; one that only carries a
	// deeper tag beneath it (the user picked `occasion.exile`, the verse
	// says `occasion.exile.hardship`) matches in spirit but not in
	// specificity.
	//
	// This matters because the candidate pool is already filtered by the
	// selected tags, so a binary score was a *constant* within the pool:
	// every surviving verse scored 1.0, and the axis could not break a
	// single tie.  Grading it lets the axis say which verse addresses the
	// user's situation most precisely.
	selTags := e.selectedTags(sel)
	if len(selTags) == 0 {
		s.scenario = 0.55
	} else {
		var hit, total float64
		for _, tag := range selTags {
			total++
			m := scenarioMatch(v, tag)
			hit += m
			if m > 0 {
				s.hits = append(s.hits, tagLabel(tag))
			}
		}
		if total > 0 {
			s.scenario = hit / total
		}
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
//
// The axis used to be a near-constant: it checked only whether the verse had
// any concepts and whether its philosophy field was non-empty -- two things
// that are true of every verse that reaches the top of the ranking.  Its
// standard deviation over the decision region was therefore 0.00, and a
// third of the total weight produced no difference at all.
//
// It now measures something that varies: how *specific* the verse's meaning
// is.  A verse carrying ten concepts, several of them deep (moon.reunion,
// human.brother), says something particular; a verse carrying two shallow
// ones says little.  Specificity is a real property of the text, it varies
// continuously across the candidates, and it needs no model to compute.
func semanticScore(v model.Verse, sel Selection) float64 {
	a := v.Annotation

	// Specificity: the concepts a verse carries, weighted by how deep they
	// sit in the lexicon.  A deep concept names a particular state; a
	// shallow one names a broad region.
	//
	// The normalisation is measured, not guessed: across the corpus the
	// depth sum runs from 2 (p10) through 8 (median) to 49 (max).  Dividing
	// by the 90th percentile and *soft*-saturating keeps the ordering
	// visible: a plain min(1, x/24) made every concept-rich verse tie at
	// the ceiling, which merely moved the arbitrariness rather than
	// removing it.
	var specificity float64
	for _, ck := range a.Concepts {
		specificity += conceptDepth(ck)
	}
	score := softSaturation(specificity / specificityP90)

	// The free-text description, when given, *replaces* the specificity
	// score rather than competing with it.
	//
	// It has to replace it: the user has just said something specific, and
	// a verse that merely happens to be concept-rich must not outrank one
	// that actually understands them.  Taking the maximum let the
	// specificity of unrelated verses win, and the description stopped
	// mattering again.
	if sel.Description != "" {
		return descriptionScore(v, sel.Description)
	}
	return score
}

// conceptDepth returns how specific a concept key is, by the depth of its
// name in the lexicon's dotted hierarchy (moon.reunion -> 2).
func conceptDepth(key string) float64 {
	n := 1
	for _, r := range key {
		if r == '.' {
			n++
		}
	}
	return float64(n)
}

// descriptionScore measures the overlap between the user's free text and the
// verse's own vocabulary.
func descriptionScore(v model.Verse, desc string) float64 {
	// The description is mapped onto the concept lexicon first, and the
	// verse is then scored by the concepts it carries.
	//
	// The earlier version intersected *single characters* with the keywords.
	// Because words like 我 / 了 / 想 appear in a large share of the
	// keywords, almost every description matched almost every verse, and
	// five different descriptions returned the same line.  Going through the
	// concepts fixes both the precision and the reasonability: the match
	// says *which state* was understood.
	cs := bridgeConcepts(desc)
	if len(cs) == 0 {
		return 0
	}
	hit := 0
	for _, ck := range v.Annotation.Concepts {
		if cs[ck] {
			hit++
		}
	}
	if hit == 0 {
		return 0
	}
	// Scale by how much of the description was understood, so that a
	// description naming two states does not beat one naming three.
	return math.Min(1, float64(hit)/math.Sqrt(float64(len(cs)+1)))
}

// bridgeConcepts maps a free-text description onto the concept lexicon.
//
// A multi-character keyword that occurs in the description is strong
// evidence; a single character is only consulted when no strong keyword
// matched at all.  This is the same strong/weak distinction the annotation
// engine applies to a verse.
func bridgeConcepts(desc string) map[string]bool {
	out := map[string]bool{}
	concepts := curated.Concepts()
	for key, c := range concepts {
		for _, kw := range c.Keywords {
			if len([]rune(kw)) < 2 {
				continue
			}
			if strings.Contains(desc, kw) {
				out[key] = true
				break
			}
		}
	}
	if len(out) > 0 {
		return out
	}
	// No specific phrase matched: fall back to single characters.
	for key, c := range concepts {
		for _, kw := range c.Keywords {
			if len([]rune(kw)) != 1 {
				continue
			}
			if strings.Contains(desc, kw) {
				out[key] = true
				break
			}
		}
	}
	return out
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
// RankIn ranks the given pool of verses under the selection.
//
// It is the same ranking MatchIn performs, exposed so that the audit can
// look at the score decomposition of the candidates the user actually
// sees.
func (e *Engine) RankIn(pool []model.Verse, sel Selection, limit int) []Candidate {
	sub := &Engine{verses: pool}
	return sub.Rank(sel, limit)
}

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

// specificityP90 is the 90th percentile of the concept-depth sum across the
// corpus, measured with tools/audit.  A verse at the median scores about
// 0.5, one at the 90th percentile scores about 0.63, and the most
// concept-dense verse in the corpus reaches 1.
const specificityP90 = 16.0

// softSaturation maps a non-negative value onto (0, 1] without a hard
// ceiling, so that two verses that both exceed 1 are still distinguishable.
//
// The plain min(1, x) was the original bug: once a dozen verses were
// concept-rich enough to exceed the threshold, they all scored 1.0 and
// tied, and the tie was then broken by corpus order.
func softSaturation(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return x / (1 + x)
}

// betterVerse decides which of two verses comes first.
//
// The primary comparison is the total score.  When two verses score
// exactly the same -- which can happen for two distinct lines carrying the
// same concepts -- the order must still be *decided*, and it must not be
// decided by where the verse happens to sit in the corpus file.
//
// The rule is deliberately explainable rather than clever:
//
//  1. the shorter verse wins, because a short line is more like something a
//     person would actually say out loud;
//  2. then the verse ID, which is a stable machine key, so the result is
//     reproducible byte for byte.
//
// The second rule is arbitrary in the way that any last-resort tie-break
// must be, but it is *manifest*ly arbitrary: it does not depend on the
// order of the input, so reversing the corpus cannot change the answer.
func betterVerse(a, b scored) bool {
	ta, tb := a.total(), b.total()
	if ta != tb {
		return ta > tb
	}
	la, lb := len([]rune(a.v.Text)), len([]rune(b.v.Text))
	if la != lb {
		return la < lb
	}
	return a.v.ID < b.v.ID
}

// scenarioMatch scores how precisely a verse matches a selected tag.
//
// 1.0  the verse carries the tag itself
// 0.8  the verse carries a *deeper* tag under it (the user picked a broad
//
//	situation, the verse names the specific one)
//
// 0.6  the verse carries a shallower ancestor of it (the user picked a
//
//	specific situation, the verse stays broad)
func scenarioMatch(v model.Verse, tag string) float64 {
	best := 0.0
	for _, tags := range v.Annotation.Tags {
		for _, t := range tags {
			switch {
			case t == tag:
				return 1.0
			case strings.HasPrefix(t, tag+"."):
				// deeper: the verse is more specific than the selection
				if 0.8 > best {
					best = 0.8
				}
			case strings.HasPrefix(tag, t+"."):
				// shallower: the verse is broader than the selection
				if 0.6 > best {
					best = 0.6
				}
			}
		}
	}
	return best
}
