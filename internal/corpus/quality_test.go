package corpus

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/sumate/sumate/internal/annotate"
)

// TestCorpusCopiesAreIdentical guards the single-source-of-truth rule.
//
// data/dongpo_corpus.json is the source of truth; the copy under
// internal/corpus/data/ exists only because go:embed cannot reach a parent
// directory.  If the two ever drift, the embedded copy silently wins at
// build time and the corpus that gets validated is not the corpus that
// ships.
func TestCorpusCopiesAreIdentical(t *testing.T) {
	canonical, err := os.ReadFile(filepath.Join("..", "..", "data", "dongpo_corpus.json"))
	if err != nil {
		t.Fatalf("read canonical corpus: %v", err)
	}
	embedded, err := os.ReadFile(filepath.Join("data", "dongpo_corpus.json"))
	if err != nil {
		t.Fatalf("read embedded corpus: %v", err)
	}
	if !bytes.Equal(canonical, embedded) {
		t.Error("the embedded corpus differs from data/dongpo_corpus.json; run tools/sync_corpus.sh")
	}
}

// TestCorpusDomainInvariants is the Go side of the corpus validator.
//
// It encodes the same ci-specific constraints as tools/lint_corpus.py, so
// that a change to the corpus builder cannot silently break them.
//
// The constraints are the ci analogue of the 楹联等长 check: a 词牌 has a
// fixed character count, and a poem that deviates from it is suspicious.
func TestCorpusDomainInvariants(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}

	// --- character count per tune
	byTune := map[string][]int{}
	texts := map[string][]string{}
	for _, p := range c.Poems {
		if !p.HasText {
			continue
		}
		n := cjkRuneCount(p.Text)
		byTune[p.Tune] = append(byTune[p.Tune], n)
		texts[p.Tune] = append(texts[p.Tune], p.Text)
	}

	// known short forms, verified against the source
	shortForms := map[string]bool{
		"江城子|腻红匀脸衬檀唇": true, // 四库 prints 【原缺】 inside the verse
		"风回仙|开扇":      true, // 风回仙 is a short 词牌 form
	}

	tunes := make([]string, 0, len(byTune))
	for t := range byTune {
		tunes = append(tunes, t)
	}
	sort.Strings(tunes)

	for _, tune := range tunes {
		counts := byTune[tune]
		if len(counts) < 2 {
			continue
		}
		mode := modeOf(counts)
		for i, n := range counts {
			if abs(n-mode) <= 4 {
				continue
			}
			if knownShort(tune, texts[tune][i], shortForms) {
				continue
			}
			t.Errorf("词牌「%s」usually %d characters, found %d: %s",
				tune, mode, n, texts[tune][i])
		}
	}
}

// TestNoEditorialApparatusInVerses checks that the 四库 apparatus never
// ends up inside a verse.
func TestNoEditorialApparatusInVerses(t *testing.T) {
	c, _ := Load()
	for _, p := range c.Poems {
		if !p.HasText {
			continue
		}
		if strings.Contains(p.Text, "【") {
			t.Errorf("verse %s (%s) still carries an editorial bracket: %s",
				p.ID, p.Tune, p.Text)
		}
		if strings.Contains(p.Text, "】") {
			t.Errorf("verse %s (%s) still carries an editorial bracket", p.ID, p.Tune)
		}
	}
}

// TestEverySentenceIsTerminated checks the 句读归属 constraint.
func TestEverySentenceIsTerminated(t *testing.T) {
	c, _ := Load()
	for _, p := range c.Poems {
		if !p.HasText {
			continue
		}
		text := strings.TrimSuffix(p.Text, "东坡词")
		text = strings.TrimSuffix(text, "吷")
		if !strings.HasSuffix(text, "。") &&
			!strings.HasSuffix(text, "？") &&
			!strings.HasSuffix(text, "！") {
			t.Errorf("verse %s (%s) does not end with a terminator: ...%s",
				p.ID, p.Tune, lastN(p.Text, 12))
		}
	}
}

// TestHasTextAgreesWithText catches the flag bug: an entry must not claim
// to have text when its cleaned verse is empty.
func TestHasTextAgreesWithText(t *testing.T) {
	c, _ := Load()
	for _, p := range c.Poems {
		if p.HasText && strings.TrimSpace(p.Text) == "" {
			t.Errorf("verse %s (%s) claims text but has none", p.ID, p.Tune)
		}
	}
}

// TestEveryPoemIsRenderable makes sure the corpus survives the whole
// pipeline with a complete annotation.
func TestEveryPoemIsRenderable(t *testing.T) {
	c, _ := Load()
	p := NewPipeline(c, annotate.New())
	verses, err := p.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, v := range verses {
		if v.Text == "" {
			continue
		}
		if v.Annotation.Literal == "" || v.Annotation.Attitude == "" {
			t.Errorf("verse %s has an incomplete annotation", v.ID)
		}
		if v.Background.Historical == "" {
			t.Errorf("verse %s has no background", v.ID)
		}
	}
}

// cjkRuneCount counts the Chinese characters of a string.
func cjkRuneCount(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			n++
		}
	}
	return n
}

// modeOf returns the most common value.
func modeOf(xs []int) int {
	counts := map[int]int{}
	best, bestN := 0, 0
	for _, x := range xs {
		counts[x]++
		if counts[x] > bestN {
			best, bestN = x, counts[x]
		}
	}
	return best
}

// abs returns the absolute value.
func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// knownShort reports whether the deviation is a documented short form.
func knownShort(tune, text string, known map[string]bool) bool {
	for key := range known {
		t, phrase, _ := strings.Cut(key, "|")
		if t == tune && strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

// lastN returns the last n characters of a string.
func lastN(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return fmt.Sprintf("%s", string(r[len(r)-n:]))
}
