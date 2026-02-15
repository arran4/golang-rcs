#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="55-rcs-append-file.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# Create source file with access list
echo "Source Content" > source.txt
ci -q -i -t- -m"initial" -wtester -d'2020-01-01 00:00:00Z' source.txt </dev/null
rcs -ajules,martha source.txt

# Create target file
echo "Target Content" > target.txt
ci -q -i -t- -m"initial" -wtester -d'2020-01-01 00:00:00Z' target.txt </dev/null

# Capture initial state of target
cp target.txt,v input.txt,v

# Run rcs -A to append access list from source.txt to target.txt
# Note: In the test environment, we might need to be careful about file paths.
# The txtar test runner usually puts files in the current directory.
# We will assume 'source.txt,v' is available.

rcs -Asource.txt target.txt

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs append access list from file

-- options.conf --
{"args": ["-Asource.txt", "input.txt"] }

-- source.txt,v --
$(cat source.txt,v)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat target.txt,v)
EOF
