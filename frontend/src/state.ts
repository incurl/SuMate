// Application state: the drilling path, the derived views and the helpers
// that keep the breadcrumb and the tree in sync.
//
// The selection is stored as a single ordered path, but grouped by root
// dimension: every root owns a contiguous segment of the path.  This is what
// lets the user drill inside the mood dimension, switch to the weather
// dimension, drill there too, and still have a single coherent selection.
import { RootKey } from "../bindings/github.com/incurl/sumate/internal/model/models";
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
 * canonical root order.  A node that is a *descendant* of the deepest
 * selected node in its root extends the segment; a node at a shallower level
 * replaces the segment from that level down.
 */
export function pushPath(path: Filter[], node: TreeNode): Filter[] {
  const inRoot = path.filter((f) => f.Root === node.root);
  const deepest = inRoot[inRoot.length - 1];

  if (deepest && deepest.NodeID === node.id) {
    // already selected: nothing to change
    return path;
  }

  if (deepest && isDescendant(node.id, deepest.NodeID)) {
    // drilling deeper: keep the whole segment and append
    return sortByRoot([...path, { NodeID: node.id, Root: node.root }]);
  }

  // A different branch: keep the part of the segment that still leads to
  // the new node (its ancestors) and drop the rest.
  const keptAncestors = inRoot.filter(
    (f) => f.NodeID === node.id || isDescendant(node.id, f.NodeID),
  );
  const others = path.filter((f) => f.Root !== node.root);
  return sortByRoot([...others, ...keptAncestors, { NodeID: node.id, Root: node.root }]);
}

/**
 * Reports whether `id` lies below `ancestor` in the tag tree.
 *
 * The tree ids are dotted paths, so a descendant is recognised by its prefix.
 */
export function isDescendant(id: string, ancestor: string): boolean {
  return id.startsWith(ancestor + ".");
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
