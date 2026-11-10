// The philosophical layer of the UI.
//
// Two things are shown to the reader:
//
//  1. the *caveat* on the current match -- the condition under which the
//     verse would not speak for them;
//  2. the *reading* -- the positions on how to live that are available to
//     them right now, and on what condition.
//
// The tone matters: the caveat must not sound like a disclaimer, and the
// reading must not sound like a lecture.
import type { Assessment } from "../bindings/github.com/incurl/sumate/internal/philosophy/models";
import type { PhilosophicalCaveat } from "../bindings/github.com/incurl/sumate/internal/model/models";

export type { Assessment, PhilosophicalCaveat };

/** The label shown for each fit. */
export const FIT_LABEL: Record<string, string> = {
  held: "正合",
  strained: "只合一半",
  unknown: "待定",
  failed: "不合",
};

/** The label shown for each Dongpo temperament. */
export const TANGENT_LABEL: Record<string, string> = {
  sorrow_with_transcendence: "悲而能超",
  hardship_with_optimism: "苦中作乐",
  reconciliation_with_imperfection: "与不完美和解",
  open_minded_tolerance: "挫败后的豁达",
  freedom_from_worldly_obsession: "不萦于物",
  peaceful_acceptance_of_fate: "平静接受命运",
};

/** The stance a verse takes, in the reader's own words. */
export function stanceSentence(a: Assessment): string {
  return `${a.claim}——但这一句有它的前提：${trim(a.premise)}。`;
}

/** A short, quiet warning for a caveat. */
export function caveatSentence(c: PhilosophicalCaveat): string {
  return c.note;
}

/** Trim a trailing full stop. */
function trim(s: string): string {
  return s.replace(/[。.]\s*$/, "");
}
