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

	"github.com/sumate/sumate/internal/curated"
	"github.com/sumate/sumate/internal/model"
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

// MatchedConcepts returns the concepts a verse text triggers, ordered by
// relevance (number of distinct keywords hit, then keyword length).
func (e *Engine) MatchedConcepts(text string) []string {
	idx := e.keywordIndex()
	score := map[string]int{}
	for kw, keys := range idx {
		if kw == "" {
			continue
		}
		n := strings.Count(text, kw)
		if n == 0 {
			continue
		}
		// Longer keywords are more specific and count for more.
		w := len([]rune(kw))
		for _, k := range keys {
			score[k] += n * w * w
		}
	}
	out := make([]string, 0, len(score))
	for k, s := range score {
		if s > 0 {
			out = append(out, k)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if score[out[i]] != score[out[j]] {
			return score[out[i]] > score[out[j]]
		}
		return out[i] < out[j]
	})
	return out
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
	if len(a.Concepts) == 0 {
		a.Concepts = e.MatchedConcepts(v.Text)
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
	dominant := dominantConcepts(a.Concepts, 3)
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
