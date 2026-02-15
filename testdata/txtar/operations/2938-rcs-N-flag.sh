#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="2938-rcs-N-flag.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# 1.1
printf 'Initial content\n' > file.txt
ci -q -i -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt

# 1.2
printf 'Modified content\n' > file.txt
ci -q -mmodified -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# Assign TAG to 1.1
rcs -nTAG:1.1 file.txt

cp file.txt,v input.txt,v
co -q -p -r1.2 file.txt > input.txt

# Overwrite/Move TAG to 1.2
rcs -NTAG:1.2 file.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -N flag (overwrite symbolic name)

-- options.conf --
{"args": ["-NTAG:1.2", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
