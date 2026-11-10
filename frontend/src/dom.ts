// Tiny DOM helpers.  The app is small enough that a framework would only
// add ceremony; these helpers keep the render code declarative.

/** Create an element with optional class, text and attributes. */
export function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  opts: {
    class?: string;
    text?: string;
    html?: string;
    attrs?: Record<string, string>;
    on?: Partial<Record<keyof HTMLElementEventMap, (ev: Event) => void>>;
  } = {},
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  if (opts.class) node.className = opts.class;
  if (opts.text !== undefined) node.textContent = opts.text;
  if (opts.html !== undefined) node.innerHTML = opts.html;
  if (opts.attrs) {
    for (const [k, v] of Object.entries(opts.attrs)) node.setAttribute(k, v);
  }
  if (opts.on) {
    for (const [k, v] of Object.entries(opts.on)) {
      node.addEventListener(k, v as EventListener);
    }
  }
  return node;
}

/** Remove every child of a node. */
export function clear(node: HTMLElement): void {
  while (node.firstChild) node.removeChild(node.firstChild);
}

/** Replace the children of a node. */
export function replace(node: HTMLElement, children: HTMLElement[]): void {
  clear(node);
  for (const c of children) node.appendChild(c);
}
