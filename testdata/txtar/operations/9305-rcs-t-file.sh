#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/9305-rcs-t-file.txtar"
TMP_OUT="$OUTDIR/.9305-rcs-t-file.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "content\n" > file.txt

# Initial checkin
ci -q -t-"initial" -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

cp file.txt,v input.txt,v

echo "file description" > desc.txt

# Mutate: change description using file
rcs -tdesc.txt file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
rcs -tFILE replaces description

-- options.conf --
{"args": ["-tdesc.txt","input.txt"] }

-- desc.txt --
$(cat desc.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF

mv -f "$TMP_OUT" "$OUT"
