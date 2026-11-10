package corpus

import (
	"strings"
	"testing"
)

// TestFamousPoemsHaveAccurateHistory checks the historical annotation of the
// verses that carry the most weight in the product.
func TestFamousPoemsHaveAccurateHistory(t *testing.T) {
	c, _ := Load()
	p := NewPipeline(c, newTestEngine())
	verses, _ := p.Build()

	want := []struct {
		phrase string
		era    string
		place  string
	}{
		{"明月几时有", "熙宁", "山东密州"},
		{"莫听穿林打叶声", "元丰", "湖北黄冈"},
		{"大江东去浪淘尽", "元丰", "湖北黄冈"},
		{"十年生死两茫茫", "熙宁", "山东密州"},
		{"谁道人生无再少", "元丰", "湖北黄冈"},
		{"拣尽寒枝不肯栖", "绍圣", "广东惠州"},
		{"世事一场大梦", "元丰", "湖北黄冈"},
	}

	for _, w := range want {
		var bg = struct{ era, place string }{}
		found := false
		for _, v := range verses {
			if strings.Contains(v.Text, w.phrase) {
				bg.era = string(v.Background.Era)
				bg.place = v.Background.Place
				found = true
				break
			}
		}
		if !found {
			t.Errorf("verse %q not found", w.phrase)
			continue
		}
		if bg.era != w.era {
			t.Errorf("verse %q: era = %q, want %q", w.phrase, bg.era, w.era)
		}
		if bg.place != w.place {
			t.Errorf("verse %q: place = %q, want %q", w.phrase, bg.place, w.place)
		}
	}
}

// TestEveryVerseHasBackground makes sure no verse is left without a note.
func TestEveryVerseHasBackground(t *testing.T) {
	c, _ := Load()
	p := NewPipeline(c, newTestEngine())
	verses, _ := p.Build()
	for _, v := range verses {
		if v.Background.Historical == "" {
			t.Fatalf("verse %s (%s) has no historical note", v.ID, v.Text)
		}
		if v.Background.Stage == "" {
			t.Fatalf("verse %s (%s) has no life stage", v.ID, v.Text)
		}
	}
}

// TestGanzhiCycle checks the sexagenary-cycle computation against known
// anchors.
func TestGanzhiCycle(t *testing.T) {
	cases := map[int]string{
		1064: "甲辰",
		1076: "丙辰",
		1084: "甲子",
		1086: "丙寅",
		1094: "甲戌",
		1099: "己卯",
		1101: "辛巳",
	}
	for year, want := range cases {
		if got := ganzhiOf(year); got != want {
			t.Errorf("ganzhiOf(%d) = %q, want %q", year, got, want)
		}
	}
}
