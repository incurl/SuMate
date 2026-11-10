package service

import (
	"fmt"
	"testing"

	"github.com/sumate/sumate/internal/tagtree"
)

// TestUserScenario walks the scenario the product is designed for: a user
// who is lonely, at night, in the rain, missing someone far away.
func TestUserScenario(t *testing.T) {
	s := newService(t)
	roots := s.Roots()

	// 心境 -> 低回与疗愈 -> 孤独
	moodKids := s.Children(roots[0].ID, nil)
	var lowID string
	for _, k := range moodKids {
		if k.ID == "mood.low" {
			lowID = k.ID
		}
	}
	if lowID == "" {
		t.Fatal("no 低回与疗愈 node")
	}
	path := []tagtree.Filter{
		{NodeID: roots[0].ID, Root: roots[0].Root},
		{NodeID: lowID, Root: roots[0].Root},
	}
	lonelyKids := s.Children(lowID, path)
	for _, k := range lonelyKids {
		if k.ID == "mood.low.lonely" {
			path = append(path, tagtree.Filter{NodeID: k.ID, Root: k.Root})
		}
	}

	// 时令 -> 风雨 -> 风雨骤至
	weatherKids := s.Children(roots[2].ID, path)
	var rainID string
	for _, k := range weatherKids {
		if k.ID == "weather.rain" {
			rainID = k.ID
		}
	}
	if rainID != "" {
		path = append(path, tagtree.Filter{NodeID: roots[2].ID, Root: roots[2].Root},
			tagtree.Filter{NodeID: rainID, Root: roots[2].Root})
		stormKids := s.Children(rainID, path)
		for _, k := range stormKids {
			if k.ID == "weather.rain.storm" {
				path = append(path, tagtree.Filter{NodeID: k.ID, Root: k.Root})
			}
		}
	}

	// 人际 -> 情爱 -> 相思
	relKids := s.Children(roots[3].ID, path)
	var loveID string
	for _, k := range relKids {
		if k.ID == "relation.romance" {
			loveID = k.ID
		}
	}
	if loveID != "" {
		path = append(path, tagtree.Filter{NodeID: roots[3].ID, Root: roots[3].Root},
			tagtree.Filter{NodeID: loveID, Root: roots[3].Root})
	}

	fmt.Println("\n======== 场景：夜里下雨，孤单，想一个人 ========")
	for _, c := range s.Breadcrumb(path) {
		fmt.Printf("  %-50s %5d\n", c.Path[len(c.Path)-1], c.Remaining)
	}
	res := s.Match(path, "夜里下着雨，很想一个人")
	fmt.Printf("\n  【东坡嘴替】\n  %s\n", res.Primary.Text)
	fmt.Printf("  %s · %s\n", res.Primary.Tune, res.Primary.Background.Historical)
	fmt.Printf("  综合匹配 %d（意境 %d / 情绪 %d / 情境 %d / 气质 %d）\n",
		res.Explanation.Overall, res.Explanation.Semantic,
		res.Explanation.Emotional, res.Explanation.Scenario, res.Explanation.Temperament)
	fmt.Printf("  表层: %s\n", res.Primary.Annotation.Literal)
	fmt.Printf("  态度: %s\n", res.Primary.Annotation.Attitude)
	fmt.Println("  备选:")
	for _, a := range res.Alternatives {
		fmt.Printf("    · %s\n", a.Text)
	}
	fmt.Println()
}
