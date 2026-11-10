// SuMate (东坡嘴替) — the progressive spiritual mining frontend.
//
// The user drills down through the five root dimensions; every selection
// refreshes the candidate count, records a breadcrumb and allows free
// backtracking.  At any level the user may preview the current matches, and
// at the leaf the system converges on one verse.
import { api } from "./api";
import { clear, el } from "./dom";
import {
  ROOT_LABEL,
  renderCrumbs,
  renderMatch,
  renderNode,
  renderReading,
  renderVerse,
} from "./render";
import {
  deepestOf,
  pushPath,
  truncatePath,
  type AppState,
} from "./state";
import { RootKey } from "../bindings/github.com/incurl/sumate/internal/model/models";
import type { MatchResult, TreeNode } from "./types";

const ROOT_ORDER: RootKey[] = [
  RootKey.RootMood,
  RootKey.RootPlace,
  RootKey.RootWeather,
  RootKey.RootRelation,
  RootKey.RootOccasion,
];

const state: AppState = {
  path: [],
  visible: [],
  crumbs: [],
  preview: [],
  activeRoot: "",
  focus: "",
};

let match: MatchResult | null = null;
let previewOpen = false;

// ---------------------------------------------------------------------------
// DOM handles
// ---------------------------------------------------------------------------
const $ = <T extends HTMLElement>(id: string): T =>
  document.getElementById(id) as T;

const treeEl = $("tree");
const crumbsEl = $("crumbs");
const previewListEl = $("preview-list");
const matchEl = $("match");
const readingEl = $("reading");
const rootsEl = $("roots");
const previewBtn = $("preview-btn") as HTMLButtonElement;
const resetBtn = $("reset-btn") as HTMLButtonElement;
const searchEl = $("search") as HTMLInputElement;
const searchResultsEl = $("search-results");
const statsEl = $("stats");
const hintEl = $("hint");

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------
async function loadRoots(): Promise<void> {
  const roots = await api.roots();
  state.visible = roots;
  state.activeRoot = "";
  state.focus = "";
  state.crumbs = await api.breadcrumb(state.path);
  await refreshPreview();
  drawRoots();
  drawTree();
  drawCrumbs();
}

/**
 * Show the children of the deepest node selected in `root`.
 *
 * When the dimension has not been entered yet its own children are shown;
 * otherwise the drilling continues below the node already chosen.
 */
async function showDimension(root: RootKey): Promise<void> {
  state.activeRoot = root;
  const deepest = deepestOf(state.path, root);
  if (deepest) {
    const children = await api.children(deepest.NodeID, state.path);
    state.visible = children;
    state.focus = deepest.NodeID;
  } else {
    const roots = await api.roots();
    const rootNode = roots.find((r) => r.root === root);
    if (rootNode) {
      const children = await api.children(rootNode.id, state.path);
      state.visible = children;
      state.focus = rootNode.id;
    }
  }
  await refreshPreview();
  drawRoots();
  drawTree();
  drawCrumbs();
}

async function refreshPreview(): Promise<void> {
  state.preview = await api.preview(state.path, 60);
  state.crumbs = await api.breadcrumb(state.path);
  await refreshReading();
}

async function refreshReading(): Promise<void> {
  const assessments = await api.philosophicalReading(state.path);
  clear(readingEl);
  readingEl.appendChild(renderReading(assessments));
}

// ---------------------------------------------------------------------------
// Drawing
// ---------------------------------------------------------------------------
function drawRoots(): void {
  clear(rootsEl);
  for (const r of ROOT_ORDER) {
    const btn = el("button", {
      class: `root-tab root-tab--${r}`,
      text: ROOT_LABEL[r],
      attrs: { type: "button" },
    });
    if (r === state.activeRoot) btn.classList.add("is-active");
    const depth = state.path.filter((f) => f.Root === r).length;
    if (depth > 0) {
      btn.classList.add("has-selection");
      btn.appendChild(el("span", { class: "root-tab__dot", text: "•" }));
    }
    btn.addEventListener("click", () => void showDimension(r));
    rootsEl.appendChild(btn);
  }
}

