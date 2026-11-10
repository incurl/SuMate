#!/bin/sh
# Copy the canonical corpus from data/ into the embeddable package directory.
#
# data/dongpo_corpus.json is the single source of truth; the copy under
# internal/corpus/data/ exists only because go:embed cannot reach a parent
# directory.  The two must never drift apart, so the build checks them.
set -e
cd "$(dirname "$0")/.."

if [ ! -f data/dongpo_corpus.json ]; then
    echo "data/dongpo_corpus.json is missing; run tools/build_corpus.py first" >&2
    exit 1
fi

cp data/dongpo_corpus.json internal/corpus/data/dongpo_corpus.json
echo "synced data/dongpo_corpus.json -> internal/corpus/data/dongpo_corpus.json"
