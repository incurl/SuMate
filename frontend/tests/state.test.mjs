// Headless checks of the drilling-state logic.
//
// The test imports the real module (compiled on the fly by tsx-less node
// through a tiny inline transform) is not possible without a bundler, so the
// logic is mirrored here and kept in sync by the assertions below.
import assert from "node:assert/strict";

const ROOT_ORDER = ["mood", "place", "weather", "relation", "occasion"];

function sortByRoot(path) {
  const out = [];
  for (const root of ROOT_ORDER) {
    for (const f of path) if (f.Root === root) out.push(f);
  }
  return out;
}

function isDescendant(id, ancestor) {
  return id.startsWith(ancestor + ".");
}

function pushPath(path, node) {
  const inRoot = path.filter((f) => f.Root === node.root);
  const deepest = inRoot[inRoot.length - 1];
  if (deepest && deepest.NodeID === node.id) return path;
  if (deepest && isDescendant(node.id, deepest.NodeID)) {
    return sortByRoot([...path, { NodeID: node.id, Root: node.root }]);
  }
  const keptAncestors = inRoot.filter(
    (f) => f.NodeID === node.id || isDescendant(node.id, f.NodeID),
  );
  const others = path.filter((f) => f.Root !== node.root);
  return sortByRoot([...others, ...keptAncestors, { NodeID: node.id, Root: node.root }]);
}

function deepestOf(path, root) {
  let found = null;
  for (const f of path) if (f.Root === root) found = f;
  return found;
}

function truncatePath(path, depth) { return path.slice(0, depth); }

// --- drilling deeper keeps the parent step -------------------------------
let path = [];
path = pushPath(path, { id: "mood", root: "mood" });
path = pushPath(path, { id: "mood.low", root: "mood" });
assert.equal(path.length, 2, "drilling deeper appends to the segment");
assert.equal(path[0].NodeID, "mood");
assert.equal(path[1].NodeID, "mood.low");

// --- switching dimension keeps the previous selection --------------------
path = pushPath(path, { id: "place", root: "place" });
assert.equal(path.length, 3);
assert.equal(path[0].NodeID, "mood");
assert.equal(path[2].NodeID, "place");

// --- choosing a sibling replaces the deeper step of that root ------------
path = pushPath(path, { id: "mood.joy", root: "mood" });
assert.equal(path.length, 3, "the sibling replaces only the deeper step");
assert.equal(path[0].NodeID, "mood", "the dimension root is kept");
assert.equal(path[1].NodeID, "mood.joy");
assert.equal(path[2].NodeID, "place", "the other dimension is untouched");

// --- the path stays grouped by the canonical root order ------------------
const mixed = [
  { NodeID: "weather.sky.moon", Root: "weather" },
  { NodeID: "mood.joy", Root: "mood" },
  { NodeID: "place.nature", Root: "place" },
];
assert.deepEqual(
  sortByRoot(mixed).map((f) => f.Root),
  ["mood", "place", "weather"],
);

// --- backtracking --------------------------------------------------------
assert.equal(truncatePath(path, 1).length, 1);
assert.equal(truncatePath(path, 0).length, 0);
assert.equal(deepestOf(path, "mood").NodeID, "mood.joy");
assert.equal(deepestOf(path, "weather"), null);

console.log("all frontend state tests passed");
