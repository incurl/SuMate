#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""SuMate (东坡嘴替) corpus builder.

Merges two sources into one canonical corpus JSON:

  A) 东坡词.txt -- 钦定四库全书 edition of 《东坡词》.
     Provides: tune (词牌), subtitle (词题/小序, usually carrying the exact
     year, place and occasion) and the authoritative 四库 body text, with
     ideographic-space stanza (片) boundaries preserved.

  B) chinese-poetry `宋词/ci.db` (MIT, JackeyGao) -- modern punctuation and
     sentence segmentation for the same poems.

The two editions differ in glyph variants (収/收, 囘/回, 叟/搜 ...), so the
merge aligns the two strings character-wise and *transfers* the punctuation
onto the 四库 wording, which stays the authority for the final text.

Output: data/dongpo_corpus.json
"""
import collections
import difflib
import json
import os
import re
import sqlite3

try:
    import zhconv
except ImportError:          # pragma: no cover - optional dependency
    zhconv = None

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
IDSP = '\u3000'              # ideographic space (indentation + stanza marker)
CJK = re.compile(r'[\u4e00-\u9fff]')
SENT_END = re.compile(r'(?<=[。！？])')


# --------------------------------------------------------------------------
# Source A: 四库全书《东坡词》
# --------------------------------------------------------------------------
# 词牌 names that appear in the 四库《东坡词》.  The list is used to
# recognise headers unambiguously, because a 词题 can also be a short
# all-CJK string (e.g. `【送别】`, `【春情】`).
SIKU_TUNES = [
    "阳关曲", "如梦令", "生查子", "昭君怨", "点绛唇", "减字木兰花",
    "诉衷情", "菩萨鬘", "天怜豪俊", "风回仙", "采桑子", "卜算子",
    "好事近", "华清引", "谒金门", "清平乐", "更漏子", "乌夜啼",
    "阮郎归", "虞美人影", "西江月", "少年游", "瑶池燕", "南柯子",
    "浪淘沙", "鹧鸪天", "玉楼春", "南乡子", "寒玉细凝肤", "怅望送春杯",
    "何处倚阑干", "鹊桥仙", "瑞鹧鸪", "虞美人", "醉落魄", "临江仙",
    "蝶恋花", "荷华媚", "渔家傲", "十拍子", "苏幙遮", "行香子",
    "青玉案", "殢人娇", "江城子", "千秋岁", "河满子", "祝英台近",
    "皂罗特髻", "洞仙歌", "劝金船", "意难忘", "满江红", "满庭芳",
    "水调歌头", "雨中花慢", "八声甘州", "醉蓬莱", "三部乐", "念奴娇",
    "水龙吟", "归朝欢", "永遇乐", "贺新郎", "稍遍", "戚氏",
    # 词牌 that appear only *inline* inside a 四库 body line
    "浣溪沙", "定风波", "江城梅花引", "踏莎行",
]
TUNE_SET = set(SIKU_TUNES)
# longest first so that `减字木兰花` wins over a hypothetical `木兰花`
TUNE_SORTED = sorted(TUNE_SET, key=len, reverse=True)


def find_header(text, start=0):
    """Find the next poem header in `text` at or after `start`.

    Returns (tune, sub, body_end, header_end) where `tune` is '' for an
    `又【词题】` continuation, `body_end` is where the preceding poem's body
    ends and `header_end` is just past the closing bracket.

    A header is `词牌【词题】` or `又【词题】`.  Because the 词牌 names are
    known, the match is exact: we look for a tune name that is immediately
    followed by `【` and that starts on a boundary (start of text, ideographic
    space, or punctuation).
    """
    n = len(text)
    best = None
    for tune in TUNE_SORTED:
        idx = text.find(tune + '\u3010', start)
        while idx != -1:
            # The 四库 sometimes runs a header straight on from the previous
            # body text (e.g. `...芳草迷归路浣溪沙【新秋...】`), so any
            # position is allowed as long as the tune name is exact.
            if True:
                close = text.find('\u3011', idx + len(tune))
                if close != -1:
                    cand = (tune, text[idx + len(tune) + 1:close], idx, close + 1)
                    if best is None or cand[2] < best[2]:
                        best = cand
            idx = text.find(tune + '\u3010', idx + 1)
    if best is not None:
        return best

    # `又【词题】` continuation.  The 四库 runs the next 词题 straight on
    # from the preceding verse, so any position is allowed.
    idx = text.find('\u53c8\u3010', start)
    while idx != -1:
        close = text.find('\u3011', idx + 1)
        if close != -1:
            return ('', text[idx + 2:close], idx, close + 1)
        idx = text.find('\u53c8\u3010', idx + 1)
    return None


def _starts_with_quote(text, pos):
    """True if `text[pos:]` starts with `词牌【作者】` (a 集句 quote marker)."""
    i = pos
    n = len(text)
    while i < n and CJK.match(text[i]) and i - pos < 12:
        i += 1
    if i == pos or i >= n or text[i] != '\u3010':
        return False
    j = i + 1
    while j < n and CJK.match(text[j]) and j - i <= 4:
        j += 1
    return (j > i + 1 and j < n and text[j] == '\u3011')


def parse_siku(path):
    """Parse the 四库《东坡词》 text into poem entries.

    Handled entry shapes:
      * `词牌【词题】`          -- new tune
      * `又【词题】` / `又`      -- continuation under the current tune
      * `词牌【作者】句...`      -- 集句 (anthology): the bracket names the
                                    quoted author, so the rest is content
      * `词牌【原缺】...【缺】`  -- lacunae in the transmitted text
    """
    lines = [l for l in open(path, encoding='utf-8').read().split('\n') if l.strip()]
    start = None
    for i, l in enumerate(lines):
        s = l.strip()
        if s.startswith('宋') and '苏轼' in s and '撰' in s:
            start = i + 1
            break
    if start is None:
        raise SystemExit('cannot locate body start in %s' % path)

    entries = []
    cur = None
    curtune = None

    def open_entry(tune, sub, body='', kind='normal'):
        e = {'tune': tune, 'sub': sub, 'body': body, 'kind': kind}
        entries.append(e)
        return e

    for raw in lines[start:]:
        # Keep the leading ideographic space: it separates a header from the
        # preceding stanza when a header appears inline.
        text = raw.strip()
        if text.lstrip(IDSP) == '又':
            cur = open_entry(curtune, '')
            continue

        pos = 0
        while pos < len(text):
            hdr = find_header(text, pos)
            if hdr is None:
                if cur is not None:
                    cur['body'] += text[pos:]
                break
            tune, sub, body_end, header_end = hdr
            if cur is not None:
                # A pending 集句 entry (its 词题 is 【集句】) is completed by
                # a line that starts with `词牌【作者】` -- the quoted
                # author's name.  That line is content, not a header.
                if (cur['sub'] == '集句' and not cur['body']
                        and _starts_with_quote(text, pos)):
                    cur['body'] += text[pos:]
                    pos = len(text)
                    break
                cur['body'] += text[pos:body_end]
            if tune:
                curtune = tune
            if sub.startswith('原缺】') or sub.startswith('缺】'):
                cur = open_entry(curtune, '原缺', kind='lacuna')
            else:
                cur = open_entry(curtune, sub)
            pos = header_end

    for e in entries:
        # Editorial brackets that belong to the 四库 apparatus rather than to
        # the verse itself: 【阙】/【原缺】 gaps, and 【一作...】 variants.
        e['lacunae'] = len(re.findall(r'\u3010(?:\u539f)?\u7f3a\u3011', e['body']))
        e['body'] = re.sub(r'\u3010(?:\u539f)?\u7f3a\u3011', '', e['body'])
        e['body'] = re.sub(r'\u3010\u4e00\u4f5c[^\u3011]*\u3011', '', e['body'])
        e['body'] = e['body'].replace('\u3000', '')

        # 集句 (anthology poem): the body quotes other poets and marks each
        # quoted line with 【作者】.  Detect it from those markers rather than
        # from the 词题, because a 词题 such as 【赤壁怀古】 is also a short
        # all-CJK string.
        quotes = re.findall(r'【([\u4e00-\u9fff]{2,4})】', e['body'])
        if quotes:
            e['kind'] = 'jiju'
            e['quotes'] = quotes
            e['body'] = re.sub(r'【[\u4e00-\u9fff]{2,4}】', '', e['body'])
            e['sub'] = '集句'
        e['body'] = e['body'].strip()
        e['stanzas'] = split_stanzas(e['body'])
        e['norm'] = ''.join(CJK.findall(e['body']))
    return entries


def split_stanzas(body):
    """Split a poem body into 片 (stanzas) on ideographic spaces."""
    parts = [p for p in body.split(IDSP)]
    out = []
    for p in parts:
        p = p.strip()
        if p:
            out.append(p)
    return out


# --------------------------------------------------------------------------
# Source B: chinese-poetry ci.db
# --------------------------------------------------------------------------
def parse_cidb(path):
    con = sqlite3.connect(path)
    db = []
    for rhythmic, author, content in con.execute(
            "select rhythmic, author, content from ci where author like '%苏轼%'"):
        text = content.split(' >> ')[0]
        text = re.sub(r'\s+', '', text)
        sents = [s for s in SENT_END.split(text) if s]
        db.append({
            'tune': rhythmic,
            'sent': sents,
            'norm': ''.join(CJK.findall(text)),
        })
    con.close()
    return db


# --------------------------------------------------------------------------
# Alignment helpers
# --------------------------------------------------------------------------
def bigrams(s):
    return collections.Counter(s[i:i + 2] for i in range(len(s) - 1))


def jaccard(a, b):
    A, B = bigrams(a), bigrams(b)
    inter = sum((A & B).values())
    union = sum(A.values()) + sum(B.values()) - inter
    return inter / union if union else 0.0


def similarity(a, b):
    if not a or not b:
        return 0.0
    sm = difflib.SequenceMatcher(None, a, b).ratio()
    ja = jaccard(a, b)
    score = max(sm, ja)
    if a in b or b in a:
        score += 0.12
    return min(score, 1.0)


def transfer_punctuation(siku_text, db_sents):
    """Project the punctuation of `db_sents` onto the 四库 wording.

    The two editions use variant glyphs (収/收, 囘/回 ...), so we align them
    character-wise with difflib opcodes, map every db character onto its
    四库 counterpart, and then insert each sentence terminator at the
    matching 四库 position.

    Returns the 四库 text with modern punctuation inserted.
    """
    siku_cjk_pos = [i for i, ch in enumerate(siku_text) if CJK.match(ch)]
    siku_norm = ''.join(siku_text[i] for i in siku_cjk_pos)

    db_sents = [s for s in db_sents if s]
    db_norm = ''.join(CJK.findall(''.join(db_sents)))

    if not siku_norm or not db_norm:
        return siku_text

    # db_norm index -> siku_norm index
    db_to_siku = {}
    sm = difflib.SequenceMatcher(None, siku_norm, db_norm, autojunk=False)
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag == 'equal':
            for k in range(i2 - i1):
                db_to_siku[j1 + k] = i1 + k
        elif tag == 'replace':
            # map positionally inside the replaced block
            span = i2 - i1
            if span > 0:
                for k in range(j2 - j1):
                    db_to_siku[j1 + k] = min(i1 + k, i2 - 1)
        elif tag == 'insert':
            for k in range(j1, j2):
                db_to_siku[k] = min(i1, len(siku_norm) - 1)

    # locate every db sentence inside db_norm and remember its terminator
    marks = {}          # siku_norm index -> terminator char
    cursor = 0
    for sent in db_sents:
        s_norm = ''.join(CJK.findall(sent))
        if not s_norm:
            continue
        term = ''
        for ch in sent:
            if ch in '。！？':
                term = ch
                break
        if not term:
            continue
        pos = db_norm.find(s_norm, cursor)
        if pos < 0:
            continue
        cursor = pos + len(s_norm)
        # the db index of the last CJK char of this sentence
        last_db = pos + len(s_norm) - 1
        siku_idx = db_to_siku.get(last_db)
        if siku_idx is None:
            continue
        marks[siku_idx] = term

    # simple two-pass: collect term per original index
    term_at = {}
    for siku_idx, term in marks.items():
        if siku_idx < len(siku_cjk_pos):
            term_at[siku_cjk_pos[siku_idx]] = term
    rebuilt = []
    for i, ch in enumerate(siku_text):
        rebuilt.append(ch)
        if i in term_at:
            rebuilt.append(term_at[i])
    return ''.join(rebuilt)


# 四库 glyph variants that the traditional/simplified mapping does not cover
# but which modern editions print differently.
VARIANTS = {
    # 四库异体字 / 俗字，改为现代通行正字
    '乆': '久', '囘': '回', '晩': '晚', '疎': '疏', '氷': '冰',
    '収': '收', '慿': '凭', '顔': '颜', '絶': '绝', '戸': '户',
    '髪': '发', '沉': '沈', '樽': '尊', '蕋': '蕊', '鬬': '斗',
    '竒': '奇', '葢': '盖', '嬾': '懒', '顦': '憔', '顇': '悴',
    '疉': '叠', '剪': '翦', '辜': '孤', '谩': '漫', '纒': '缠',
    '座': '坐', '邱': '丘', '栏': '阑', '懐': '怀', '脩': '修',
    '朶': '朵', '鸎': '莺', '斾': '旆', '羣': '群', '牀': '床',
    '胷': '胸', '寛': '宽', '敎': '教', '夲': '本', '逺': '远',
    '椉': '乘', '昏': '醺', '氊': '毡', '蔌': '簌', '趂': '趁',
    '霄': '宵', '借': '藉', '茰': '萸', '嫋': '袅', '劒': '剑',
    '徧': '遍', '様': '样', '巵': '卮', '縁': '缘', '腕': '碗',
    '镮': '环', '濶': '阔', '懃': '勤', '娆': '饶', '厯': '历',
    '毬': '球', '鳬': '凫', '醖': '酝', '釭': '缸', '蓛': '簌',
    '着': '着', '欹': '欹', '秾': '秾', '歩': '步', '裛': '裛',
    '縠': '縠', '頳': '頳', '巻': '卷', '亸': '亸', '骎': '骎',
    '巉': '巉', '霅': '霅', '噀': '噀', '麋': '麋', '猱': '猱',
    '蒨': '茜', '檾': '苘', '擡': '抬', '眊': '眊', '矂': '搊',
    '搢': '揎', '纻': '苧', '嚬': '颦', '咍': '哈', '臞': '癯',
    '樷': '栊', '揾': '揾', '歜': '歜', '坰': '坰', '鞓': '鞯',
    '甃': '甃', '莳': '莳', '羨': '羡', '黃': '黄', '舊': '旧',
    '盡': '尽', '臺': '台', '葉': '叶', '華': '华', '處': '处',
}


def to_simplified(text):
    """Convert a 四库 text to Simplified Chinese.

    The transmitted text is traditional; the product and the concept lexicon
    are written in Simplified Chinese, so the corpus is normalised once at
    build time.  A small table fixes the variant glyphs that the
    traditional/simplified mapping does not cover.
    """
    for a, b in VARIANTS.items():
        text = text.replace(a, b)
    if zhconv is not None:
        text = zhconv.convert(text, 'zh-cn')
    return text


def adopt_modern_spelling(siku_text, siku_body, modern_norm):
    """Rewrite the 四库 text with the modern edition's character forms.

    The two editions differ only in glyph variants (収/收, 囘/回, 防/岸).
    Aligning the 四库 characters to the modern ones lets us copy the modern
    spelling across without touching the 四库 wording or punctuation.
    """
    if not siku_text or not modern_norm:
        return siku_text
    siku_cjk_pos = [i for i, ch in enumerate(siku_text) if CJK.match(ch)]
    siku_norm = ''.join(siku_text[i] for i in siku_cjk_pos)
    if not siku_norm:
        return siku_text

    out = list(siku_text)
    sm = difflib.SequenceMatcher(None, siku_norm, modern_norm, autojunk=False)
    for tag, i1, i2, j1, j2 in sm.get_opcodes():
        if tag != 'replace':
            continue
        n = min(i2 - i1, j2 - j1)
        for k in range(n):
            si = siku_cjk_pos[i1 + k]
            if out[si] != modern_norm[j1 + k]:
                out[si] = modern_norm[j1 + k]
    return ''.join(out)


# --------------------------------------------------------------------------
def main():
    siku = parse_siku(os.path.join(ROOT, '东坡词.txt'))
    db = parse_cidb('/tmp/ci.db')
    by_tune = collections.defaultdict(list)
    for d in db:
        by_tune[d['tune']].append(d)

    poems = []
    stats = collections.Counter()
    used_db = set()
    for e in siku:
        norm = e['norm']
        if not norm:
            stats['no_text'] += 1
            poems.append({
                'tune': e['tune'], 'subtitle': e['sub'], 'body': e['body'],
                'stanzas': e['stanzas'], 'text': '', 'sentences': [],
                'align': 0.0, 'has_text': False,
            })
            continue

        cands = by_tune.get(e['tune'], []) + db
        best = max(cands, key=lambda d: similarity(norm, d['norm']))
        score = similarity(norm, best['norm'])
        text = ''
        sents = []
        if score >= 0.72:
            text = transfer_punctuation(e['body'], best['sent'])
            sents = [s for s in SENT_END.split(text) if s]
            used_db.add(id(best))
            stats['aligned'] += 1
        else:
            stats['unaligned'] += 1

        # The 四库 prints variant glyphs (防 for 岸, 戚 for 酹, ...) that the
        # traditional/simplified mapping does not cover.  When the modern
        # edition is an exact textual match, prefer its spelling.
        if best is not None:
            ratio = difflib.SequenceMatcher(None, norm, best['norm'],
                                            autojunk=False).ratio()
            if ratio >= 0.90:
                modern = best['norm']
                # keep the 四库 wording but adopt the modern character forms
                # where the two agree on the surrounding context
                text = adopt_modern_spelling(text, e['body'], modern)
                sents = [s for s in SENT_END.split(text) if s]
        poems.append({
            'tune': e['tune'],
            'subtitle': to_simplified(e['sub']),
            'body': to_simplified(e['body']),
            'stanzas': [to_simplified(st) for st in e['stanzas']],
            'text': to_simplified(text),
            'sentences': [to_simplified(x) for x in sents],
            'align': round(score, 4),
            'has_text': bool(norm),
        })

    out = os.path.join(ROOT, 'data', 'dongpo_corpus.json')
    os.makedirs(os.path.dirname(out), exist_ok=True)
    with open(out, 'w', encoding='utf-8') as f:
        json.dump({
            'source': {
                'siku': '钦定四库全书《东坡词》',
                'punctuation': 'chinese-poetry 宋词/ci.db (MIT, JackeyGao)',
            },
            'stats': {
                'poems': len(poems),
                'aligned': stats['aligned'],
                'unaligned': stats['unaligned'],
                'no_text': stats['no_text'],
            },
            'poems': poems,
        }, f, ensure_ascii=False, indent=1)
    print('poems=%d aligned=%d unaligned=%d no_text=%d -> %s'
          % (len(poems), stats['aligned'], stats['unaligned'], stats['no_text'], out))


if __name__ == '__main__':
    main()
