#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/5928-rcs-comment-leader-replace.txtar"
TMP_OUT="$OUTDIR/.5928-rcs-comment-leader-replace.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "content\n" > file.txt

# setup: create 1.1 with initial comment leader #
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -c"# " file.txt

cp file.txt   input.txt
cp file.txt,v input.txt,v

# execution: replace with //
rcs -c"// " file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
rcs -c// replaces existing comment leader

-- options.conf --
{"args": ["-c// ", "input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF

mv -f "$TMP_OUT" "$OUT"
