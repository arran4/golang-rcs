#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="3491-rcsclean-u-unlock.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# 1. Setup
printf 'Initial content\n' > file.txt
ci -q -i -t-desc -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt

# 2. Save initial RCS state (input.txt,v)
# We want input.txt,v to be the state BEFORE the test command.
cp file.txt,v input.txt,v

# 3. Save working file state (input.txt)
# We want input.txt to be the state BEFORE the test command.
cp file.txt input.txt

# 4. Run the test command
# This simulates what we expect the system under test to do.
rcsclean -q -u file.txt

# 5. Generate txtar
cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs clean -u (unlock)

-- options.conf --
{"args": ["-q", "-u", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs clean

-- expected.txt,v --
$(cat file.txt,v)
EOF
