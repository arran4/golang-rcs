#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="1942-rcsclean-r-rev.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# 1. Setup
printf 'Initial content\n' > file.txt
ci -q -i -t-desc -minitial -wtester -d'2020-01-01 00:00:00Z' file.txt </dev/null
# Check out, modify, check in to make 1.2
co -l -q file.txt
printf 'Second content\n' > file.txt
ci -q -msecond -wtester -d'2020-01-02 00:00:00Z' file.txt </dev/null

# Now check out 1.1 (unlocked)
co -r1.1 -q file.txt

# 2. Save initial RCS state (input.txt,v)
cp file.txt,v input.txt,v

# 3. Save working file state (input.txt)
# This content matches 1.1, so rcsclean -r1.1 should remove it.
cp file.txt input.txt

# 4. Run the test command
# -r1.1 should check against 1.1.
# If it matches, file is removed.
# Note: input.txt,v is not modified by rcsclean unless -u is used (or unless it removes locks? No, rcsclean without -u doesn't touch RCS file unless it's unlocking? Wait).
# rcsclean without -u:
# - unlocks if locked? No.
# - removes working file if matches revision.
# So RCS file should be unchanged.
rcsclean -q -r1.1 file.txt

# 5. Generate txtar
cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs clean -r1.1 (check against specific revision)

-- options.conf --
{"args": ["-q", "-r1.1", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs clean

-- expected.txt,v --
$(cat file.txt,v)
EOF
