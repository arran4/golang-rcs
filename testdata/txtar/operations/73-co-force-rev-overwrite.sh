#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/73-co-force-rev-overwrite.txtar"
TMP_OUT="$OUTDIR/.73-co-force-rev-overwrite.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "v1\n" > file.txt

# setup: create 1.1
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# modify and create 1.2
co -q -l file.txt
printf "v2\n" > file.txt
ci -q -u -m"r2" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# prepare modified working file
chmod u+w file.txt
printf "modified\n" > file.txt

# save input state
cp file.txt   input.txt
cp file.txt,v input.txt,v

# execution: -f1.1 forces overwrite with specific revision
co -q -f1.1 file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
co -f<REV> forces overwrite of modified working file with specific revision

-- options.conf --
{"args": ["-q","-f1.1","input.txt"] }

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
