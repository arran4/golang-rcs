#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="3572-rcs-m-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf 'Initial content\n' > file.txt
ci -q -i -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt

cp file.txt,v input.txt,v
cp file.txt input.txt

rcs -m1.1:"new message" file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -m flag (change log message)

-- options.conf --
{"args": ["-m1.1:new message", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
