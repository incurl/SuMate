// Rendering of the tree nodes, the breadcrumb, the preview list and the
// match card.
import { clear, el, replace } from "./dom";
import type { Breadcrumb, TreeNode, Verse } from "./types";
import type { MatchResult } from "./types";

export const ROOT_LABEL: Record<string, string> = {
  mood: "心境",
  place: "地点",
  weather: "时令",
  relation: "人际",
  occasion: "境遇",
};

const STAGE_LABEL: Record<string, string> = {
  early_career: "早年科第",
  in_office: "宦海升沉",
  reform_conflict: "党争外任",
  huangzhou: "黄州贬谪",
  return: "哲宗朝复起",
  huizhou: "惠州贬谪",
  danzhou: "儋州贬谪",
  leisure: "退隐闲居",
  travel: "羁旅行役",
  late_years: "晚年归途",
};

const POLARITY_LABEL: Record<string, string> = {
  positive: "正面",
  neutral: "中性",
  melanchol: "低回",
  mixed: "复合",
};

const TANGENT_LABEL: Record<string, string> = {
  sorrow_with_transcendence: "悲而能超",
  hardship_with_optimism: "苦中作乐",
  reconciliation_with_imperfection: "与不完美和解",
  open_minded_tolerance: "挫败后的豁达",
  freedom_from_worldly_obsession: "不萦于物",
  peaceful_acceptance_of_fate: "平静接受命运",
};

const DIMENSION_LABEL: Record<string, string> = {
  positive: "正面情绪",
  neutral: "中性哲思",
  melanchol: "低回疗愈",
  dongpo: "东坡气质",
};

/** Render one tree node as a clickable card. */
export function renderNode(
  node: TreeNode,
  onSelect: (n: TreeNode) => void,
): HTMLElement {
  const card = el("button", { class: "node", attrs: { type: "button" } });
  card.classList.add(`node--${node.root}`);
  if (node.remaining === 0) card.classList.add("is-empty");

  const head = el("div", { class: "node__head" });
  head.appendChild(el("span", { class: "node__label", text: node.label }));
  head.appendChild(
    el("span", { class: "node__count", text: String(node.remaining) }),
  );
  card.appendChild(head);

  if (node.hint) {
    card.appendChild(el("div", { class: "node__hint", text: node.hint }));
  }
  if (!node.is_leaf) {
    card.appendChild(el("span", { class: "node__more", text: "▸" }));
  }
  card.addEventListener("click", () => onSelect(node));
  return card;
}

/** Render the breadcrumb trail. */
export function renderCrumbs(
  crumbs: Breadcrumb[],
  onJump: (depth: number) => void,
): HTMLElement {
  const nav = el("nav", { class: "crumbs" });
  if (crumbs.length === 0) {
    nav.appendChild(el("span", { class: "crumbs__lead", text: "从五个维度开始" }));
    return nav;
  }
  const lead = el("button", {
    class: "crumbs__item crumbs__lead",
    text: "重来",
    attrs: { type: "button" },
  });
  lead.addEventListener("click", () => onJump(0));
  nav.appendChild(lead);

  crumbs.forEach((c, i) => {
    nav.appendChild(el("span", { class: "crumbs__sep", text: "›" }));
    const item = el("button", {
      class: "crumbs__item",
      text: c.label,
      attrs: { type: "button", title: (c.path ?? []).join(" › ") },
    });
    item.addEventListener("click", () => onJump(i + 1));
    nav.appendChild(item);
    const count = el("span", { class: "crumbs__count", text: `${c.remaining}` });
    nav.appendChild(count);
  });
  return nav;
}

/** Render a verse as a preview card. */
export function renderVerse(v: Verse, compact = false): HTMLElement {
  const card = el("article", { class: "verse" });
  card.classList.toggle("verse--compact", compact);

  const top = el("div", { class: "verse__top" });
  top.appendChild(el("span", { class: "verse__tune", text: v.tune }));
  if (v.background.year) {
    top.appendChild(el("span", { class: "verse__year", text: v.background.year }));
  }
  if (v.background.place) {
    top.appendChild(el("span", { class: "verse__place", text: v.background.place }));
  }
  top.appendChild(
    el("span", {
      class: "verse__kind",
      text: v.kind === "couplet" ? "联" : "句",
    }),
  );
  card.appendChild(top);

  card.appendChild(el("p", { class: "verse__text", text: v.text }));

  if (!compact) {
    const a = v.annotation;
    card.appendChild(el("p", { class: "verse__literal", text: a.literal }));
    if (a.conception) {
      card.appendChild(el("p", { class: "verse__conception", text: a.conception }));
    }
    if (a.philosophy) {
      card.appendChild(el("p", { class: "verse__philosophy", text: a.philosophy }));
    }
    const tags = el("div", { class: "verse__tags" });
    tags.appendChild(
      el("span", {
        class: "tag tag--polarity",
        text: POLARITY_LABEL[a.polarity] ?? a.polarity,
      }),
    );
    for (const e of (a.emotions ?? []).slice(0, 4)) {
      tags.appendChild(
        el("span", {
          class: `tag tag--${e.dimension}`,
          text: e.name,
          attrs: { title: DIMENSION_LABEL[e.dimension] ?? e.dimension },
        }),
      );
    }
    for (const t of (a.tangents ?? [])) {
      tags.appendChild(
        el("span", {
          class: "tag tag--dongpo",
          text: TANGENT_LABEL[t] ?? t,
        }),
      );
    }
    card.appendChild(tags);
    if (v.background.historical) {
      card.appendChild(
        el("p", { class: "verse__history", text: v.background.historical }),
      );
    }
  }
  return card;
}

