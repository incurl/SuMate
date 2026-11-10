#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Audit the concept lexicon for *modern* vocabulary.

The vectorisation experiment (docs/EVAL_向量化是否有用.md) diagnosed the
free-text failure as a vocabulary gap, not a semantic one: all 936 trigger
words are classical, and none of the words a modern reader actually types
(累 / 困 / 想家 / 难过 ...) appears in the lexicon.

This script makes that diagnosis checkable without any model.

Usage:
    python3 tools/audit_modern_vocab.py
"""
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CJK = re.compile(r'[㐀-䶿一-鿿]')

# Words a reader might type when describing a state.  Grouped by the state
# they name.
MODERN = {
    "疲惫":   ["累", "疲劳", "疲", "倦", "疲乏", "没劲", "无力", "歇", "不想动"],
    "孤独":   ["孤独", "一个人", "独自", "寂寞", "没人", "自己待着", "落单"],
    "想家":   ["想家", "回家", "老家", "故乡", "思念", "念家"],
    "低落":   ["难过", "低落", "郁闷", "emo", "不开心", "心沉", "堵得慌"],
    "开心":   ["开心", "高兴", "快乐", "愉快", "舒心", "轻快", "得意"],
    "迷茫":   ["迷茫", "不知道怎么办", "没方向", "茫然", "纠结", "该何去何从"],
    "释然":   ["释然", "放下", "想通", "算了", "接受", "认了"],
    "被压着": ["压力", "喘不过气", "扛不住", "撑不住", "被逼", "窒息"],
    "平淡":   ["平静", "安静", "平常", "淡", "没意思", "一般"],
}


def load_keywords():
    """Collect every trigger word the lexicon actually uses.

    The lexicon is assembled from three places: the two concept tables and
    the modern colloquial table, so all three are read.
    """
    src = ""
    for f in ("knowledge.go", "knowledge_extra.go", "modern.go"):
        p = os.path.join(ROOT, "internal", "curated", f)
        if os.path.exists(p):
            src += open(p, encoding="utf-8").read()
    kws = set()
    for m in re.finditer(r'Keywords:\s*\[\]string\{([^}]*)\}', src):
        kws.update(re.findall(r'"([^"]+)"', m.group(1)))
    # The modern table is a map literal, not a Keywords: field.
    for m in re.finditer(r'"([a-z][a-z0-9_.]+)":\s*\{(.*?)\}', src, re.S):
        if m.group(1) in ("modernTriggers",):
            continue
        kws.update(re.findall(r'"([^"]+)"', m.group(2)))
    return kws


def main():
    kws = load_keywords()
    if not kws:
        print("no keywords found -- is internal/curated/knowledge.go present?")
        return 1

    singles = [k for k in kws if len(k) == 1]
    print(f"触发词 {len(kws)} 条（其中单字 {len(singles)} 条，多字 {len(kws)-len(singles)} 条）")
    print()

    total = hit = 0
    for state, words in MODERN.items():
        got = [w for w in words if w in kws]
        total += len(words)
        hit += len(got)
        mark = "✓" if got else "✗"
        print(f"  {mark} {state:6} 命中 {len(got):2}/{len(words):2}  {got if got else '（全无）'}")
    print()
    print(f"  现代口语在触发词表中的命中率：{hit}/{total} "
          f"({100.0*hit/total:.0f}%)")
    if hit < total:
        print()
        print("  结论：缺口是**词汇**的，不是语义的。")
        print("  补上这些词，比引入一个 200 MB 的向量模型便宜得多，")
        print("  而且仍然可逐条解释。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
