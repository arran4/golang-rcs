#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="7654-rcs-s-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf 'Initial content\n' > file.txt
ci -q -i -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt

cp file.txt,v input.txt,v
cp file.txt input.txt

rcs -sRel:1.1 file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -s flag (change state)

-- options.conf --
{"args": ["-sRel:1.1", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
