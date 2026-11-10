#!/bin/sh
# The SuMate build pipeline: generate → validate → build.
#
# The order is fixed and not negotiable.  The corpus is generated from the
# source texts, then validated in two layers (format + domain), and only a
# corpus that passes both is embedded into the binary.  Building without
# validating would ship data that the format checks cannot vouch for.
#
#   tools/build_corpus.py   生成   merge the 四库 text with the punctuated 宋词
#   tools/lint_corpus.py    校验   format + domain checks
#   go test ./internal/...  校验   the same constraints as Go assertions
#   wails3 build            构建   the desktop application
set -e
cd "$(dirname "$0")/.."

STEP=${1:-all}

run_corpus() {
    echo "==> 生成 corpus"
    if [ -f /tmp/ci.db ]; then
        python3 tools/build_corpus.py
    else
        echo "    (跳过：/tmp/ci.db 不存在，保留现有 corpus)"
    fi
    ./tools/sync_corpus.sh
}

run_lint() {
    echo "==> 校验 corpus (format + domain)"
    python3 tools/lint_corpus.py
}

run_gotest() {
    echo "==> 校验 Go 断言"
    go test ./internal/...
}

run_audit() {
    echo "==> 审计匹配引擎（各轴是否真的在区分）"
    go run ./cmd/audit
}

run_frontend() {
    echo "==> 校验前端"
    ( cd frontend && npm run typecheck && npm run test )
}

run_build() {
    echo "==> 构建桌面应用"
    if command -v wails3 >/dev/null 2>&1; then
        wails3 build
    else
        go build -tags production -o bin/sumate .
        echo "    (wails3 未安装，已用 go build 构建)"
    fi
}

case "$STEP" in
    corpus)  run_corpus ;;
    lint)    run_lint ;;
    audit)   run_audit ;;
    gotest)  run_gotest ;;
    frontend) run_frontend ;;
    app)     run_build ;;
    all)
        run_corpus
        run_lint
        run_gotest
        run_audit
        run_frontend
        run_build
        ;;
    *)
        echo "usage: $0 [all|corpus|lint|gotest|audit|frontend|app]" >&2
        exit 2
        ;;
esac

echo "==> 完成"
