#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="9412-rcs-o-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# 1.1
printf 'Initial content\n' > file.txt
ci -q -i -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt

# 1.2
printf 'Modified content\n' > file.txt
# Use -u to keep the working file (unlocked, read-only is fine)
ci -q -u -mmodified -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

cp file.txt,v input.txt,v
cp file.txt input.txt

# Delete 1.2
rcs -o1.2 file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -o flag (delete revision)

-- options.conf --
{"args": ["-o1.2", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
# rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
