#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="87-rcs-nonstrict-idempotent.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

printf "UNLOCKME\n" > file.txt

# Create initial RCS file with non-strict locking
ci -q -i -u -m"r1" -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
rcs -q -U file.txt

# Save input state
cp file.txt,v input.txt,v

# Run the command to be tested (should be no-op)
rcs -U file.txt

# Save output state
cp file.txt,v expected.txt,v

cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs set non-strict locking (idempotent)

-- options.conf --
{"args": ["-U","input.txt"] }

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat expected.txt,v)
EOF
