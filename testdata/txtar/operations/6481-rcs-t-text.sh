#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="6481-rcs-t-text.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf 'Initial content\n' > file.txt
ci -q -i -minitial -wtester -d'2020-01-01 00:00:00Z' -t-"Old description" file.txt </dev/null
co -q -l file.txt

cp file.txt,v input.txt,v
cp file.txt input.txt

rcs -t-"New description text" file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -t-TEXT flag (replace description)

-- options.conf --
{"args": ["-t-New description text", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
