// Command audit recomputes the claims of docs/EVAL_形而上评估.md.
//
// A claim that cannot be recomputed is not a claim; this makes the
// evaluation's central numbers reproducible.
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/incurl/sumate/internal/annotate"
	"github.com/incurl/sumate/internal/corpus"
	"github.com/incurl/sumate/internal/engine"
	"github.com/incurl/sumate/internal/model"
	"github.com/incurl/sumate/internal/service"
	"github.com/incurl/sumate/internal/tagtree"
)

func main() {
	c, err := corpus.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	verses, err := corpus.NewPipeline(c, annotate.New()).Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	e := engine.New(verses)

	cases := []struct {
		label string
		tags  map[model.RootKey][]string
	}{
		{"选 mood.low.longing", map[model.RootKey][]string{
			model.RootMood: {"mood.low.longing"},
		}},
		{"选 mood.low（宽）", map[model.RootKey][]string{
			model.RootMood: {"mood.low"},
		}},
	}

	for _, cs := range cases {
		cands := e.Rank(engine.Selection{Tags: cs.tags}, 0)
		axes := map[string][]float64{
			"语义": {}, "情绪": {}, "情境": {}, "气质": {},
		}
		for _, cd := range cands {
			p := cd.Parts
			axes["语义"] = append(axes["语义"], float64(p.Semantic))
			axes["情绪"] = append(axes["情绪"], float64(p.Emotional))
			axes["情境"] = append(axes["情境"], float64(p.Scenario))
			axes["气质"] = append(axes["气质"], float64(p.Temperament))
		}

		fmt.Printf("\n=== %s ===\n", cs.label)
		fmt.Printf("  候选 %d 句\n", len(cands))
		for _, name := range []string{"语义", "情绪", "情境", "气质"} {
			xs := axes[name]
			distinct := map[float64]bool{}
			for _, x := range xs {
				distinct[x] = true
			}
			mean := mean(xs)
			sd := pstdev(xs)
			flag := ""
			if sd == 0 {
				flag = "  ← 常数"
			}
			fmt.Printf("  %s : 均值 %6.1f  标准差 %6.2f  取值 %d 种%s\n",
				name, mean, sd, len(distinct), flag)
		}

		// The disagreement between the concept table and the tag table.
		//
		// Only verses that actually *carry* the selected tag can
		// contradict it.  Counting all the candidates instead measured
		// something else: most of the corpus does not carry the tag at
		// all, so it was never a contradiction to begin with, and the
		// number was inflated accordingly.
		disagree, tagged := 0, 0
		for _, cd := range cands {
			if !carriesAny(cd.Verse, cs.tags[model.RootMood]) {
				continue
			}
			tagged++
			if cd.Parts.Emotional == 0 {
				disagree++
			}
		}
		if tagged > 0 {
			fmt.Printf("  带所选标签却情绪判 0：%d / %d 句 (%.0f%%)\n",
				disagree, tagged, 100*float64(disagree)/float64(tagged))
		}

		// The same measures over the *top* of the ranking, which is where
		// the decision is actually made.  An axis that discriminates over
		// the whole corpus but not over the winners is not deciding.
		top := cands
		if len(top) > 50 {
			top = top[:50]
		}
		fmt.Println("  --- 仅看前 50 名（决策真正发生的地方）---")
		taxes := map[string][]float64{"语义": {}, "情绪": {}, "情境": {}, "气质": {}}
		for _, cd := range top {
			p := cd.Parts
			taxes["语义"] = append(taxes["语义"], float64(p.Semantic))
			taxes["情绪"] = append(taxes["情绪"], float64(p.Emotional))
			taxes["情境"] = append(taxes["情境"], float64(p.Scenario))
			taxes["气质"] = append(taxes["气质"], float64(p.Temperament))
		}
		for _, name := range []string{"语义", "情绪", "情境", "气质"} {
			xs := taxes[name]
			distinct := map[float64]bool{}
			for _, x := range xs {
				distinct[x] = true
			}
			flag := ""
			if pstdev(xs) == 0 {
				flag = "  ← 常数"
			}
			fmt.Printf("  %s : 均值 %6.1f  标准差 %6.2f  取值 %d 种%s\n",
				name, mean(xs), pstdev(xs), len(distinct), flag)
		}
	}

	// The free-text channel.
	fmt.Println("\n=== 自由文本是否改变排序 ===")
	base := e.Rank(engine.Selection{}, 5)
	with := e.Rank(engine.Selection{Description: "我困了想睡"}, 5)
	same := 0
	for i := range base {
		if base[i].Verse.ID == with[i].Verse.ID {
			same++
		}
	}
	fmt.Printf("  前五中有 %d 句相同\n", same)

	// Order dependence: if the top of the ranking ties, the winner is
	// decided by corpus order rather than by fit.
	fmt.Println("\n=== 同分时答案是否随底本顺序改变 ===")
	broad := engine.Selection{Tags: map[model.RootKey][]string{
		model.RootMood: {"mood.low"},
	}}
	top30 := e.Rank(broad, 30)
	for i := 0; i < len(top30) && i < 5; i++ {
		fmt.Printf("  %2d. %-32s 总分 %d\n", i+1,
			trunc(top30[i].Verse.Text, 30), top30[i].Parts.Overall)
	}
	tied := 0
	for _, cd := range top30 {
		if cd.Parts.Overall == top30[0].Parts.Overall {
			tied++
		}
	}
	fmt.Printf("  前 %d 名中与第 1 名同分者：%d 句\n", len(top30), tied)

	rev := make([]model.Verse, 0, len(verses))
	for i := len(verses) - 1; i >= 0; i-- {
		rev = append(rev, verses[i])
	}
	revEngine := engine.New(rev)
	a := e.Match(broad)
	b := revEngine.Match(broad)
	fmt.Printf("  正序底本 → %s\n", trunc(a.Primary.Text, 30))
	fmt.Printf("  倒序底本 → %s\n", trunc(b.Primary.Text, 30))
	if a.Primary.ID == b.Primary.ID {
		fmt.Println("  答案不随底本顺序改变")
	} else {
		fmt.Println("  ⚠ 答案随底本顺序改变：同分时排序不稳定")
	}

	// Tree depth.
	fmt.Println("\n=== 树深对答案的影响 ===")
	fmt.Println("  （在 service 层测量：那里才真正按标签过滤语料）")
	svc, err := service.New()
	if err != nil {
		fmt.Printf("  service 不可用：%v\n", err)
		return
	}
	roots := svc.Roots()
	if len(roots) == 0 {
		return
	}
	mood := roots[0]
	path := []tagtree.Filter{{NodeID: mood.ID, Root: mood.Root}}
	report := func(label string) {
		res := svc.Match(path, "")
		top := ""
		if res.Primary.ID != "" {
			top = trunc(res.Primary.Text, 26)
		}
		fmt.Printf("  %-22s 候选 %5d → %s\n", label, res.Candidates, top)
	}
	report("根 " + mood.ID)
	kids := svc.Children(mood.ID, nil)
	if len(kids) == 0 {
		return
	}
	path = append(path, tagtree.Filter{NodeID: kids[0].ID, Root: mood.Root})
	report("二层 " + kids[0].ID)
	gk := svc.Children(kids[0].ID, nil)
	if len(gk) == 0 {
		return
	}
	path = append(path, tagtree.Filter{NodeID: gk[0].ID, Root: mood.Root})
	report("三层 " + gk[0].ID)
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func pstdev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := mean(xs)
	var v float64
	for _, x := range xs {
		v += (x - m) * (x - m)
	}
	return math.Sqrt(v / float64(len(xs)))
}

// carriesAny reports whether the verse carries at least one of the tags.
func carriesAny(v model.Verse, tags []string) bool {
	for _, want := range tags {
		for _, xs := range v.Annotation.Tags {
			for _, t := range xs {
				if t == want {
					return true
				}
			}
		}
	}
	return false
}
