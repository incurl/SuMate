#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Audit the concept lexicon: how many verses does each keyword actually fire on?

The oblique discipline (SKILL_03b) says a lexicon's quality is measured by
what it *excludes*.  A keyword that fires on a large fraction of the corpus
is a high-frequency ordinary word, not a distinguishing feature.
"""
import collections, json, os, re, sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'tools'))

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CJK = re.compile(r'[一-鿿]')

corpus = json.load(open(os.path.join(ROOT, 'data', 'dongpo_corpus.json'), encoding='utf-8'))
texts = [p['text'] for p in corpus['poems'] if p['text']]

# Pull the keyword lists out of the Go source (single source of truth).
src = open(os.path.join(ROOT, 'internal', 'curated', 'knowledge.go'), encoding='utf-8').read()
src += open(os.path.join(ROOT, 'internal', 'curated', 'knowledge_extra.go'), encoding='utf-8').read()

kws = []
for m in re.finditer(r'Keywords:\s*\[\]string\{([^}]*)\}', src):
    for k in re.findall(r'"([^"]+)"', m.group(1)):
        if k not in kws:
            kws.append(k)

print('distinct keywords: %d' % len(kws))
specific = [k for k in kws if len(k) >= 2]
single = [k for k in kws if len(k) == 1]
print('  specific (>=2 chars): %d   single (weak evidence): %d'
      % (len(specific), len(single)))

# How many poems does a specific keyword actually reach?
covered = sum(1 for t in texts if any(k in t for k in specific))
print('  poems reached by a specific keyword: %d/%d (%.0f%%)'
      % (covered, len(texts), 100.0 * covered / len(texts)))
rows = []
for k in kws:
    hits = sum(1 for t in texts if k in t)
    rows.append((hits, k))
rows.sort(reverse=True)

print()
print('--- keywords firing on >5%% of the corpus (suspicious) ---')
n = len(texts)
for hits, k in rows:
    if hits > n * 0.05:
        print('  %-8s %4d  (%.1f%%)' % (k, hits, 100.0*hits/n))
print()
print('--- keywords firing on <=1 verse (dead weight) ---')
dead = [(h, k) for h, k in rows if h <= 1]
print('  count: %d' % len(dead))
for h, k in dead[:40]:
    print('  %-10s %d' % (k, h))
