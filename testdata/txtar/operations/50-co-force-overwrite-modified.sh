#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/50-co-force-overwrite-modified.txtar"
TMP_OUT="$OUTDIR/.50-co-force-overwrite-modified.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "v1\n" > file.txt

# setup: create 1.1
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# prepare modified working file
chmod u+w file.txt
printf "modified\n" > file.txt

# save input state
cp file.txt   input.txt
cp file.txt,v input.txt,v

# execution: -f forces overwrite
co -q -f file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
co -f forces overwrite of modified working file

-- options.conf --
{"args": ["-q","-f","input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
co

-- expected.txt --
$(cat file.txt)
EOF

mv -f "$TMP_OUT" "$OUT"
