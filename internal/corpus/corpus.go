// Package corpus loads the Su Dongpo corpus and splits every poem into
// independent minimal matching units: single classic lines and symmetrical
// couplets.
package corpus

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/incurl/sumate/internal/model"
)

//go:embed data/dongpo_corpus.json
var rawJSON []byte

// rawCorpus is the on-disk shape of data/dongpo_corpus.json.
type rawCorpus struct {
	Source map[string]string `json:"source"`
	Stats  struct {
		Poems     int `json:"poems"`
		Aligned   int `json:"aligned"`
		Unaligned int `json:"unaligned"`
		NoText    int `json:"no_text"`
	} `json:"stats"`
	Poems []rawPoem `json:"poems"`
}

// rawPoem is the on-disk shape of one poem.
//
// The layer fields record what the 四库 transmission contains beyond the
// verse itself, so that the provenance of a verse stays visible: how many
// characters are missing from the text, and which poets a 集句 quotes.
type rawPoem struct {
	Tune      string   `json:"tune"`
	Subtitle  string   `json:"subtitle"`
	Body      string   `json:"body"`
	Stanzas   []string `json:"stanzas"`
	Text      string   `json:"text"`
	Sentences []string `json:"sentences"`
	Align     float64  `json:"align"`
	HasText   bool     `json:"has_text"`
	Kind      string   `json:"kind"`
	Lacunae   int      `json:"lacunae"`
	Quotes    []string `json:"quotes"`
}

// Corpus is the loaded corpus.
type Corpus struct {
	Poems  []model.Poem
	Source map[string]string
}

// poemID builds the stable id of a poem.
func poemID(tune string, order int) string {
	return fmt.Sprintf("%s-%04d", tuneKey(tune), order)
}

// tuneKey turns a 词牌 into an ASCII-safe id fragment.
func tuneKey(tune string) string {
	var b strings.Builder
	for _, r := range tune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		case r > 127:
			// keep a stable numeric code point instead of the glyph
			fmt.Fprintf(&b, "u%x", r)
		default:
			b.WriteByte('-')
		}
	}
	return b.String()
}

// Load reads and parses the embedded corpus.
func Load() (*Corpus, error) {
	var rc rawCorpus
	if err := json.Unmarshal(rawJSON, &rc); err != nil {
		return nil, fmt.Errorf("parse corpus: %w", err)
	}
	c := &Corpus{Source: rc.Source}
	seen := map[string]int{}
	for i, p := range rc.Poems {
		base := poemID(p.Tune, i)
		seen[base]++
		id := base
		if seen[base] > 1 {
			id = fmt.Sprintf("%s-%d", base, seen[base])
		}
		// The kind is recorded by the corpus builder, which saw the 四库
		// apparatus before it was stripped; deriving it here from the 词题
		// would misclassify a 词题 that merely mentions 集句.
		kind := p.Kind
		if kind == "" {
			kind = "normal"
		}
		c.Poems = append(c.Poems, model.Poem{
			ID:      id,
			Tune:    p.Tune,
			Title:   p.Subtitle,
			Text:    p.Text,
			Stanzas: p.Stanzas,
			Sents:   p.Sentences,
			Kind:    kind,
			Lacunae: p.Lacunae,
			Quotes:  p.Quotes,
			HasText: p.HasText,
			Order:   i,
		})
	}
	return c, nil
}

// sentenceRe splits a punctuated verse into sentence units.
var sentenceRe = regexp.MustCompile(`[^。！？!?]+[。！？!?]?`)

// SplitUnits decomposes a poem into its minimal matching units.
//
// A 词 is built from 句 (sentences); a 句 may be a single line or a pair of
// symmetrical clauses.  The unit list therefore contains both the individual
// clauses and the couplets that pair them, so that the matcher can offer
// either granularity.
func SplitUnits(p model.Poem) []Unit {
	var units []Unit
	order := 0
	clauses := splitClauses(p.Sents)
	for i, c := range clauses {
		units = append(units, Unit{
			Order:   order,
			Kind:    "line",
			Text:    c,
			Stanza:  stanzaOf(p, c),
			ClauseA: i,
		})
		order++
	}
	// symmetrical couplets: pair consecutive clauses of similar length
	for i := 0; i+1 < len(clauses); i += 2 {
		a, b := clauses[i], clauses[i+1]
		if !isSymmetric(a, b) {
			continue
		}
		units = append(units, Unit{
			Order:   order,
			Kind:    "couplet",
			Text:    a + b,
			Stanza:  stanzaOf(p, a),
			ClauseA: i,
		})
		order++
	}
	return units
}

// Unit is one minimal matching unit of a poem.
type Unit struct {
	Order   int
	Kind    string
	Text    string
	Stanza  int
	ClauseA int
}

// splitClauses breaks the sentences of a poem into clauses.
func splitClauses(sents []string) []string {
	var out []string
	for _, s := range sents {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

// isSymmetric reports whether two clauses look like a balanced couplet.
func isSymmetric(a, b string) bool {
	la, lb := runeLen(a), runeLen(b)
	if la < 4 || lb < 4 {
		return false
	}
	diff := la - lb
	if diff < 0 {
		diff = -diff
	}
	// classical couplets are usually within one character of each other
	return diff <= 1
}

// runeLen counts the Chinese characters of a string.
func runeLen(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			n++
		}
	}
	return n
}

// stanzaOf returns the index of the 片 a clause belongs to.
func stanzaOf(p model.Poem, clause string) int {
	for i, st := range p.Stanzas {
		if strings.Contains(st, strings.TrimSuffix(clause, "。")) {
			return i
		}
	}
	return 0
}
