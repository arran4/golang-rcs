#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="7295-rcs-t-file.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf 'Initial content\n' > file.txt
ci -q -i -minitial -wtester -d'2020-01-01 00:00:00Z' -t-"Old description" file.txt </dev/null
co -q -l file.txt

cp file.txt,v input.txt,v
cp file.txt input.txt

printf 'Description from file\n' > desc.txt
rcs -tdesc.txt file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -tFILE flag (replace description from file)

-- options.conf --
{"args": ["-tdesc.txt", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- desc.txt --
$(cat desc.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
