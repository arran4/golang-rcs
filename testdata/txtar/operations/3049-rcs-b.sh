#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="3049-rcs-b.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# 1. Setup
printf 'Initial content\n' > file.txt
ci -q -t-desc -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
co -q -l file.txt
# Create a branch revision
printf 'Branch content\n' > file.txt
ci -q -f -r1.1.1.1 -mbranch -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# 2. Save initial RCS state (input.txt,v)
cp file.txt,v input.txt,v

# 3. Modify working file (not really needed for rcs command but good practice)
# rcs command operates on RCS file.

# 4. Save working file state (input.txt)
cp file.txt input.txt

# 5. Run the test command
rcs -b1.1.1.1 file.txt

# 6. Generate txtar
cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -b sets default branch

-- options.conf --
{"args": ["-b1.1.1.1", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
