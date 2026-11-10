package corpus

import (
	"strings"
	"testing"

	"github.com/sumate/sumate/internal/annotate"
)

func TestPipelineProducesAnnotatedVerses(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	if len(c.Poems) < 300 {
		t.Fatalf("expected the full corpus, got %d poems", len(c.Poems))
	}
	p := NewPipeline(c, annotate.New())
	verses, err := p.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(verses) < 1500 {
		t.Fatalf("expected a substantial verse set, got %d", len(verses))
	}

	lines, couplets := 0, 0
	tagged := 0
	for _, v := range verses {
		switch v.Kind {
		case "line":
			lines++
		case "couplet":
			couplets++
		}
		if len(v.Annotation.Concepts) > 0 {
			tagged++
		}
	}
	t.Logf("verses=%d lines=%d couplets=%d conceptTagged=%d",
		len(verses), lines, couplets, tagged)
	if tagged*2 < len(verses) {
		t.Errorf("only %d of %d verses carry concepts", tagged, len(verses))
	}
}

func TestFamousVersesArePresentAndAnnotated(t *testing.T) {
	c, _ := Load()
	p := NewPipeline(c, annotate.New())
	verses, _ := p.Build()

	want := []string{
		"但愿人长久",
		"何似在人间",
		"此事古难全",
		"大江东去",
		"莫听穿林打叶声",
		"也无风雨也无晴",
		"拣尽寒枝不肯栖",
		"谁道人生无再少",
		"十年生死两茫茫",
		"此心安处是吾乡",
	}
	for _, frag := range want {
		found := false
		for _, v := range verses {
			if strings.Contains(v.Text, frag) {
				found = true
				if len(v.Annotation.Concepts) == 0 {
					t.Errorf("verse %q has no concepts", frag)
				}
				if v.Annotation.Literal == "" || v.Annotation.Conception == "" {
					t.Errorf("verse %q has an empty annotation", frag)
				}
				break
			}
		}
		if !found {
			t.Errorf("famous verse %q not found in the verse set", frag)
		}
	}
}

func TestHistoryIsDerivedForDatedPoems(t *testing.T) {
	c, _ := Load()
	p := NewPipeline(c, annotate.New())
	verses, _ := p.Build()

	var dated int
	for _, v := range verses {
		if v.Background.Era != "" {
			dated++
		}
		if v.Background.Historical == "" {
			t.Fatalf("verse %s has no historical background", v.ID)
		}
	}
	t.Logf("verses with an explicit era: %d/%d", dated, len(verses))
	if dated == 0 {
		t.Error("no verse carries an explicit era")
	}
}