function drawTree(): void {
  clear(treeEl);
  if (state.visible.length === 0) {
    treeEl.appendChild(
      el("p", {
        class: "empty",
        text: "这一层之下没有可匹配的句子，换一个方向试试。",
      }),
    );
    return;
  }
  const grid = el("div", { class: "node-grid" });
  for (const node of state.visible) {
    grid.appendChild(renderNode(node, onSelectNode));
  }
  treeEl.appendChild(grid);

  hintEl.textContent =
    state.activeRoot === ""
      ? "选择上方一个维度开始逐层挖掘"
      : `在「${ROOT_LABEL[state.activeRoot]}」维度下继续选择`;
}

function drawCrumbs(): void {
  clear(crumbsEl);
  crumbsEl.appendChild(renderCrumbs(state.crumbs, onJumpCrumb));
}

function drawPreview(): void {
  clear(previewListEl);
  if (!previewOpen) return;
  if (state.preview.length === 0) {
    previewListEl.appendChild(
      el("p", { class: "empty", text: "当前组合没有匹配结果。" }),
    );
    return;
  }
  const total = state.crumbs.length
    ? state.crumbs[state.crumbs.length - 1].remaining
    : state.preview.length;
  previewListEl.appendChild(
    el("p", {
      class: "preview__count",
      text: `当前组合共 ${total} 句，展示前 ${state.preview.length} 句`,
    }),
  );
  for (const v of state.preview) {
    previewListEl.appendChild(renderVerse(v));
  }
}

// ---------------------------------------------------------------------------
// Interaction
// ---------------------------------------------------------------------------
async function onSelectNode(node: TreeNode): Promise<void> {
  if (node.remaining === 0) return;
  state.path = pushPath(state.path, node);
  match = null;
  clear(matchEl);

  if (node.is_leaf) {
    await runMatch();
  }

  const children = await api.children(node.id, state.path);
  state.visible = children.length > 0 ? children : [node];
  state.focus = node.id;
  state.activeRoot = node.root;

  await refreshPreview();
  drawRoots();
  drawTree();
  drawCrumbs();
  drawPreview();
}

async function onJumpCrumb(depth: number): Promise<void> {
  state.path = truncatePath(state.path, depth);
  match = null;
  clear(matchEl);
  await refreshPreview();

  const last = state.path[state.path.length - 1];
  if (last) {
    await showDimension(last.Root);
  } else {
    await loadRoots();
  }
  drawPreview();
}

async function runMatch(): Promise<void> {
  const description = searchEl.value.trim();
  match = await api.match(state.path, description);
  clear(matchEl);
  matchEl.appendChild(renderMatch(match));
  matchEl.classList.add("is-filled");
}

async function doSearch(): Promise<void> {
  const q = searchEl.value.trim();
  clear(searchResultsEl);
  if (q === "") return;
  const vs = await api.search(q, 20);
  if (vs.length === 0) {
    searchResultsEl.appendChild(
      el("p", { class: "empty", text: "没有找到相关的句子。" }),
    );
    return;
  }
  searchResultsEl.appendChild(
    el("p", { class: "preview__count", text: `找到 ${vs.length} 句` }),
  );
  for (const v of vs) {
    searchResultsEl.appendChild(renderVerse(v, true));
  }
}

async function reset(): Promise<void> {
  state.path = [];
  state.visible = [];
  state.preview = [];
  match = null;
  clear(matchEl);
  matchEl.classList.remove("is-filled");
  previewOpen = false;
  previewBtn.classList.remove("is-active");
  previewBtn.textContent = "预览当前匹配";
  await loadRoots();
  drawPreview();
}

function togglePreview(): void {
  previewOpen = !previewOpen;
  previewBtn.classList.toggle("is-active", previewOpen);
  previewBtn.textContent = previewOpen ? "收起预览" : "预览当前匹配";
  drawPreview();
}

// ---------------------------------------------------------------------------
// Boot
// ---------------------------------------------------------------------------
async function boot(): Promise<void> {
  const stats = await api.stats();
  clear(statsEl);
  statsEl.appendChild(
    el("span", {
      text: `${stats.verses} 句 · ${stats.tunes} 词牌 · ${stats.couplets} 联`,
    }),
  );

  previewBtn.addEventListener("click", togglePreview);
  resetBtn.addEventListener("click", () => void reset());
  searchEl.addEventListener("input", () => void doSearch());
  searchEl.addEventListener("keydown", (ev) => {
    if ((ev as KeyboardEvent).key === "Enter") void doSearch();
  });

  await loadRoots();
  drawPreview();
}

void boot();
