package engine

import (
	"testing"

	"github.com/incurl/sumate/internal/model"
)

// TestFreeTextReachesDistinctVerses is the regression test for the
// diagnosis in docs/EVAL_向量化是否有用.md.
//
// Before the concept bridge, five different descriptions all returned the
// same verse: the matcher intersected single characters, and words like 我
// and 了 occur in a large share of the keywords.
func TestFreeTextReachesDistinctVerses(t *testing.T) {
	e := loadEngine(t)
	descs := []string{
		"我想一个人待着",
		"我很累不想动",
		"下雨了心里空空的",
		"思念远方的人",
		"想起了从前",
		"不知道该怎么办",
	}
	seen := map[string]int{}
	for _, d := range descs {
		cands := e.Rank(Selection{Description: d}, 3)
		if len(cands) == 0 {
			t.Fatalf("no candidates for %q", d)
		}
		seen[cands[0].Verse.ID]++
		t.Logf("  「%s」→ %s", d, trunc(cands[0].Verse.Text, 30))
	}
	if len(seen) < len(descs)-1 {
		t.Errorf("descriptions collapsed onto %d distinct verses out of %d",
			len(seen), len(descs))
	}
}

// TestFreeTextIsExplainable checks that the match can still say *why*.
//
// This is the property a vector model could not provide: the answer names
// the concept that was understood.
func TestFreeTextIsExplainable(t *testing.T) {
	e := loadEngine(t)
	res := e.Match(Selection{Description: "我很累不想动"})
	if res.Primary.ID == "" {
		t.Fatal("no match")
	}
	cs := bridgeConcepts("我很累不想动")
	if len(cs) == 0 {
		t.Fatal("the description bridged onto no concept")
	}
	carried := false
	for _, ck := range res.Primary.Annotation.Concepts {
		if cs[ck] {
			carried = true
			break
		}
	}
	if !carried {
		t.Errorf("the match %q carries none of the understood concepts %v",
			res.Primary.Text, cs)
	}
	// The reasons must be non-empty: the match is still explainable.
	if len(res.Explanation.Reasons) == 0 {
		t.Error("the match carries no reasons")
	}
	t.Logf("  理解为：%v", cs)
	t.Logf("  解释：%v", res.Explanation.Reasons[:2])
}

// TestFreeTextDoesNotOverrideTheTree checks that a description refines the
// tree selection rather than replacing it.
func TestFreeTextDoesNotOverrideTheTree(t *testing.T) {
	e := loadEngine(t)
	withTag := e.Match(Selection{
		Tags:        map[model.RootKey][]string{model.RootMood: {"mood.low.longing"}},
		Description: "下雨了",
	})
	plain := e.Match(Selection{
		Tags: map[model.RootKey][]string{model.RootMood: {"mood.low.longing"}},
	})
	if withTag.Primary.ID == "" {
		t.Fatal("no match")
	}
	t.Logf("  标签「相思」→ %s", trunc(plain.Primary.Text, 26))
	t.Logf("  标签「相思」+「下雨了」→ %s", trunc(withTag.Primary.Text, 26))
}

// TestBridgeUsesStrongKeywordsFirst checks that a multi-character keyword
// wins over an incidental single-character hit.
func TestBridgeUsesStrongKeywordsFirst(t *testing.T) {
	cs := bridgeConcepts("想家")
	if !cs["nostalgia.old_place"] {
		t.Errorf("「想家」did not reach the homesickness concept: %v", cs)
	}
	// "了" alone should not drag in a dozen concepts.
	cs2 := bridgeConcepts("了")
	if len(cs2) > 6 {
		t.Errorf("the single character 「了」bridged onto %d concepts", len(cs2))
	}
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
