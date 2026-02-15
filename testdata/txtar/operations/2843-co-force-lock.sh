#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/2843-co-force-lock.txtar"
TMP_OUT="$OUTDIR/.2843-co-force-lock.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "v1\n" > file.txt
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# Modify working file
chmod u+w file.txt
printf "modified\n" > file.txt

# save input state
cp file.txt   input.txt
cp file.txt,v input.txt,v

# execution: -f -l forces overwrite and locks
co -q -f -l file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
co -f -l forces overwrite of modified working file and locks it

-- options.conf --
{"args": ["-q","-f","-l","input.txt"] }

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
co

-- expected.txt --
$(cat file.txt)

-- expected.txt,v --
$(cat file.txt,v)
EOF

mv -f "$TMP_OUT" "$OUT"
