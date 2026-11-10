// Package annotate turns the raw Su Dongpo corpus into fully annotated
// minimal matching units: single classic lines and symmetrical couplets.
//
// The engine combines three sources of truth:
//
//   - the concept lexicon of package curated, which maps classical imagery to
//     machine-readable semantic units;
//   - the historical metadata parsed from the 词题/小序 of each poem;
//   - a set of hand-written overrides for the verses that carry the most
//     weight in the product.
package annotate

import (
	"regexp"
	"sort"
	"strings"

	"github.com/incurl/sumate/internal/curated"
	"github.com/incurl/sumate/internal/model"
)

// Engine annotates the corpus.
type Engine struct {
	concepts  map[string]curated.Concept
	emotions  map[string]curated.EmotionDef
	overrides map[string]curated.Concept
}

// New returns an engine backed by the curated knowledge base.
func New() *Engine {
	return &Engine{
		concepts:  curated.Concepts(),
		emotions:  curated.Emotions(),
		overrides: curated.Overrides(),
	}
}

// Concepts exposes the concept lexicon.
func (e *Engine) Concepts() map[string]curated.Concept { return e.concepts }

// keywordIndex maps a keyword to the concept keys that claim it.
func (e *Engine) keywordIndex() map[string][]string {
	idx := map[string][]string{}
	for k, c := range e.concepts {
		for _, kw := range c.Keywords {
			idx[kw] = append(idx[kw], k)
		}
	}
	return idx
}

// The relative weight of a single-character keyword against a specific
// multi-character phrase.
//
// A specific phrase such as 「拣尽寒枝」 scores 16 (4 characters, squared);
// a single character such as 「风」 scores 1.  This is the equivalent of
// oblique's long-word-first rule: the specific always wins, but the
// ordinary word still contributes evidence.
const (
	weakWeight = 1
	weakScore  = 1
)

// MatchedConcepts returns the concepts a verse text triggers, ordered by
// relevance.
//
// The scoring distinguishes two kinds of keyword:
//
//   - a **specific** keyword (two characters or more, e.g. 「婵娟」,
//     「拣尽寒枝」) is strong evidence on its own;
//   - a **common** keyword (a single character such as 「风」「月」「无」)
//     is only weak evidence, because it occurs in a large fraction of the
//     corpus and would otherwise attach a concept to almost every verse.
//
// A concept backed only by single characters is still recorded -- dropping
// it would lose verses that really are about the topic -- but it is ranked
// below every concept that a specific phrase matched, and it is kept out
// of the *primary* concept list.
//
// This is the equivalent of oblique's Y1 rule: a word that is also an
// ordinary word must not be trusted on its own, but it is still evidence.
func (e *Engine) MatchedConcepts(text string) []string {
	ranked, _ := e.matchConcepts(text)
	return ranked
}

// matchConcepts returns the ranked concepts and the subset that was backed
// by a specific (multi-character) keyword.
func (e *Engine) matchConcepts(text string) (ranked []string, specific []string) {
	idx := e.keywordIndex()
	score := map[string]int{}
	// strong records, per concept, whether a multi-character keyword hit.
	strong := map[string]bool{}

	for kw, keys := range idx {
		if kw == "" {
			continue
		}
		n := strings.Count(text, kw)
		if n == 0 {
			continue
		}
		w := len([]rune(kw))
		for _, k := range keys {
			if w >= 2 {
				// Longer keywords are more specific and count for more.
				score[k] += n * w * w
				strong[k] = true
			} else {
				// A single character is a weak hint: it is accumulated
				// separately so that it can never outweigh a specific
				// phrase.
				score[k] += n * weakWeight
			}
		}
	}

	ranked = make([]string, 0, len(score))
	for k, s := range score {
		if s <= 0 {
			continue
		}
		if !strong[k] {
			score[k] = weakScore
		}
		ranked = append(ranked, k)
	}
	sort.Slice(ranked, func(i, j int) bool {
		// Concepts with specific evidence always come first.
		if strong[ranked[i]] != strong[ranked[j]] {
			return strong[ranked[i]]
		}
		if score[ranked[i]] != score[ranked[j]] {
			return score[ranked[i]] > score[ranked[j]]
		}
		return ranked[i] < ranked[j]
	})

	for _, k := range ranked {
		if strong[k] {
			specific = append(specific, k)
		}
	}
	return ranked, specific
}

