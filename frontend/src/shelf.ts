// The shelf of unnamed verses.
//
// Every verse in the corpus can be retrieved, but a verse is only
// *interpretable* if a concept reached it.  The ones that matched nothing
// cannot appear in the tag tree, so without a place of their own they would
// sit in the corpus and be unreachable from the interface.
//
// The shelf is not a trophy cabinet and not an error list.  It is the edge of
// the taxonomy made visible: a place where a reader can see what the
// annotation could not name, and why.
import { el } from "./dom";
import type { UnnamedShelf, UnnamedVerse } from "./types";

/** Render one unnamed verse. */
function renderUnnamed(uv: UnnamedVerse): HTMLElement {
  const card = el("article", { class: "unnamed" });

  const top = el("div", { class: "unnamed__top" });
  top.appendChild(el("span", { class: "unnamed__tune", text: uv.verse.tune }));
  if (uv.verse.background.year) {
    top.appendChild(
      el("span", { class: "unnamed__year", text: uv.verse.background.year }),
    );
  }
  // A hint, not a verdict: the verse holds a character seen nowhere else in
  // the corpus, so a corrupted glyph is one possible reason for the miss.
  if (uv.suspect_glyph) {
    top.appendChild(
      el("span", {
        class: "unnamed__hint",
        text: "疑讹",
        attrs: { title: "含全语料仅见一次的字，可能为传本讹字" },
      }),
    );
  }
  card.appendChild(top);

  card.appendChild(el("p", { class: "unnamed__text", text: uv.verse.text }));
  return card;
}

/** Render the whole shelf. */
export function renderShelf(
  shelf: UnnamedShelf,
  onBack: () => void,
): HTMLElement {
  const wrap = el("section", { class: "shelf" });

  // --- the honest preamble
  const head = el("header", { class: "shelf__head" });
  head.appendChild(el("h2", { class: "shelf__title", text: shelf.title }));
  head.appendChild(
    el("span", {
      class: "shelf__count",
      text: `${shelf.total} / ${shelf.corpus}`,
    }),
  );
  wrap.appendChild(head);
  wrap.appendChild(el("p", { class: "shelf__preamble", text: shelf.preamble }));

  const back = el("button", {
    class: "shelf__back",
    text: "← 回到五维",
    attrs: { type: "button" },
  });
  back.addEventListener("click", onBack);
  wrap.appendChild(back);

  // --- the groups, most readable first
  for (const g of shelf.groups ?? []) {
    const section = el("section", { class: `shelf__group shelf__group--${g.reason}` });

    const gh = el("div", { class: "shelf__grouphead" });
    gh.appendChild(el("h3", { class: "shelf__grouplabel", text: g.label }));
    gh.appendChild(el("span", { class: "shelf__groupcount", text: String(g.count) }));
    section.appendChild(gh);

    if (g.note) {
      section.appendChild(el("p", { class: "shelf__groupnote", text: g.note }));
    }

    const grid = el("div", { class: "shelf__verses" });
    for (const uv of g.verses ?? []) {
      grid.appendChild(renderUnnamed(uv));
    }
    section.appendChild(grid);
    wrap.appendChild(section);
  }

  return wrap;
}
