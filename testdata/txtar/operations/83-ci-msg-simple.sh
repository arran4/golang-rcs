#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/83-ci-msg-simple.txtar"
TMP_OUT="$OUTDIR/.83-ci-msg-simple.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

cat > file.txt <<'EOF'
This is a simple file.
EOF

# execution
ci -q -i -u -m"Simple message" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

cat > "$TMP_OUT" <<EOF
-- description.txt --
ci checkin with simple message

-- options.conf --
{"args": ["-q","-i","-u","-m","Simple message","-wtester","-d","2020-01-01 00:00:00Z","input.txt"] }

-- input.txt --
This is a simple file.

-- tests.txt --
ci

-- expected.txt,v --
$(cat file.txt,v)
EOF

mv -f "$TMP_OUT" "$OUT"
