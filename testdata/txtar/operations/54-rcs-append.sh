#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="54-rcs-append.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

echo "Content" > file.txt
ci -q -i -t- -m"initial" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null

# Capture initial state as input
cp file.txt,v input.txt,v

# Run rcs -a to append logins
rcs -ajules,martha file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs append logins

-- options.conf --
{"args": ["-ajules,martha", "input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
