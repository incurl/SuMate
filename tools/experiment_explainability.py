#!/usr/bin/env python3
"""Can a vector match still explain itself?

The project's value proposition includes auditability: the engine says *why*
a verse matched (concepts, emotions, temperaments, caveats).  A pure vector
score is a number without reasons.  The question is whether the two can be
combined without losing either.
"""
import re, json, sys
import numpy as np
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from embed_model import BertEmbedder

CJK = re.compile(r'[㐀-䶿一-鿿]')
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MODEL_DIR = os.environ.get('SUMATE_EMBED_DIR', '/tmp/embed')

src = (open(f'{ROOT}/internal/curated/knowledge.go', encoding='utf-8').read()
       + open(f'{ROOT}/internal/curated/knowledge_extra.go', encoding='utf-8').read())
concepts = []
for m in re.finditer(r'Key: "([^"]+)", Name: "([^"]+)"(.*?)Keywords:\s*\[\]string\{([^}]*)\}', src, re.S):
    kws = re.findall(r'"([^"]+)"', m.group(4))
    concepts.append({'key': m.group(1), 'name': m.group(2), 'keywords': kws})

corpus = json.load(open(f'{ROOT}/data/dongpo_corpus.json', encoding='utf-8'))
poems = [p for p in corpus['poems'] if p['text']]

emb = BertEmbedder(MODEL_DIR)
texts = [p['text'] for p in poems]
V = emb.embed(texts)

# For each description, find the top verse by vector AND ask: can the
# concept lexicon still name why?
CASES = ["我想一个人待着", "我很累不想动", "下雨了心里空空的",
         "思念远方的人", "想起了从前", "不知道该怎么办"]

print("=== 向量找到的句子，词库能否解释？ ===")
explained = 0
for desc in CASES:
    q = emb.embed([desc])[0]
    vs = V @ q
    order = np.argsort(-vs)[:5]
    # which concepts does the top verse carry?
    top = poems[order[0]]
    carried = []
    for c in concepts:
        for kw in c['keywords']:
            if len(kw) >= 2 and kw in top['text']:
                carried.append(c['name']); break
    ok = len(carried) > 0
    explained += ok
    print(f"  「{desc}」")
    print(f"      → {top['text'][:36]}")
    print(f"      概念层可解释：{'、'.join(carried) if carried else '（无）'}  {'✓' if ok else '✗'}")
    # what are the runners-up, and do they agree?
    names = []
    for i in order[:3]:
        p = poems[i]
        cc = [c['name'] for c in concepts
              if any(len(kw) >= 2 and kw in p['text'] for kw in c['keywords'])]
        names.append((p['text'][:18], cc[:2], round(float(vs[i]), 3)))
    for t, cc, sc in names:
        print(f"         {sc}  {t:20} {cc}")
print(f"\n  前五句中词库可解释的比例：{explained}/{len(CASES)}")

# Does the vector ranking agree with the concept ranking?  If yes, the
# vector is a *bridge* to the lexicon, not a replacement for it.
print("\n=== 向量与词库是否指向同一批概念？ ===")
for desc in CASES:
    q = emb.embed([desc])[0]
    vs = V @ q
    order = np.argsort(-vs)[:20]
    # concepts the top-20 vector verses carry
    vec_cs = {}
    for i in order:
        for c in concepts:
            for kw in c['keywords']:
                if len(kw) >= 2 and kw in poems[i]['text']:
                    vec_cs[c['key']] = vec_cs.get(c['key'], 0) + 1
                    break
    top_cs = [k for k, _ in sorted(vec_cs.items(), key=lambda x: -x[1])[:3]]
    names = [c['name'] for c in concepts if c['key'] in top_cs]
    print(f"  「{desc:12}」 → {'、'.join(names)}")
