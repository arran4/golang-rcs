#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="2165-rlog-s-flag-multiple.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf 'Initial content\n' > file.txt
ci -q -i -sRel -t- -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt
echo "New content" > file.txt
ci -q -sProd -mnew -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null
co -q -l file.txt
echo "New content 2" > file.txt
ci -q -sExp -mnew2 -wtester -d'2020-01-03 00:00:00Z' file.txt </dev/null

cp file.txt,v input.txt,v

# Capture output of rlog -sRel,Exp
rlog -sRel,Exp file.txt | sed 's/file.txt/input.txt/g' > expected.out || true

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rlog -s flag (filter by multiple states)

-- options.conf --
{"args": ["-sRel,Exp", "input.txt"]}

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# rlog

-- expected.stdout --
$(cat expected.out)
EOF
