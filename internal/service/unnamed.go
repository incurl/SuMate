package service

import (
	"sort"

	"github.com/incurl/sumate/internal/model"
)

// Unnamed returns the verses that no concept reached.
//
// These are the edge of the taxonomy.  They cannot appear in the tag tree --
// a verse with no concept has no tags -- so without a shelf of their own they
// would be in the corpus but unreachable from the interface.  Giving them a
// place turns the blind spot into something a reader can walk into.
//
// The groups are honest about why each verse is here.  Most are simply too
// short to carry an image; a few contain a glyph that occurs nowhere else in
// the corpus and may be corrupted; the rest are complete lines the lexicon
// has no concept for.
func (s *Service) Unnamed() model.UnnamedShelf {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// How often each character occurs across the corpus.  A character seen
	// exactly once is a hint that the glyph is mangled rather than rare.
	freq := map[rune]int{}
	for _, v := range s.verses {
		for _, r := range v.Text {
			if r >= 0x4e00 && r <= 0x9fff {
				freq[r]++
			}
		}
	}

	byReason := map[string][]model.UnnamedVerse{}
	total := 0
	for _, v := range s.verses {
		if len(v.Annotation.Concepts) != 0 {
			continue
		}
		total++

		n := 0
		rare := false
		for _, r := range v.Text {
			if r < 0x4e00 || r > 0x9fff {
				continue
			}
			n++
			if freq[r] == 1 {
				rare = true
			}
		}

		// Classify by length -- a fact -- and carry the glyph anomaly as a
		// hint rather than as a category, because a rare character is not
		// evidence of corruption on its own: 鹅, 禹 and 尹 are perfectly
		// legible.  Only where a rare glyph makes the line unreadable does
		// the hint matter, and the reader can judge that.
		reason, note := model.UnnamedNoConcept, ""
		switch {
		case n < 5:
			reason = model.UnnamedTooShort
			note = "太短，承载不了意象"
		case n < 8:
			reason = model.UnnamedShort
			note = "偏短，意象尚未成形"
		default:
			reason = model.UnnamedNoConcept
			note = "长度如常，只是词库里没有与它对应的概念"
		}
		byReason[reason] = append(byReason[reason], model.UnnamedVerse{
			Verse:        v,
			Reason:       reason,
			Note:         note,
			SuspectGlyph: rare,
		})
	}

	shelf := model.UnnamedShelf{
		Total:  total,
		Corpus: len(s.verses),
		Title:  "无名之句",
		Preamble: "这 " + itoa(total) + " 句没有被任何概念命中，" +
			"因此进不了五个维度，也从来不会出现在匹配结果里。" +
			"它们不是坏的——多数只是太短，承载不了意象；" +
			"少数可能带着传本留下的讹字；还有十几句是完整的，" +
			"只是词库里没有与它对应的概念。",
	}

	// The groups run from the most readable to the least: a reader arriving
	// here wants the complete lines first.
	order := []struct{ reason, label, note string }{
		{model.UnnamedNoConcept, "长度如常，而未有名",
			"这些句子和别的句子一样长，只是词库里没有与它对应的概念。" +
				"这是标注的边界，不是文本的边界。"},
		{model.UnnamedShort, "偏短，意象未成形",
			"五到七字。在词里这是正常的句长，但短到还构不成一个可以命名的意象。"},
		{model.UnnamedTooShort, "太短，无以名之",
			"这些是真实的词句——《如梦令》的「但洗。但洗。」就在其中——" +
				"只是短到没有可以命名的东西。"},
	}
	for _, o := range order {
		vs := byReason[o.reason]
		if len(vs) == 0 {
			continue
		}
		// Longer verses first within a group: the complete ones are the
		// ones a reader is most likely to want to read.
		sort.SliceStable(vs, func(i, j int) bool {
			return cjkLen(vs[i].Verse.Text) > cjkLen(vs[j].Verse.Text)
		})
		shelf.Groups = append(shelf.Groups, model.UnnamedGroup{
			Reason: o.reason,
			Label:  o.label,
			Note:   o.note,
			Count:  len(vs),
			Verses: vs,
		})
	}
	return shelf
}

// cjkLen counts the Chinese characters of a string.
func cjkLen(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			n++
		}
	}
	return n
}

// itoa renders an int without importing strconv into this file.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
