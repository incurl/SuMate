// Package philosophy evaluates whether a verse's philosophical stance
// actually applies to the reader who is receiving it.
//
// The product's promise is that Su Dongpo speaks *for* the user.  That
// promise is only kept if the product knows when a verse would mislead.
//
// A verse like 「此心安处是吾乡」 is true and is consolation to a man who
// has been exiled and can find no other footing.  To a man who still has a
// road to travel, the same words are an excuse to stop walking.  The words
// do not change; the reader does.  So the product must reason about the
// reader's position, not only about the verse's meaning.
//
// This package therefore does two things:
//
//  1. it reads the philosophical stance off a verse (which stance it takes,
//     and what that stance presupposes);
//  2. it tests the stance's premise against the user's own selection, and
//     reports whether the premise holds, is strained, or has failed.
package philosophy

import (
	"sort"
	"strings"

	"github.com/incurl/sumate/internal/curated"
	"github.com/incurl/sumate/internal/model"
)

// Fit describes how well a stance's premise matches the reader's situation.
type Fit string

const (
	// FitHeld means the premise is met: the stance helps.
	FitHeld Fit = "held"
	// FitStrained means the premise is partially met: the stance helps, but
	// only with a caveat.
	FitStrained Fit = "strained"
	// FitFailed means the premise does not hold: the stance is likely to
	// mislead rather than comfort.
	FitFailed Fit = "failed"
	// FitUnknown means there is not enough of the reader's situation known
	// to judge.
	FitUnknown Fit = "unknown"
)

// Assessment is the result of testing a stance against the reader.
type Assessment struct {
	// Stance is the key of the stance.
	Stance string `json:"stance"`
	// Name is the display name of the stance.
	Name string `json:"name"`
	// Claim is the proposition the stance asserts.
	Claim string `json:"claim"`
	// Fit is how well the premise matches the reader.
	Fit Fit `json:"fit"`
	// Premise is what must hold for the stance to help.
	Premise string `json:"premise"`
	// Shadow is what the stance becomes when the premise fails.
	Shadow string `json:"shadow"`
	// Earned describes the condition under which Su Dongpo could say it.
	Earned string `json:"earned"`
	// Supports are the reader's tags that satisfy the premise.
	Supports []string `json:"supports"`
	// Undermines are the reader's tags that strain or break the premise.
	Undermines []string `json:"undermines"`
	// Note is the sentence shown to the reader.
	Note string `json:"note"`
}

// Reader is the reader's known situation.
type Reader struct {
	// Tags are the tag ids the reader has selected.
	Tags map[model.RootKey][]string
	// Stages are the life stages the reader's situation corresponds to.
	Stages []model.LifeStage
}

// HasTag reports whether the reader has selected a tag.
func (r Reader) HasTag(tag string) bool {
	for _, tags := range r.Tags {
		for _, t := range tags {
			if t == tag || strings.HasPrefix(t, tag+".") {
				return true
			}
		}
	}
	return false
}

// Assess tests a stance against the reader.
//
// The test is deliberately simple and auditable: the stance lists the tags
// that satisfy its premise and the tags that undermine it, and the reader's
// selection is checked against both lists.  Nothing is inferred beyond what
// the user actually chose.
func Assess(s curated.Stance, r Reader) Assessment {
	a := Assessment{
		Stance:  s.Key,
		Name:    s.Name,
		Claim:   s.Claim,
		Premise: s.Premise,
		Shadow:  s.Shadow,
		Earned:  s.Earned,
	}

	supportHits, undermineHits := 0, 0
	for _, tag := range s.Supports {
		if r.HasTag(tag) {
			a.Supports = append(a.Supports, tag)
			supportHits++
		}
	}
	for _, tag := range s.Undermines {
		if r.HasTag(tag) {
			a.Undermines = append(a.Undermines, tag)
			undermineHits++
		}
	}

	switch {
	case supportHits > 0 && undermineHits == 0:
		a.Fit = FitHeld
	case supportHits > 0 && undermineHits > 0:
		a.Fit = FitStrained
	case supportHits == 0 && undermineHits > 0:
		a.Fit = FitFailed
	case supportHits == 0 && undermineHits == 0:
		a.Fit = FitUnknown
	}

	a.Note = noteFor(a)
	return a
}

