#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="89-rcs-erase.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

echo "Content" > file.txt
ci -q -i -t- -m"initial" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

# Setup initial state with logins
rcs -ajules,martha file.txt
cp file.txt,v input.txt,v

# Run rcs -e to erase one login
rcs -ejules file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs erase login

-- options.conf --
{"args": ["-ejules", "input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