// Annotate produces the full annotation of one verse.
func (e *Engine) Annotate(v *model.Verse, poem model.Poem) model.Annotation {
	a := model.Annotation{
		Polarity: model.Neutral,
		Concepts: e.MatchedConcepts(v.Text),
	}

	// --- override first: hand-written knowledge wins.
	if ov, ok := e.overrides[v.ID]; ok {
		a = applyOverride(a, ov)
	}

	// --- derive from concepts.
	//
	// The primary concept list (backed by a specific phrase) drives the
	// prose; the full list contributes the tags, so that a verse whose only
	// evidence is a common word still shows up under the right dimension.
	ranked, specific := e.matchConcepts(v.Text)
	a.Concepts = ranked
	if len(specific) > 0 {
		a.Primary = specific
	} else {
		a.Primary = ranked
	}

	var (
		emotionScore = map[string]int{}
		tangentSet   = map[string]bool{}
		tagSet       = map[string]bool{}
		polPos       int
		polNeu       int
		polMel       int
	)
	for _, ck := range a.Concepts {
		c := e.concepts[ck]
		for _, em := range c.Emotions {
			emotionScore[em]++
		}
		for _, tg := range c.Tangents {
			tangentSet[tg] = true
		}
		for _, tag := range c.Tags {
			tagSet[tag] = true
		}
		switch c.Polarity {
		case "positive":
			polPos++
		case "neutral":
			polNeu++
		case "melanchol":
			polMel++
		case "mixed":
			polPos++
			polMel++
		}
	}

	// --- emotions with graded strength.
	emotions := make([]model.Emotion, 0, len(emotionScore))
	for key, hits := range emotionScore {
		def, ok := e.emotions[key]
		if !ok {
			continue
		}
		emotions = append(emotions, model.Emotion{
			ID:        def.Key,
			Name:      def.Name,
			Dimension: def.Dimension,
			Strength:  gradeStrength(hits),
		})
	}
	sort.Slice(emotions, func(i, j int) bool {
		if emotions[i].Strength.Rank() != emotions[j].Strength.Rank() {
			return emotions[i].Strength.Rank() > emotions[j].Strength.Rank()
		}
		return emotions[i].ID < emotions[j].ID
	})
	a.Emotions = emotions

	// --- tangents, in canonical order.
	for _, t := range model.TangentOrder {
		if tangentSet[string(t)] {
			a.Tangents = append(a.Tangents, t)
		}
	}

	// --- polarity.
	if len(a.Concepts) > 0 {
		switch {
		case polMel > polPos && polMel > polNeu:
			a.Polarity = model.Melanchol
		case polPos > polNeu && polPos > polMel:
			a.Polarity = model.Positive
		case polNeu >= polPos && polNeu >= polMel:
			a.Polarity = model.Neutral
		default:
			a.Polarity = model.Mixed
		}
	}

	// --- prose fields: assemble from the dominant concepts.
	dominant := dominantConcepts(a.Primary, 3)
	if len(dominant) > 0 {
		c0 := e.concepts[dominant[0]]
		a.Literal = c0.Surface
		a.Conception = c0.Deep
		a.Philosophy = c0.Philosophy
		a.Attitude = c0.Attitude
		a.Crowd = c0.Name + "之境"
		a.MoodThreshold = "情绪强度中等以上"
		a.ScenarioBoundary = "与「" + c0.Name + "」相关的生活场景"
		a.Healing = "以「" + c0.Name + "」的通透化解，含蓄而不说破"
		a.Keywords = c0.Keywords
		for _, ck := range dominant[1:] {
			c := e.concepts[ck]
			a.Conception += " " + c.Deep
			a.Philosophy += " " + c.Philosophy
			a.ScenarioBoundary += "；亦近「" + c.Name + "」"
		}
	} else {
		a.Literal = "句意平淡，取其境而不取其辞。"
		a.Conception = "语淡而味长，寄情于景。"
		a.Philosophy = "不着一字，尽得风流。"
		a.Attitude = "从容自适。"
		a.Crowd = "日常心绪"
		a.MoodThreshold = "情绪强度不限"
		a.ScenarioBoundary = "泛用"
		a.Healing = "留白式慰藉"
	}

	// --- stances: the philosophical positions the verse takes.
	//
	// A stance is derived from the concepts the verse matched, so it is
	// always traceable back to a hand-written claim rather than inferred.
	a.Stances = stanceKeys(a.Concepts)

	// --- tags, grouped by root dimension.
	a.Tags = map[model.RootKey][]string{}
	for tag := range tagSet {
		root := rootOf(tag)
		if root == "" {
			continue
		}
		a.Tags[root] = append(a.Tags[root], tag)
	}
	for r := range a.Tags {
		sort.Strings(a.Tags[r])
	}

	return a
}

// applyOverride merges a hand-written concept into an annotation.
func applyOverride(a model.Annotation, ov curated.Concept) model.Annotation {
	if ov.Surface != "" {
		a.Literal = ov.Surface
	}
	if ov.Deep != "" {
		a.Conception = ov.Deep
	}
	if ov.Philosophy != "" {
		a.Philosophy = ov.Philosophy
	}
	if ov.Attitude != "" {
		a.Attitude = ov.Attitude
	}
	if ov.Polarity != "" {
		a.Polarity = model.Polarity(ov.Polarity)
	}
	if len(ov.Keywords) > 0 {
		a.Keywords = ov.Keywords
	}
	return a
}

// dominantConcepts returns the n most relevant concepts.
func dominantConcepts(concepts []string, n int) []string {
	if len(concepts) <= n {
		return concepts
	}
	return concepts[:n]
}

// gradeStrength maps a hit count to a graded strength.
func gradeStrength(hits int) model.Strength {
	switch {
	case hits >= 4:
		return model.Profound
	case hits == 3:
		return model.Obvious
	case hits == 2:
		return model.Moderate
	default:
		return model.Slight
	}
}

// rootOf returns the root dimension a tag belongs to.
func rootOf(tag string) model.RootKey {
	i := strings.Index(tag, ".")
	if i < 0 {
		return ""
	}
	return model.RootKey(tag[:i])
}

// cjk runs of the text
var cjkRe = regexp.MustCompile(`[\p{Han}]+`)

// splitSentences splits a punctuated verse into sentences.
func splitSentences(text string) []string {
	var out []string
	for _, s := range cjkRe.FindAllString(text, -1) {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// stanceKeys returns the philosophical stances carried by a set of concepts.
func stanceKeys(concepts []string) []string {
	idx := curated.StanceKeys()
	seen := map[string]bool{}
	var out []string
	for _, ck := range concepts {
		for _, st := range idx[ck] {
			if seen[st] {
				continue
			}
			seen[st] = true
			out = append(out, st)
		}
	}
	// Keep the canonical order so that the annotation is deterministic.
	order := map[string]int{}
	for i, k := range curated.StanceOrder {
		order[k] = i
	}
	sort.SliceStable(out, func(i, j int) bool {
		return order[out[i]] < order[out[j]]
	})
	return out
}
