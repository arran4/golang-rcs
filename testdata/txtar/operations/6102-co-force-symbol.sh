#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$OUTDIR/6102-co-force-symbol.txtar"
TMP_OUT="$OUTDIR/.6102-co-force-symbol.txtar.tmp.$$"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp" "$TMP_OUT"' EXIT
cd "$tmp"

printf "v1\n" > file.txt
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# Create v2
co -q -l file.txt
printf "v2\n" > file.txt
ci -q -u -m"r2" -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# Tag v1
rcs -nTAG:1.1 file.txt

# Modify working file
chmod u+w file.txt
printf "modified\n" > file.txt

# save input state
cp file.txt   input.txt
cp file.txt,v input.txt,v

# execution: -fTAG forces overwrite
co -q -fTAG file.txt

cat > "$TMP_OUT" <<EOF
-- description.txt --
co -fTAG forces overwrite of modified working file with symbolic revision

-- options.conf --
{"args": ["-q","-fTAG","input.txt"] }

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
