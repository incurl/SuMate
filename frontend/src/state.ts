// Application state: the drilling path, the derived views and the helpers
// that keep the breadcrumb and the tree in sync.
//
// The selection is stored as a single ordered path, but grouped by root
// dimension: every root owns a contiguous segment of the path.  This is what
// lets the user drill inside the mood dimension, switch to the weather
// dimension, drill there too, and still have a single coherent selection.
import { RootKey } from "../bindings/github.com/sumate/sumate/internal/model/models";
import type { Breadcrumb, Filter, TreeNode, Verse } from "./types";

export interface AppState {
  /** The full drilling path, grouped by root dimension. */
  path: Filter[];
  /** The nodes currently displayed (the children of the focused node). */
  visible: TreeNode[];
  /** The breadcrumb derived from the path. */
  crumbs: Breadcrumb[];
  /** Verses matching the current selection (the arbitrary-layer preview). */
  preview: Verse[];
  /** The dimension the tree is currently showing. */
  activeRoot: RootKey | "";
  /** The node whose children are shown; "" means the root's own children. */
  focus: string;
}

export function emptyState(): AppState {
  return {
    path: [],
    visible: [],
    crumbs: [],
    preview: [],
    activeRoot: "",
    focus: "",
  };
}

/**
 * The deepest selected node of a root dimension, or null when the dimension
 * has not been entered yet.
 */
export function deepestOf(path: Filter[], root: RootKey): Filter | null {
  let found: Filter | null = null;
  for (const f of path) {
    if (f.Root === root) found = f;
  }
  return found;
}

/**
 * Append a node to the path inside its own root segment.
 *
 * The path is rebuilt so that the segments stay grouped and ordered by the
 * canonical root order.
 */
export function pushPath(path: Filter[], node: TreeNode): Filter[] {
  // Drop any previous selection inside the same root: picking a new option
  // at a shallower level replaces the deeper one.
  const kept = path.filter((f) => f.Root !== node.root);
  kept.push({ NodeID: node.id, Root: node.root });
  return sortByRoot(kept);
}

/** Order the path so that the root segments follow the canonical order. */
export function sortByRoot(path: Filter[]): Filter[] {
  const order: RootKey[] = [
    RootKey.RootMood,
    RootKey.RootPlace,
    RootKey.RootWeather,
    RootKey.RootRelation,
    RootKey.RootOccasion,
  ];
  const out: Filter[] = [];
  for (const root of order) {
    for (const f of path) {
      if (f.Root === root) out.push(f);
    }
  }
  return out;
}

/**
 * Truncate the path to `depth` steps.
 *
 * Clicking a breadcrumb calls this to go back to an earlier state.
 */
export function truncatePath(path: Filter[], depth: number): Filter[] {
  return path.slice(0, depth);
}

/** The number of selected steps in a root dimension. */
export function depthInRoot(path: Filter[], root: RootKey): number {
  return path.filter((f) => f.Root === root).length;
}
