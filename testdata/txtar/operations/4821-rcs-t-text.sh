#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/4821-rcs-t-text.txtar"
TMP_OUT="$OUTDIR/.4821-rcs-t-text.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "content\n" > file.txt

# Initial checkin with description "initial"
ci -q -t-"initial" -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

cp file.txt,v input.txt,v

# Mutate: change description to "new description"
rcs -t-"new description" file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
rcs -t-TEXT replaces description

-- options.conf --
{"args": ["-t-new description","input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF

mv -f "$TMP_OUT" "$OUT"
