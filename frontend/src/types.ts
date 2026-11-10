// The domain types come straight from the Wails-generated bindings, so the
// frontend and the Go backend can never drift apart.
import { RootKey } from "../bindings/github.com/sumate/sumate/internal/model/models";

export type {
  Annotation,
  Background,
  Breadcrumb,
  Emotion,
  MatchExplanation,
  MatchResult,
  Polarity,
  Stats,
  Strength,
  Tangent,
  TreeNode,
  Verse,
} from "../bindings/github.com/sumate/sumate/internal/model/models";

export type { Filter } from "../bindings/github.com/sumate/sumate/internal/tagtree/models";

export { RootKey };

/** The five dimensions in their canonical display order. */
export const ROOT_ORDER: RootKey[] = [
  RootKey.RootMood,
  RootKey.RootPlace,
  RootKey.RootWeather,
  RootKey.RootRelation,
  RootKey.RootOccasion,
];
