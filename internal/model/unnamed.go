package model

// UnnamedVerse is a verse that no concept reached.
//
// Every verse in the corpus is *retrievable*, but a verse is only
// *interpretable* if a concept matched it.  The ones that matched nothing
// cannot appear in the tag tree and therefore never surface in a normal
// drill-down.  They are not broken; they are the edge of the taxonomy.
type UnnamedVerse struct {
	// Verse is the verse itself.
	Verse Verse `json:"verse"`
	// Reason is the machine key for why nothing matched.
	Reason string `json:"reason"`
	// Note is the sentence shown to the reader.
	Note string `json:"note"`
	// SuspectGlyph is a *hint*, not a verdict: the verse contains a
	// character that occurs nowhere else in the corpus, so a corrupted
	// glyph is one possible explanation for the miss.
	SuspectGlyph bool `json:"suspect_glyph"`
}

// Unnamed reasons.
//
// The classification is by length, because length is a fact.  Whether a
// verse is "complete" is a judgement; whether it is five characters long is
// not.  Grouping by a defensible fact keeps the shelf honest.
const (
	// UnnamedTooShort: the 句 is too short to carry an image.  「但洗。」
	// is a real line of 如梦令, but there is nothing in it to name.
	UnnamedTooShort = "too_short"
	// UnnamedShort: the 句 is shorter than the corpus median but is not a
	// fragment.  The image has not yet formed at this length.
	UnnamedShort = "short"
	// UnnamedNoConcept: the 句 is of ordinary length, and the lexicon simply
	// has no concept for it.
	UnnamedNoConcept = "no_concept"
)

// UnnamedGroup is one shelf of unnamed verses, grouped by reason.
type UnnamedGroup struct {
	// Reason is the machine key.
	Reason string `json:"reason"`
	// Label is the Chinese display name.
	Label string `json:"label"`
	// Note is an honest description of what this group contains.
	Note string `json:"note"`
	// Count is the number of verses in the group.
	Count int `json:"count"`
	// Verses are the verses themselves.
	Verses []UnnamedVerse `json:"verses"`
}

// UnnamedShelf is the whole collection of unnamed verses.
type UnnamedShelf struct {
	// Total is the number of unnamed verses.
	Total int `json:"total"`
	// Corpus is the total number of verses in the corpus.
	Corpus int `json:"corpus"`
	// Title is the display title of the shelf.
	Title string `json:"title"`
	// Preamble states honestly what the shelf is.
	Preamble string `json:"preamble"`
	// Groups are the verses, grouped by reason.
	Groups []UnnamedGroup `json:"groups"`
}
