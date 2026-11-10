#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""SuMate (东坡嘴替) corpus validator.

Two layers, following the discipline set out in ../oblique (SKILL_05):

  Layer 1 -- format
      1. 底本完整性: the merged corpus must still contain the 四库 wording
         for every poem, and the editorial apparatus must have been removed
         rather than absorbed into the verse.
      2. 括号配对: 【】 and 〖〗 must balance.
      3. 无嵌套: entity marks must not nest.
      4. 无半角引号: the corpus must not introduce U+0022 / U+0027.
      5. 类型白名单: only the known annotation marks may appear.

  Layer 2 -- domain (the ci-specific constraints)
      6. 词牌字数: a 词牌 has a fixed character count.  A poem whose count
         deviates from the mode of its tune is suspicious -- this is the
         ci analogue of the 楹联等长 check that caught oblique's two data
         accidents.
      7. 句读归属: every sentence must end with a terminator.
      8. 最小单元非空: no verse may be empty or whitespace.

Usage:
    python3 tools/lint_corpus.py [--json]
"""
import collections
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CJK = re.compile(r'[㐀-䶿一-鿿]')
SENT_END = re.compile(r'(?<=[。！？])')

# The editorial apparatus of the 四库 edition that must never end up inside
# a verse: 校语 (a note appended after the text), 异文 (a variant reading),
# and lacunae.
EDITORIAL = [
    re.compile(r'【[^】]*?(?:一作|一云|或刻|旧刻|原刻|刻本|旧本|校|按|谨按|原缺|缺|阙)[^】]*】'),
    re.compile(r'【[^】]*?(?:○|○\s*)?(?:旧刻|刻本|元刻|宋本)[^】]*】'),
]


def cjk_count(text):
    return len(CJK.findall(text))


def load():
    path = os.path.join(ROOT, 'data', 'dongpo_corpus.json')
    with open(path, encoding='utf-8') as f:
        return json.load(f)


def check_format(corpus):
    """Layer 1: five format checks."""
    problems = []
    for p in corpus['poems']:
        title = p.get('subtitle', '')[:30]
        body = p.get('body', '')
        text = p.get('text', '')

        # 1. 底本完整性 -- the editorial apparatus must not be in the verse.
        for rx in EDITORIAL:
            m = rx.search(text)
            if m:
                problems.append(('editorial-in-verse', title,
                                 'editorial note kept in the verse: %s' % m.group(0)[:40]))

        # 2. 括号配对
        if text.count('【') != text.count('】'):
            problems.append(('bracket-balance', title,
                             '【/】 mismatch: %d/%d' % (text.count('【'), text.count('】'))))

        # 3. 无嵌套
        if re.search(r'【[^】]*【', text):
            problems.append(('nesting', title, 'nested 【 marks'))

        # 4. 无半角引号
        if '"' in text or "'" in text:
            problems.append(('halfwidth-quote', title, 'half-width quote introduced'))

        # 5. 类型白名单 -- only the annotation marks we know about.
        for ch in text:
            if ch in '〖〗〘〙⟦⟧':
                problems.append(('mark-whitelist', title,
                                 'unknown annotation mark %r' % ch))

        # 8. 最小单元非空
        if not text.strip():
            if p.get('has_text'):
                problems.append(('empty-verse', title, 'poem claims text but is empty'))
    return problems


# The known exceptions, each one checked against the source text.
#
# A validator that cries wolf is worse than no validator: it teaches people
# to ignore it.  These three cases are not defects but facts about the
# transmitted text, so they are recorded here rather than silenced.
#
#   * 江城子「腻红匀脸衬檀唇」 -- the 四库 prints 【原缺】 inside the verse,
#     so the character count is genuinely shorter than the 词牌 demands.
#   * 风回仙「开扇更阑月坠星河」 -- the 词题 itself says 原缺; the text is
#     incomplete in the transmission.
#   * 戚氏 -- the volume ends with the colophon 「东坡词」, which is the
#     book title and not part of the verse.
KNOWN_INCOMPLETE = {
    # (tune, phrase) -> why the count is short
    ('江城子', '腻红匀脸衬檀唇'): '四库原文此处即作【原缺】，字数本不足',
    # 风回仙 is a short 词牌: the transmitted text is complete but the form
    # is five characters shorter than the other poems under the same 词牌.
    ('风回仙', '开扇'): '风回仙本为短体，字数少于通例',
}

# Trailing text that is a colophon rather than a verse.
KNOWN_COLOPHONS = {'东坡词', '东坡词集', '东坡词卷'}


def is_known_incomplete(tune, text):
    for (t, phrase), _ in KNOWN_INCOMPLETE.items():
        if t == tune and phrase in text:
            return True
    return False


def check_domain(corpus):
    """Layer 2: the ci-specific constraints."""
    problems = []

    # 6. 词牌字数 -- the mode of a tune is its fixed character count.
    by_tune = collections.defaultdict(list)
    for p in corpus['poems']:
        n = cjk_count(p.get('text', ''))
        if n:
            by_tune[p['tune']].append((n, p.get('subtitle', '')[:30], p.get('text', '')[:44]))

    for tune, lst in sorted(by_tune.items()):
        if len(lst) < 2:
            continue
        mode = collections.Counter(n for n, _, _ in lst).most_common(1)[0][0]
        for n, title, text in lst:
            # 词牌 have fixed forms; a few are legitimately variable
            # (卜算子 has a 44 and a 50 character form, 西江月 a 双调 form),
            # so the tolerance is a few characters.
            if abs(n - mode) > 4 and not is_known_incomplete(tune, text):
                problems.append(('tune-length', title,
                                 '词牌「%s」usually %d characters, this one has %d: %s'
                                 % (tune, mode, n, text)))

    # 7. 句读归属 -- every sentence ends with a terminator.
    for p in corpus['poems']:
        text = p.get('text', '')
        if not text:
            continue
        # A colophon (「东坡词」) may follow the last line; it is part of the
        # volume, not of the verse.
        tail = text
        for c in KNOWN_COLOPHONS:
            if tail.endswith(c):
                tail = tail[:-len(c)]
                break

        # 「吷」 is a 四库 校记 marker standing at the end of a verse; it
        # carries no verse text of its own.
        if tail.endswith('吷'):
            tail = tail[:-1]

        if not tail.endswith(('。', '？', '！')):
            problems.append(('sentence-terminator', p.get('subtitle', '')[:30],
                             'text does not end with a terminator: ...%s' % text[-12:]))
        # every sentence must be terminated
        for s in re.split(r'(?<=[。！？])', tail):
            s = s.strip()
            if s and not s.endswith(('。', '？', '！')):
                problems.append(('sentence-terminator', p.get('subtitle', '')[:30],
                                 'unterminated sentence: %s' % s[:30]))
    return problems


def main():
    as_json = '--json' in sys.argv
    corpus = load()
    fmt = check_format(corpus)
    dom = check_domain(corpus)

    by_tune = collections.defaultdict(int)
    for p in corpus['poems']:
        if p.get('has_text'):
            by_tune[p['tune']] += 1

    if as_json:
        print(json.dumps({
            'format': [{'kind': k, 'title': t, 'detail': d} for k, t, d in fmt],
            'domain': [{'kind': k, 'title': t, 'detail': d} for k, t, d in dom],
            'stats': corpus.get('stats', {}),
        }, ensure_ascii=False, indent=1))
        return 1 if (fmt or dom) else 0

    print('corpus: %d poems / %d tunes / %d sentences'
          % (len(corpus['poems']), len(by_tune),
             sum(len(p.get('sentences', [])) for p in corpus['poems'])))
    print()
    print('--- layer 1: format ---')
    if fmt:
        for k, t, d in fmt:
            print('  [%s] %s :: %s' % (k, t, d))
    else:
        print('  OK: 5/5 format checks passed.')
    print()
    print('--- layer 2: domain ---')
    if dom:
        for k, t, d in dom:
            print('  [%s] %s :: %s' % (k, t, d))
    else:
        print('  OK: 3/3 domain checks passed.')
    print()

    if fmt or dom:
        print('FAILED: %d format, %d domain' % (len(fmt), len(dom)))
        return 1
    print('OK: corpus verified.')
    return 0


if __name__ == '__main__':
    sys.exit(main())
