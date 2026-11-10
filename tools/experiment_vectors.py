#!/usr/bin/env python3
"""The decisive experiment: would vectorisation help SuMate's metaphysics?

Three questions, measured rather than argued:

  Q1  Does a vector model connect the user's words to the corpus better than
      the current character-overlap matcher?
  Q2  Would vectors fix the *semantic* axis, whose standard deviation in the
      decision region is 0.00?
  Q3  What would vectors cost -- in size, in latency, in auditability?

The corpus and the lexicon come from the Go side (exported to JSON), so the
experiment runs against exactly what ships.
"""
import json, os, os, re, sys, time
import numpy as np

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
CJK = re.compile(r'[㐀-䶿一-鿿]')
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MODEL_DIR = os.environ.get('SUMATE_EMBED_DIR', '/tmp/embed')

# ---------------------------------------------------------------- data
corpus = json.load(open(f'{ROOT}/data/dongpo_corpus.json', encoding='utf-8'))
poems = [p for p in corpus['poems'] if p['text']]
texts = [p['text'] for p in poems]

# The concept lexicon, parsed out of the Go source so there is one truth.
src = (open(f'{ROOT}/internal/curated/knowledge.go', encoding='utf-8').read()
       + open(f'{ROOT}/internal/curated/knowledge_extra.go', encoding='utf-8').read())
concepts = []
for m in re.finditer(r'Key: "([^"]+)", Name: "([^"]+)".*?Keywords:\s*\[\]string\{([^}]*)\}',
                     src, re.S):
    kws = re.findall(r'"([^"]+)"', m.group(3))
    concepts.append({'key': m.group(1), 'name': m.group(2), 'keywords': kws})
print(f'corpus {len(texts)} 句 / {len(concepts)} 个概念')

# The descriptions a real user might type, and the concept a thoughtful
# reader would map each onto.
CASES = [
    ("我想一个人待着", "serene.joy.quiet"),
    ("我很累不想动", "idleness.day"),
    ("下雨了心里空空的", "rain.drizzle"),
    ("思念远方的人", "human.love"),
    ("想起了从前", "nostalgia.old_place"),
    ("不知道该怎么办", "helpless.against_fate"),
    ("有点高兴", "serene.joy.quiet"),
    ("心里空落落的", "moon.solitude"),
    ("好累啊不想动", "idleness.day"),
    ("有点想家", "nostalgia.old_place"),
]

# ------------------------------------------------------- the baselines
def char_overlap(desc, poem):
    """The current matcher: single-character set overlap with the keywords.

    This mirrors internal/engine/engine.go descriptionScore exactly.
    """
    d = set(CJK.findall(desc))
    if not d:
        return 0.0
    hit = 0
    for kw in poem['_kw']:
        for ch in kw:
            if ch in d:
                hit += 1
                break
    if hit == 0:
        return 0.0
    return min(1.0, hit / (len(poem['_kw']) + 1))

def phrase_overlap(desc, poem):
    """A better lexical matcher: multi-character keywords only."""
    hit = 0
    for kw in poem['_kw']:
        if len(CJK.findall(kw)) < 2:
            continue
        if kw in desc:
            hit += 1
    return hit

def keywords_for(text):
    """The keywords a verse carries, approximated from the concept lexicon."""
    ks = []
    for c in concepts:
        for kw in c['keywords']:
            if kw and kw in text:
                ks.append(kw)
                break
    return ks

print('\nbuilding keyword index...')
t0 = time.time()
for p in poems:
    p['_kw'] = keywords_for(p['text'])
print(f'  {time.time()-t0:.1f}s')

# ------------------------------------------------------------ the vectors
VEC = os.environ.get('USE_VECTORS') == '1'
if VEC:
    from embed_model import BertEmbedder
    emb = BertEmbedder(MODEL_DIR)
    t0 = time.time()
    V = emb.embed(texts)
    t_enc = time.time() - t0
    print(f'\nencoded {len(texts)} verses in {t_enc:.1f}s ({t_enc/len(texts)*1000:.1f} ms/句)')
    model_mb = os.path.getsize(os.path.join(MODEL_DIR, 'model_O4.onnx')) / 1e6
    print(f'model size: {model_mb:.0f} MB')
else:
    print('\n(USE_VECTORS not set -- baseline only)')

# ------------------------------------------------------------- Q1
print('\n=== Q1  用户的描述 → 哪一句？ ===')
print(f'{"描述":16} {"当前(字面)":>12} {"词组匹配":>12} {"向量":>12}  应当指向的概念')
for desc, want in CASES:
    # baseline: the current matcher
    base_scores = [char_overlap(desc, p) for p in poems]
    base_top = int(np.argmax(base_scores))

    # better lexical matcher
    lex_scores = [phrase_overlap(desc, p) for p in poems]
    lex_top = int(np.argmax(lex_scores))

    if VEC:
        q = emb.embed([desc])[0]
        vs = V @ q
        vec_top = int(np.argmax(vs))
        vec_s = vs[vec_top]
    else:
        vec_top, vec_s = -1, 0.0

    # the concept a thoughtful reader would pick, and its keyword
    want_concept = next((c for c in concepts if c['key'] == want), None)
    want_kw = want_concept['keywords'][0] if want_concept else '?'

    print(f'{desc:16} {poems[base_top]["text"][:11]:12} '
          f'{poems[lex_top]["text"][:11]:12} '
          f'{poems[vec_top]["text"][:11] if VEC else "-":12}  '
          f'{want_concept["name"] if want_concept else "?"} ({want_kw})')

# ------------------------------------------------------------- Q2
print('\n=== Q2  语义轴能不能因此有区分度？ ===')
print('当前语义轴在前 50 名上的标准差是 0.00（常数）。')
print('向量能否给出一个连续的相关度？')
if VEC:
    for desc, _ in CASES[:3]:
        q = emb.embed([desc])[0]
        vs = V @ q
        top = np.argsort(-vs)[:50]
        print(f'  「{desc}」 前 50 名向量余弦：均值 {vs[top].mean():.3f} '
              f'标准差 {vs[top].std():.3f} 取值 {len(set(np.round(vs[top],4)))} 种')

# ------------------------------------------------------------- Q3
print('\n=== Q3  代价 ===')
if VEC:
    mb = os.path.getsize(os.path.join(MODEL_DIR, 'model_O4.onnx')) / 1e6
    print(f'  模型体积        {mb:.0f} MB  （当前二进制 {os.path.getsize(ROOT+"/bin/sumate")/1e6:.0f} MB）')
    print(f'  编码速度        {t_enc/len(texts)*1000:.0f} ms/句（CPU，单核）')
    print(f'  全库一次性      {t_enc:.0f} s  →  之后 304 句只需查表')
    print(f'  可审计性        ✗ 权重不可读；「为什么这句更像」无法像现在这样逐条说明')
    print(f'  离线可复算      ✗ 需要 200MB 外部模型；词库改一个字不必重训，但换模型要重编码全库')
