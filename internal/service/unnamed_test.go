package service

import (
	"fmt"
	"testing"
)

func TestUnnamedShelf(t *testing.T) {
	s := newService(t)
	shelf := s.Unnamed()

	if shelf.Total == 0 {
		t.Fatal("the shelf is empty; every verse matched a concept")
	}
	if shelf.Corpus < shelf.Total {
		t.Errorf("the shelf (%d) is larger than the corpus (%d)", shelf.Total, shelf.Corpus)
	}
	if shelf.Title == "" || shelf.Preamble == "" {
		t.Error("the shelf has no title or preamble")
	}

	// Every verse in the shelf must actually be concept-less, and the group
	// counts must add up.
	sum := 0
	for _, g := range shelf.Groups {
		if g.Count != len(g.Verses) {
			t.Errorf("group %q says %d but holds %d", g.Reason, g.Count, len(g.Verses))
		}
		sum += g.Count
		for _, uv := range g.Verses {
			if len(uv.Verse.Annotation.Concepts) != 0 {
				t.Errorf("verse %q is on the shelf but has concepts", uv.Verse.Text)
			}
			if uv.Note == "" {
				t.Errorf("verse %q has no note", uv.Verse.Text)
			}
			if uv.Reason != g.Reason {
				t.Errorf("verse %q is grouped as %q but labelled %q",
					uv.Verse.Text, g.Reason, uv.Reason)
			}
		}
	}
	if sum != shelf.Total {
		t.Errorf("group counts sum to %d, shelf says %d", sum, shelf.Total)
	}

	fmt.Printf("\n《%s》 %d / %d 句\n", shelf.Title, shelf.Total, shelf.Corpus)
	fmt.Printf("%s\n\n", shelf.Preamble)
	for _, g := range shelf.Groups {
		fmt.Printf("── %s（%d）──\n  %s\n", g.Label, g.Count, g.Note)
		for i, uv := range g.Verses {
			if i >= 4 {
				fmt.Printf("  … 另 %d 句\n", len(g.Verses)-4)
				break
			}
			fmt.Printf("    %s\n", uv.Verse.Text)
		}
		fmt.Println()
	}
}

// TestUnnamedVersesAreOtherwiseUnreachable checks the reason the shelf exists.
func TestUnnamedVersesAreOtherwiseUnreachable(t *testing.T) {
	s := newService(t)
	shelf := s.Unnamed()

	// A verse on the shelf must carry no tags, so the tree can never show it.
	checked := 0
	for _, g := range shelf.Groups {
		for _, uv := range g.Verses {
			for _, tags := range uv.Verse.Annotation.Tags {
				if len(tags) > 0 {
					t.Errorf("verse %q is on the shelf but carries tags: %v",
						uv.Verse.Text, tags)
				}
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no verses were checked")
	}
}
