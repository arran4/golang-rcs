#!/usr/bin/env bash
set -euo pipefail
export TZ=UTC LOGNAME=tester USER=tester
unset RCSINIT

OUT="3051-rcs-u.txtar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

# 1. Setup
printf 'Initial content\n' > file.txt
ci -q -t-desc -minitial -wtester -l -d'2020-01-01 00:00:00Z' file.txt </dev/null
# -l locks the revision.

# 2. Save initial RCS state (input.txt,v)
cp file.txt,v input.txt,v

# 3. Modify working file (not really needed for rcs command but good practice)
# rcs command operates on RCS file.

# 4. Save working file state (input.txt)
cp file.txt input.txt

# 5. Run the test command
rcs -u1.1 file.txt

# 6. Generate txtar
cat > "$OLDPWD/$OUT" <<EOF
-- description.txt --
rcs -u unlocks revision

-- options.conf --
{"args": ["-u1.1", "input.txt"]}

-- input.txt --
$(cat input.txt)

-- input.txt,v --
$(cat input.txt,v)

-- tests.txt --
rcs

-- expected.txt,v --
$(cat file.txt,v)
EOF
