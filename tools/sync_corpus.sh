#!/bin/sh
# Copy the canonical corpus from data/ into the embeddable package directory.
set -e
cd "$(dirname "$0")/.."
cp data/dongpo_corpus.json internal/corpus/data/dongpo_corpus.json
echo "synced data/dongpo_corpus.json -> internal/corpus/data/dongpo_corpus.json"