// noteFor composes the sentence the product shows the reader.
//
// The note is the heart of the feature: it says *when* the verse speaks for
// them, and *when* it would not.  A verse that cannot say both is a
// quotation, not a mouthpiece.
func noteFor(a Assessment) string {
	switch a.Fit {
	case FitHeld:
		return "这一句此刻是站在你的位置上说的——" + trimDot(a.Premise) + "。"
	case FitStrained:
		return "这一句大体站在你这边，但" + trimDot(a.Premise) + "，若尚不成立，就只剩下一半意思。"
	case FitFailed:
		// The note must use the stance's *own* shadow, not a generic one:
		// the way a consolation misfires is different every time.
		return "这一句未必是给你的。它成立的前提是「" + trimDot(a.Premise) +
			"」；若这前提不在，它就变成：" + a.Shadow
	case FitUnknown:
		return "这一句说的是——" + a.Claim + "。它是否合你的处境，要看你站在哪里。"
	}
	return ""
}

// AssessAll evaluates every stance the reader's situation touches and
// returns them ordered from "most applicable" to "least".
func AssessAll(r Reader) []Assessment {
	stances := curated.Stances()
	out := make([]Assessment, 0, len(stances))
	for _, key := range curated.StanceOrder {
		s, ok := stances[key]
		if !ok {
			continue
		}
		if !stanceTouches(s, r) {
			continue
		}
		out = append(out, Assess(s, r))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Fit != out[j].Fit {
			return fitRank(out[i].Fit) < fitRank(out[j].Fit)
		}
		return out[i].Stance < out[j].Stance
	})
	return out
}

// stanceTouches reports whether the reader's selection is relevant to the
// stance at all.
func stanceTouches(s curated.Stance, r Reader) bool {
	for _, tag := range s.Supports {
		if r.HasTag(tag) {
			return true
		}
	}
	for _, tag := range s.Undermines {
		if r.HasTag(tag) {
			return true
		}
	}
	return false
}

// fitRank orders the fits from best to worst.
func fitRank(f Fit) int {
	switch f {
	case FitHeld:
		return 0
	case FitStrained:
		return 1
	case FitUnknown:
		return 2
	case FitFailed:
		return 3
	}
	return 4
}

// StancesOf returns the stances a verse carries, by looking at the
// concepts the verse matched.
//
// The mapping is explicit rather than inferred: a concept names the stance
// it expresses.  This keeps the philosophical layer auditable -- every
// stance attached to a verse can be traced back to a hand-written claim.
func StancesOf(v model.Verse) []string {
	stances := curated.Stances()
	seen := map[string]bool{}
	var out []string
	for _, ck := range v.Annotation.Concepts {
		s, ok := stances[ck]
		if !ok {
			continue
		}
		if seen[s.Key] {
			continue
		}
		seen[s.Key] = true
		out = append(out, s.Key)
	}
	return out
}

// StanceOfConcepts returns the stances carried by a set of concepts.
func StanceOfConcepts(concepts []string) []string {
	stances := curated.Stances()
	seen := map[string]bool{}
	var out []string
	for _, ck := range concepts {
		s, ok := stances[ck]
		if !ok {
			continue
		}
		if seen[s.Key] {
			continue
		}
		seen[s.Key] = true
		out = append(out, s.Key)
	}
	return out
}

// trimDot removes a trailing full stop so that a premise can be embedded in
// a larger sentence.
func trimDot(s string) string {
	return strings.TrimRight(strings.TrimSpace(s), "。.")
}
