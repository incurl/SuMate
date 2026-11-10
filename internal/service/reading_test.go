package service

import (
	"fmt"
	"testing"

	"github.com/incurl/sumate/internal/tagtree"
)

func TestPhilosophicalReading(t *testing.T) {
	s := newService(t)
	roots := s.Roots()

	// A reader who feels helpless: the "there is nothing more to do"
	// stance should hold; the "keep going" stance should be strained.
	moodKids := s.Children(roots[0].ID, nil)
	var low, helpless string
	for _, k := range moodKids {
		if k.ID == "mood.low" {
			low = k.ID
		}
	}
	path := []tagtree.Filter{{NodeID: roots[0].ID, Root: roots[0].Root}}
	if low != "" {
		path = append(path, tagtree.Filter{NodeID: low, Root: roots[0].Root})
		l3 := s.Children(low, path)
		for _, k := range l3 {
			if k.ID == "mood.low.helpless" {
				helpless = k.ID
			}
		}
	}
	if helpless != "" {
		path = append(path, tagtree.Filter{NodeID: helpless, Root: roots[0].Root})
	}

	fmt.Println("\n======== 哲学读解 ========")
	for _, a := range s.PhilosophicalReading(path) {
		fmt.Printf("  [%s] %s\n", a.Fit, a.Name)
		fmt.Printf("     主张：%s\n", a.Claim)
		fmt.Printf("     前提：%s\n", a.Premise)
		fmt.Printf("     若不成立：%s\n", a.Shadow)
		fmt.Printf("     %s\n", a.Note)
		fmt.Println()
	}
	if len(s.PhilosophicalReading(path)) == 0 {
		t.Fatal("no philosophical reading produced")
	}
}