/** Render the final match card with the explanation. */
export function renderMatch(result: MatchResult): HTMLElement {
  const wrap = el("section", { class: "match" });

  const head = el("header", { class: "match__head" });
  head.appendChild(el("h2", { class: "match__title", text: "东坡嘴替" }));
  head.appendChild(
    el("span", {
      class: "match__score",
      text: `${result.explanation.overall}`,
      attrs: { title: "综合匹配分" },
    }),
  );
  wrap.appendChild(head);

  const p = result.primary;
  const quote = el("div", { class: "match__quote" });
  quote.appendChild(el("p", { class: "match__text", text: p.text }));
  const meta = el("div", { class: "match__meta" });
  meta.appendChild(el("span", { class: "match__tune", text: p.tune }));
  if (p.background.year) {
    meta.appendChild(el("span", { class: "match__year", text: p.background.year }));
  }
  if (p.background.place) {
    meta.appendChild(el("span", { class: "match__place", text: p.background.place }));
  }
  meta.appendChild(
    el("span", {
      class: "match__stage",
      text: STAGE_LABEL[p.background.stage] ?? p.background.stage,
    }),
  );
  quote.appendChild(meta);
  wrap.appendChild(quote);

  const a = p.annotation;
  const reading = el("div", { class: "match__reading" });
  reading.appendChild(el("h3", { text: "为什么是这一句" }));
  reading.appendChild(el("p", { class: "reading__literal", text: a.literal }));
  if (a.conception) {
    reading.appendChild(el("p", { class: "reading__conception", text: a.conception }));
  }
  if (a.philosophy) {
    reading.appendChild(el("p", { class: "reading__philosophy", text: a.philosophy }));
  }
  if (a.attitude) {
    reading.appendChild(el("p", { class: "reading__attitude", text: `态度：${a.attitude}` }));
  }
  if (a.healing) {
    reading.appendChild(el("p", { class: "reading__healing", text: `疗愈：${a.healing}` }));
  }
  wrap.appendChild(reading);

  const scores = el("div", { class: "match__scores" });
  const e = result.explanation;
  const rows: [string, number][] = [
    ["意境契合", e.semantic],
    ["情绪共鸣", e.emotional],
    ["情境一致", e.scenario],
    ["东坡气质", e.temperament],
  ];
  for (const [label, value] of rows) {
    const row = el("div", { class: "score" });
    row.appendChild(el("span", { class: "score__label", text: label }));
    const bar = el("span", { class: "score__bar" });
    const fill = el("span", { class: "score__fill" });
    fill.style.width = `${Math.max(0, Math.min(100, value))}%`;
    bar.appendChild(fill);
    row.appendChild(bar);
    row.appendChild(el("span", { class: "score__value", text: String(value) }));
    scores.appendChild(row);
  }
  wrap.appendChild(scores);

  if ((e.reasons ?? []).length > 0) {
    const reasons = el("ul", { class: "match__reasons" });
    for (const r of e.reasons ?? []) {
      reasons.appendChild(el("li", { text: r }));
    }
    wrap.appendChild(reasons);
  }

  if ((result.alternatives ?? []).length > 0) {
    const alt = el("div", { class: "match__alts" });
    alt.appendChild(el("h3", { text: "备选" }));
    for (const v of result.alternatives ?? []) {
      alt.appendChild(renderVerse(v, true));
    }
    wrap.appendChild(alt);
  }

  wrap.appendChild(
    el("p", {
      class: "match__candidates",
      text: `共在 ${result.candidates} 句中为你选出`,
    }),
  );
  return wrap;
}

/** Replace the children of a container. */
export function renderInto(container: HTMLElement, children: HTMLElement[]): void {
  replace(container, children);
}

export { clear };
