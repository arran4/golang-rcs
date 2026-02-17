#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="20-rcs-strict.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf "LOCKME\n" > file.txt

# Create initial RCS file with non-strict locking
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# Save input state
cp file.txt,v input.txt,v

# Run the command to be tested
rcs -L file.txt

# Save output state
cp file.txt,v expected.txt,v

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs set strict locking

-- options.conf --
{"args": ["-L","input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat expected.txt,v)
EOF
