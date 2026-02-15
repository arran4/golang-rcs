#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="2161-ci-s-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf 'Initial content\n' > file.txt
ci -q -i -sRel -t- -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt

# save input for test
cp file.txt,v input.txt,v
cp file.txt input.txt

# ci with -s
echo "New content" > file.txt
ci -q -sProd -mnew -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
ci -s flag (set state)

-- options.conf --
{"args": ["-sProd", "-mnew", "-wtester", "-d2020-01-02 00:00:00Z", "input.txt"]}

-- input.txt --
New content

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# ci

-- expected.txt,v --
$(cat file.txt,v)
EOF
